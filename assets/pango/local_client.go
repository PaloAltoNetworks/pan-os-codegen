package pango

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	pangoerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/plugin"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/pango/version"
	"github.com/PaloAltoNetworks/pango/xmlapi"
	"github.com/antchfx/xmlquery"
)

var (
	// ErrUnsupportedOperation is returned when an operation is not supported in local XML mode
	ErrUnsupportedOperation = pangoerrors.ErrUnsupportedOperation

	// ErrWriteNotSupported is returned when attempting write operations in local XML mode
	ErrWriteNotSupported = pangoerrors.ErrWriteNotSupported

	// ErrJobsNotSupported is returned when attempting job operations in local XML mode
	ErrJobsNotSupported = pangoerrors.ErrJobsNotSupported
)

// LocalXmlClient is a client implementation that operates on local XML configuration
// files without connecting to a live PAN-OS device. This is useful for testing,
// validation, and offline analysis of configurations.
//
// The client supports read-only operations (get/show actions). Write operations
// (set/edit/delete), commits, and operational commands will return errors.
//
// Example usage:
//
//	configXml, _ := os.ReadFile("running-config.xml")
//	client, err := pango.NewLocalXmlClient(configXml)
//	if err != nil {
//	    panic(err)
//	}
//
//	svc := address.NewService(client)
//	entry, err := svc.Read(ctx, loc, "web-server", "get")
type LocalXmlClient struct {
	// DOM tree for entire configuration
	rootNode *xmlquery.Node

	// Version information (detected from XML or constructor parameter)
	version version.Number

	// Device type (detected from config structure)
	deviceType deviceType

	// Metadata
	hostname   string
	systemInfo map[string]string

	// Configuration
	strictMode bool // Error on unsupported operations vs silent error

	// Concurrency control
	mu sync.RWMutex // Protects rootNode access
}

// deviceType represents the type of PAN-OS device
type deviceType int

const (
	deviceTypeFirewall deviceType = iota
	deviceTypePanorama
)

// LocalClientOption is a functional option for configuring LocalXmlClient
type LocalClientOption func(*LocalXmlClient) error

// WithVersion sets the PAN-OS version explicitly instead of detecting from XML
func WithVersion(v version.Number) LocalClientOption {
	return func(c *LocalXmlClient) error {
		c.version = v
		return nil
	}
}

// WithHostname sets the hostname for the client
func WithHostname(h string) LocalClientOption {
	return func(c *LocalXmlClient) error {
		c.hostname = h
		return nil
	}
}

// WithStrictMode enables strict mode where unsupported operations return errors
// instead of silent failures
func WithStrictMode(strict bool) LocalClientOption {
	return func(c *LocalXmlClient) error {
		c.strictMode = strict
		return nil
	}
}

// NewLocalXmlClient creates a client that operates on local XML configuration
// without connecting to a live PAN-OS device.
//
// Parameters:
//   - configXml: Raw XML configuration bytes (typically from exported running-config.xml)
//   - opts: Optional configuration (version, hostname, strict mode)
//
// The client will attempt to detect the PAN-OS version from the detail-version
// attribute in the XML. If not present, you can specify it using WithVersion option.
func NewLocalXmlClient(configXml []byte, opts ...LocalClientOption) (*LocalXmlClient, error) {
	// Parse XML into DOM tree
	doc, err := xmlquery.Parse(bytes.NewReader(configXml))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find the <config> root element for version detection
	// xmlquery.Parse returns a document node, we use that as rootNode so XPaths work as-is
	configNode := xmlquery.FindOne(doc, "/config")
	if configNode == nil {
		return nil, fmt.Errorf("expected <config> root element, not found in document")
	}

	client := &LocalXmlClient{
		rootNode:   doc, // Use document root so XPaths starting with /config work directly
		systemInfo: make(map[string]string),
	}

	// Detect version from detail-version attribute
	if versionAttr := configNode.SelectAttr("detail-version"); versionAttr != "" {
		client.version, err = version.New(versionAttr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version from detail-version attribute: %w", err)
		}
	}

	// Apply options (can override detected version)
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	// Detect device type
	if err := client.detectDeviceType(); err != nil {
		return nil, err
	}

	return client, nil
}

// detectDeviceType determines if this is a Panorama or Firewall config
func (c *LocalXmlClient) detectDeviceType() error {
	// Check for Panorama-specific node at /config/panorama
	panoramaNode := xmlquery.FindOne(c.rootNode, "/config/panorama")
	if panoramaNode != nil {
		c.deviceType = deviceTypePanorama
		return nil
	}

	// Default to Firewall
	c.deviceType = deviceTypeFirewall
	return nil
}

// Versioning returns the version number of PAN-OS
func (c *LocalXmlClient) Versioning() version.Number {
	return c.version
}

// Plugins returns an empty list (not supported in local mode)
func (c *LocalXmlClient) Plugins(ctx context.Context) ([]plugin.Info, error) {
	return []plugin.Info{}, nil
}

// GetTarget returns the target param (always empty for local client)
func (c *LocalXmlClient) GetTarget() string {
	return ""
}

// IsPanorama returns true if this is a Panorama configuration
func (c *LocalXmlClient) IsPanorama() (bool, error) {
	return c.deviceType == deviceTypePanorama, nil
}

// IsFirewall returns true if this is a Firewall configuration
func (c *LocalXmlClient) IsFirewall() (bool, error) {
	return c.deviceType == deviceTypeFirewall, nil
}

// Clock returns an error (not supported in local mode)
func (c *LocalXmlClient) Clock(ctx context.Context) (time.Time, error) {
	return time.Time{}, ErrUnsupportedOperation
}

// ReadFromConfig returns the XML at the given XPATH location.
// This provides compatibility with the existing local inspection mode.
func (c *LocalXmlClient) ReadFromConfig(ctx context.Context, path []string, withPackaging bool, ans any) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path is empty")
	}

	// Convert path to XPath string
	xpath := "/" + util.AsXpath(path)

	// Execute query
	nodes := xmlquery.Find(c.rootNode, xpath)
	if len(nodes) == 0 {
		return nil, pangoerrors.ObjectNotFound()
	}

	// Format response
	var buf bytes.Buffer
	if withPackaging {
		buf.WriteString("<q>")
	}

	for _, node := range nodes {
		buf.WriteString(node.OutputXML(false))
	}

	if withPackaging {
		buf.WriteString("</q>")
	}

	result := buf.Bytes()

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(result, ans); err != nil {
			return result, err
		}
	}

	return result, nil
}

// Communicate sends the given command to the local XML tree.
// Supports read operations (get/show) and write operations (set/edit).
func (c *LocalXmlClient) Communicate(ctx context.Context, cmd util.PangoCommand, strip bool, ans any) ([]byte, *http.Response, error) {
	// Type assert to xmlapi.Config
	config, ok := cmd.(*xmlapi.Config)
	if !ok {
		return nil, nil, ErrUnsupportedOperation
	}

	// Check context before attempting lock
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	// Acquire appropriate lock based on action
	if config.Action == "get" || config.Action == "show" {
		c.mu.RLock()
		defer c.mu.RUnlock()
	} else {
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	// Check context again after acquiring lock
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Handle operation based on action
	switch config.Action {
	case "get", "show":
		return c.handleRead(config, strip, ans)

	case "set":
		return c.handleSet(config, strip, ans)

	case "edit":
		return c.handleEdit(config, strip, ans)

	case "delete":
		return c.handleDelete(config, strip, ans)

	case "rename":
		return c.handleRename(config, strip, ans)

	case "move":
		return c.handleMove(config, strip, ans)

	default:
		return nil, nil, ErrWriteNotSupported
	}
}

// elementToString converts an element value to XML string.
// Element can be:
// - string: returned as-is
// - io.Stringer: calls String() method
// - struct: marshals to XML using xml.Marshal()
// - []byte: converts to string
func elementToString(element any) (string, error) {
	if element == nil {
		return "", fmt.Errorf("element is nil")
	}

	// Already a string
	if s, ok := element.(string); ok {
		return s, nil
	}

	// Implements io.Stringer
	if s, ok := element.(fmt.Stringer); ok {
		return s.String(), nil
	}

	// Byte slice
	if b, ok := element.([]byte); ok {
		return string(b), nil
	}

	// Marshal struct to XML
	xmlBytes, err := xml.Marshal(element)
	if err != nil {
		return "", fmt.Errorf("failed to marshal element to XML: %w", err)
	}

	return string(xmlBytes), nil
}

// copyNode creates a deep copy of an xmlquery.Node.
// This is needed when moving nodes from a parsed document to the main tree.
func copyNode(node *xmlquery.Node) *xmlquery.Node {
	if node == nil {
		return nil
	}

	// Create new node with same type and data
	copy := &xmlquery.Node{
		Type: node.Type,
		Data: node.Data,
	}

	// Copy attributes
	for _, attr := range node.Attr {
		copy.Attr = append(copy.Attr, attr)
	}

	// Recursively copy children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childCopy := copyNode(child)
		xmlquery.AddChild(copy, childCopy)
	}

	return copy
}

// handleRead executes read operations (get/show)
func (c *LocalXmlClient) handleRead(config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	// Execute XPath query directly - rootNode is document root, so XPaths work as-is
	xpath := config.Xpath
	nodes := xmlquery.Find(c.rootNode, xpath)
	if len(nodes) == 0 {
		// Return empty but valid XML - matches real API behavior
		// Callers can still unmarshal to get empty slice while checking IsObjectNotFound
		emptyXml := c.formatResponse([]*xmlquery.Node{}, strip, false)
		return emptyXml, c.mockHttpResponse(404), pangoerrors.ObjectNotFound()
	}

	// Format response
	responseXml := c.formatResponse(nodes, strip)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleSet executes SET operation (create or overwrite)
func (c *LocalXmlClient) handleSet(config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Convert element to string
	elementStr, err := elementToString(config.Element)
	if err != nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to convert element to string: %w", err))
	}

	// Parse element from config.Element
	element, err := xmlquery.Parse(strings.NewReader(elementStr))
	if err != nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to parse element: %w", err))
	}

	// Ensure parent path exists, creating all intermediate elements if needed
	// For SET, we create the full path including the final container element
	parent, err := ensureXpathExists(c.rootNode, config.Xpath, false)
	if err != nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to ensure XPath exists: %w", err))
	}

	// Get the element to insert (skip over document/text nodes to find first element)
	var parsedElement *xmlquery.Node
	for child := element.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.ElementNode {
			parsedElement = child
			break
		}
	}
	if parsedElement == nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("element is empty"))
	}

	// Remove existing child with same name if present (overwrite semantics)
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.ElementNode && child.Data == parsedElement.Data {
			// Check if it's the same entry by name attribute
			nameAttr := child.SelectAttr("name")
			newNameAttr := parsedElement.SelectAttr("name")
			if nameAttr == newNameAttr {
				xmlquery.RemoveFromTree(child)
				break
			}
		}
	}

	// Copy the parsed element to create a new node
	// Cannot add parsed node directly as it belongs to a different document tree
	newElement := copyNode(parsedElement)

	// Add new element
	xmlquery.AddChild(parent, newElement)

	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{newElement}, strip)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleEdit executes EDIT operation (merge fields)
func (c *LocalXmlClient) handleEdit(config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Ensure target path exists, creating intermediate container elements if needed
	// For EDIT, the final segment (the entry being edited) must already exist
	target, err := ensureXpathExists(c.rootNode, config.Xpath, true)
	if err != nil {
		// If error is ObjectNotFound, return as-is (final segment doesn't exist)
		// Otherwise wrap in InvalidXpath
		var notFound *pangoerrors.ErrObjectNotFound
		if errors.As(err, &notFound) {
			return nil, nil, err
		}
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to ensure XPath exists: %w", err))
	}

	// Convert element to string
	elementStr, err := elementToString(config.Element)
	if err != nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to convert element to string: %w", err))
	}

	// Parse new content - wrap in a root element to parse multiple fields
	wrappedContent := "<root>" + elementStr + "</root>"
	newContent, err := xmlquery.Parse(strings.NewReader(wrappedContent))
	if err != nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to parse element: %w", err))
	}

	// Find the root element (skip over document/text nodes)
	var rootElement *xmlquery.Node
	for child := newContent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.ElementNode {
			rootElement = child
			break
		}
	}
	if rootElement == nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to find root element in parsed content"))
	}

	// Merge: update existing fields, add new fields
	// rootElement is the <root> element, its children are the fields to merge
	for newChild := rootElement.FirstChild; newChild != nil; newChild = newChild.NextSibling {
		if newChild.Type != xmlquery.ElementNode {
			continue
		}

		// Find matching child in target
		found := false
		for targetChild := target.FirstChild; targetChild != nil; targetChild = targetChild.NextSibling {
			if targetChild.Type == xmlquery.ElementNode && targetChild.Data == newChild.Data {
				// Replace content - remove all children
				for child := targetChild.FirstChild; child != nil; {
					next := child.NextSibling
					xmlquery.RemoveFromTree(child)
					child = next
				}

				// Copy content from new child
				if newChild.FirstChild != nil {
					for child := newChild.FirstChild; child != nil; child = child.NextSibling {
						copyChild := copyNode(child)
						xmlquery.AddChild(targetChild, copyChild)
					}
				}

				found = true
				break
			}
		}

		if !found {
			// Add new field - need to copy the node
			nodeCopy := copyNode(newChild)
			xmlquery.AddChild(target, nodeCopy)
		}
	}

	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{target}, strip)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleDelete executes DELETE operation (remove element)
func (c *LocalXmlClient) handleDelete(config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Find target element
	target := xmlquery.FindOne(c.rootNode, config.Xpath)
	if target == nil {
		return nil, nil, pangoerrors.NewErrObjectNotFound(config.Xpath)
	}

	// Remove from tree
	xmlquery.RemoveFromTree(target)

	// Format empty response (element was deleted)
	responseXml := c.formatResponse([]*xmlquery.Node{}, strip)

	// Unmarshal if ans struct provided (though response is empty)
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleRename executes RENAME operation (change @name attribute)
func (c *LocalXmlClient) handleRename(config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Find target element
	target := xmlquery.FindOne(c.rootNode, config.Xpath)
	if target == nil {
		return nil, nil, pangoerrors.NewErrObjectNotFound(config.Xpath)
	}

	// Get new name
	newName := config.NewName
	if newName == "" {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("new name is required for rename operation"))
	}

	// Check for conflict: does new name already exist?
	parent := target.Parent
	if parent != nil {
		for sibling := parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
			if sibling.Type == xmlquery.ElementNode && sibling != target {
				siblingName := sibling.SelectAttr("name")
				if siblingName == newName {
					sourceName := target.SelectAttr("name")
					return nil, nil, pangoerrors.NewErrRenameConflict(config.Xpath, sourceName, newName)
				}
			}
		}
	}

	// Update name attribute
	found := false
	for i, attr := range target.Attr {
		if attr.Name.Local == "name" {
			target.Attr[i].Value = newName
			found = true
			break
		}
	}
	if !found {
		// Add name attribute if it doesn't exist
		target.Attr = append(target.Attr, xmlquery.Attr{
			Name:  xml.Name{Local: "name"},
			Value: newName,
		})
	}

	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{target}, strip)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleMove executes MOVE operation (reorder element within parent)
func (c *LocalXmlClient) handleMove(config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Find target element
	target := xmlquery.FindOne(c.rootNode, config.Xpath)
	if target == nil {
		return nil, nil, pangoerrors.NewErrObjectNotFound(config.Xpath)
	}

	// Get where parameter
	where := config.Where
	if where == "" {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("where parameter is required for move operation"))
	}

	// Validate where value
	validWhere := map[string]bool{"before": true, "after": true, "top": true, "bottom": true}
	if !validWhere[where] {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("invalid where value: %s (must be before, after, top, or bottom)", where))
	}

	parent := target.Parent
	if parent == nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("cannot move root element"))
	}

	// Remove from current position
	xmlquery.RemoveFromTree(target)

	// Reinsert at new position
	switch where {
	case "top":
		// Insert at beginning - need to find first child and insert before it
		if parent.FirstChild != nil {
			// Insert before first child
			parent.FirstChild.Parent = nil
			oldFirst := parent.FirstChild
			parent.FirstChild = target
			target.Parent = parent
			target.NextSibling = oldFirst
			if oldFirst != nil {
				oldFirst.PrevSibling = target
				oldFirst.Parent = parent
			}
		} else {
			// No children, just add
			xmlquery.AddChild(parent, target)
		}

	case "bottom":
		// Append to end
		xmlquery.AddChild(parent, target)

	case "before", "after":
		// Need destination element name
		dst := config.Destination
		if dst == "" {
			return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
				fmt.Errorf("destination is required for before/after move"))
		}

		// Find destination sibling
		var dstNode *xmlquery.Node
		for sibling := parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
			if sibling.Type == xmlquery.ElementNode {
				if sibling.SelectAttr("name") == dst {
					dstNode = sibling
					break
				}
			}
		}

		if dstNode == nil {
			return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
				fmt.Errorf("destination element not found: %s", dst))
		}

		if where == "before" {
			// Insert before destination
			target.Parent = parent
			target.PrevSibling = dstNode.PrevSibling
			target.NextSibling = dstNode

			if dstNode.PrevSibling != nil {
				dstNode.PrevSibling.NextSibling = target
			} else {
				// dstNode was first child
				parent.FirstChild = target
			}
			dstNode.PrevSibling = target
		} else { // after
			// Insert after destination
			target.Parent = parent
			target.PrevSibling = dstNode
			target.NextSibling = dstNode.NextSibling

			if dstNode.NextSibling != nil {
				dstNode.NextSibling.PrevSibling = target
			} else {
				// dstNode was last child
				parent.LastChild = target
			}
			dstNode.NextSibling = target
		}
	}

	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{target}, strip)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// formatResponse constructs a PAN-OS API-compatible XML response
func (c *LocalXmlClient) formatResponse(nodes []*xmlquery.Node, strip bool) []byte {
	var buf bytes.Buffer

	// When strip=false, include full response wrapper
	// When strip=true, only include result wrapper (normalizer expects it)
	if !strip {
		buf.WriteString(`<response status="success">`)
	}

	// Always include result wrapper with count attributes (normalizer needs this)
	buf.WriteString(fmt.Sprintf(
		`<result total-count="%d" count="%d">`,
		len(nodes), len(nodes)))

	// Serialize each node (true includes the node itself with attributes)
	for _, node := range nodes {
		buf.WriteString(node.OutputXML(true))
	}

	buf.WriteString(`</result>`)

	if !strip {
		buf.WriteString(`</response>`)
	}

	return buf.Bytes()
}

// mockHttpResponse creates a mock HTTP response for compatibility
func (c *LocalXmlClient) mockHttpResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
}

// StartJob returns an error (not supported in local mode)
func (c *LocalXmlClient) StartJob(ctx context.Context, cmd util.PangoCommand) (uint, []byte, *http.Response, error) {
	return 0, nil, nil, ErrJobsNotSupported
}

// WaitForJob returns an error (not supported in local mode)
func (c *LocalXmlClient) WaitForJob(ctx context.Context, id uint, sleep time.Duration, resp any) error {
	return ErrJobsNotSupported
}

// WaitForLogs returns an error (not supported in local mode)
func (c *LocalXmlClient) WaitForLogs(ctx context.Context, id uint, sleep time.Duration, resp any) ([]byte, error) {
	return nil, ErrJobsNotSupported
}

// MultiConfig returns an error (not supported in local mode)
// MultiConfig executes a batch of configuration commands as an atomic transaction.
// All operations must succeed for any changes to be applied. If any operation fails,
// the entire batch is rolled back.
//
// Success: All operations succeed → working copy committed to backing document
// Failure: Any operation fails → working copy discarded, backing document unchanged
//
// The error returned on failure is ErrOperationFailed which includes the index
// of the failed operation and the underlying cause.
//
// IMPORTANT: This method acquires a write lock for the entire transaction.
// All operations are executed sequentially in order.
//
// Example usage:
//
//	mc := &xmlapi.MultiConfig{
//	    Operations: []xmlapi.MultiConfigOperation{
//	        {Action: "set", XPath: "//address", Element: "<entry name='web1'/>"},
//	        {Action: "set", XPath: "//address", Element: "<entry name='web2'/>"},
//	    },
//	}
//	_, _, _, err := client.MultiConfig(ctx, mc, false, nil)
func (c *LocalXmlClient) MultiConfig(ctx context.Context, mc *xmlapi.MultiConfig, strict bool, extras url.Values) ([]byte, *http.Response, *xmlapi.MultiConfigResponse, error) {
	// Validate input
	if mc == nil || len(mc.Operations) == 0 {
		// Return empty successful response for empty batch
		mcResp := &xmlapi.MultiConfigResponse{
			Status:  "success",
			Code:    20,
			Results: []xmlapi.MultiConfigResponseElement{},
		}
		return []byte("<response status=\"success\"></response>"), c.mockHttpResponse(200), mcResp, nil
	}

	// Check context before lock
	select {
	case <-ctx.Done():
		return nil, nil, nil, ctx.Err()
	default:
	}

	// Acquire write lock (exclusive)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check context after lock
	if ctx.Err() != nil {
		return nil, nil, nil, ctx.Err()
	}

	// Create working copy for atomic transaction
	workingCopy := c.cloneDocument()

	// Execute operations sequentially on working copy
	for i, op := range mc.Operations {
		// Check context before each operation
		if ctx.Err() != nil {
			return nil, nil, nil, ctx.Err()
		}

		// Convert MultiConfigOperation to Config
		// Action is stored in XMLName.Local
		config := &xmlapi.Config{
			Action:      op.XMLName.Local,
			Xpath:       op.Xpath,
			Element:     op.Data,
			NewName:     op.NewName,
			Where:       op.Where,
			Destination: op.Destination,
		}

		// Execute operation on working copy
		err := c.executeOperationOnDoc(workingCopy, config)
		if err != nil {
			// Operation failed - discard working copy and return error with index
			return nil, nil, nil, pangoerrors.NewErrOperationFailed(i, err)
		}
	}

	// All operations succeeded - commit working copy to backing document
	c.commitWorkingCopy(workingCopy)

	// Format response (empty result for multiconfig)
	responseXml := c.formatResponse([]*xmlquery.Node{}, false)

	// Create MultiConfigResponse
	mcResp := &xmlapi.MultiConfigResponse{
		Status:  "success",
		Code:    20,
		Results: []xmlapi.MultiConfigResponseElement{},
	}

	return responseXml, c.mockHttpResponse(200), mcResp, nil
}

// ImportFile returns an error (not supported in local mode)
func (c *LocalXmlClient) ImportFile(ctx context.Context, cmd *xmlapi.Import, content []byte, filename, fp string, strip bool, ans any) ([]byte, *http.Response, error) {
	return nil, nil, ErrUnsupportedOperation
}

// ExportFile returns an error (not supported in local mode)
func (c *LocalXmlClient) ExportFile(ctx context.Context, cmd *xmlapi.Export, ans any) (string, []byte, *http.Response, error) {
	return "", nil, nil, ErrUnsupportedOperation
}

// GenerateApiKey returns an error (not supported in local mode)
func (c *LocalXmlClient) GenerateApiKey(ctx context.Context, username, password string) (string, error) {
	return "", ErrUnsupportedOperation
}

// RequestPasswordHash returns an error (not supported in local mode)
func (c *LocalXmlClient) RequestPasswordHash(ctx context.Context, v string) (string, error) {
	return "", ErrUnsupportedOperation
}

// GetTechSupportFile returns an error (not supported in local mode)
func (c *LocalXmlClient) GetTechSupportFile(ctx context.Context) (string, []byte, error) {
	return "", nil, ErrUnsupportedOperation
}

// cloneDocument creates a deep copy of the XML document tree.
// The returned copy is completely independent - modifications to the clone
// do not affect the original rootNode.
//
// This is used by MultiConfig to create working copies for atomic transactions:
// operations execute on the clone, and on success the clone replaces the backing document.
//
// Implementation uses serialize/deserialize to guarantee a complete deep copy
// of all nodes, attributes, and text content.
func (c *LocalXmlClient) cloneDocument() *xmlquery.Node {
	// Serialize current document to bytes
	var buf bytes.Buffer
	buf.WriteString(c.rootNode.OutputXML(true))

	// Parse back into new independent tree
	doc, err := xmlquery.Parse(&buf)
	if err != nil {
		// Should never happen with valid XML from our own tree
		panic(fmt.Sprintf("failed to clone document: %v", err))
	}

	return doc
}

// commitWorkingCopy atomically replaces the backing XML document with a working copy.
// The old rootNode becomes unreferenced and will be garbage collected.
//
// CRITICAL: Caller MUST hold write lock (c.mu.Lock) before calling this function.
// Calling without lock will cause race conditions.
//
// Correct usage:
//
//	c.mu.Lock()
//	defer c.mu.Unlock()
//	workingCopy := c.cloneDocument()
//	// ... execute operations on workingCopy
//	c.commitWorkingCopy(workingCopy)
func (c *LocalXmlClient) commitWorkingCopy(workingCopy *xmlquery.Node) {
	c.rootNode = workingCopy
	// Old rootNode is now unreferenced, GC will collect it
}

// validateXpath checks XPath syntax without modifying the document.
// Returns nil if the XPath is valid, or ErrInvalidXpath if malformed.
//
// This validation prevents XML corruption from malformed XPath expressions
// by catching syntax errors before execution.
//
// Example usage:
//
//	if err := c.validateXpath(xpath); err != nil {
//	    return err  // ErrInvalidXpath with underlying cause
//	}
func (c *LocalXmlClient) validateXpath(xpath string) error {
	// Empty XPath is invalid
	if xpath == "" {
		return pangoerrors.NewErrInvalidXpath(xpath, fmt.Errorf("XPath cannot be empty"))
	}

	// Test XPath by executing a safe query (doesn't modify document)
	// If query fails, XPath syntax is invalid
	_, err := xmlquery.QueryAll(c.rootNode, xpath)
	if err != nil {
		return pangoerrors.NewErrInvalidXpath(xpath, err)
	}

	return nil
}

// executeOperationOnDoc executes a single operation on the specified document tree.
// This is used by MultiConfig to execute operations on working copies.
//
// Unlike the handle* methods which operate on c.rootNode, this operates on an
// arbitrary document tree passed as a parameter. This enables atomic transactions.
//
// Supports all write operations: set, edit, delete, rename, move.
// Read operations (get/show) are not supported in MultiConfig context.
func (c *LocalXmlClient) executeOperationOnDoc(doc *xmlquery.Node, config *xmlapi.Config) error {
	// Create temporary client with working copy to reuse existing handlers
	// This avoids code duplication while maintaining proper encapsulation
	tempClient := &LocalXmlClient{
		rootNode:   doc,
		version:    c.version,
		deviceType: c.deviceType,
		logger:     c.logger,
	}

	// Validate operation is supported in MultiConfig
	switch config.Action {
	case "get", "show":
		// Read operations not allowed in MultiConfig
		return pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("read operations (get/show) not supported in MultiConfig"))

	case "set":
		_, _, err := tempClient.handleSet(config, false, nil)
		return err

	case "edit":
		_, _, err := tempClient.handleEdit(config, false, nil)
		return err

	case "delete":
		_, _, err := tempClient.handleDelete(config, false, nil)
		return err

	case "rename":
		_, _, err := tempClient.handleRename(config, false, nil)
		return err

	case "move":
		_, _, err := tempClient.handleMove(config, false, nil)
		return err

	default:
		return ErrWriteNotSupported
	}
}

// parseXpathSegments splits an XPath into individual segments, handling predicates correctly.
// Predicates (expressions in square brackets) are kept as part of their segment.
//
// Example:
//
//	Input:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='test-dg']/address"
//	Output: ["config", "devices", "entry[@name='localhost.localdomain']", "device-group", "entry[@name='test-dg']", "address"]
//
// Returns error if XPath is empty or malformed.
func parseXpathSegments(xpath string) ([]string, error) {
	// Remove leading "/" if present
	xpath = strings.TrimPrefix(xpath, "/")

	if xpath == "" {
		return nil, fmt.Errorf("empty XPath after parsing")
	}

	// Split by "/" but respect predicate brackets
	var segments []string
	var current strings.Builder
	inPredicate := false

	for _, char := range xpath {
		switch char {
		case '[':
			inPredicate = true
			current.WriteRune(char)
		case ']':
			inPredicate = false
			current.WriteRune(char)
		case '/':
			if inPredicate {
				current.WriteRune(char)
			} else {
				if current.Len() > 0 {
					segments = append(segments, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add final segment
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("empty XPath after parsing")
	}

	return segments, nil
}

// extractPredicateAttributes parses an XPath segment and extracts the element name
// and any attributes defined in predicates.
//
// Example:
//
//	Input:  "entry[@name='test-dg']"
//	Output: elementName="entry", attrs=[{Name.Local: "name", Value: "test-dg"}]
//
//	Input:  "address"
//	Output: elementName="address", attrs=[]
//
//	Input:  "entry[@name='dg'][@uuid='123']"
//	Output: elementName="entry", attrs=[{Name.Local: "name", Value: "dg"}, {Name.Local: "uuid", Value: "123"}]
//
// Returns error if predicates are malformed.
func extractPredicateAttributes(segment string) (string, []xmlquery.Attr, error) {
	// Check if segment has predicates
	bracketIdx := strings.Index(segment, "[")
	if bracketIdx == -1 {
		// No predicates - simple element name
		return segment, nil, nil
	}

	elementName := segment[:bracketIdx]
	predicateStr := segment[bracketIdx:]

	// Parse predicates using regex
	// Pattern: [@attrName='value'] or [@attrName="value"]
	re := regexp.MustCompile(`@(\w+)=['"]([^'"]+)['"]`)
	matches := re.FindAllStringSubmatch(predicateStr, -1)

	var attrs []xmlquery.Attr
	for _, match := range matches {
		if len(match) >= 3 {
			attrs = append(attrs, xmlquery.Attr{
				Name:  xml.Name{Local: match[1]},
				Value: match[2],
			})
		}
	}

	if len(attrs) == 0 && bracketIdx != -1 {
		return "", nil, fmt.Errorf("predicates found but no attributes extracted from: %s", segment)
	}

	return elementName, attrs, nil
}

// ensureXpathExists ensures all intermediate elements in an XPath exist, creating them if needed.
// Returns the final node (for SET operations) or the parent node (for EDIT operations targeting entries).
//
// This function walks the XPath segment by segment, creating missing intermediate container elements
// as needed. It does NOT auto-create entry elements (elements named "entry" with a name attribute),
// as these represent actual configuration objects that should be explicitly created via SET operations.
//
// For SET operations (adding to containers):
//
//	xpath := "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='test-dg']/address"
//	parent, err := ensureXpathExists(rootNode, xpath, false)
//	// parent will be the <address> node, creating it and any missing ancestors if needed
//
// For EDIT operations (modifying existing entries):
//
//	xpath := "/config/devices/entry[@name='localhost.localdomain']/address/entry[@name='addr1']"
//	entry, err := ensureXpathExists(rootNode, xpath, true)
//	// Will create /config/devices/entry[@name='localhost.localdomain']/address if missing
//	// But returns error if entry[@name='addr1'] doesn't exist (requireFinalSegment=true)
//
// Parameters:
//   - rootNode: The root of the XML tree
//   - xpath: The XPath to ensure exists
//   - requireFinalSegment: If true, the final segment must already exist (for EDIT). If false, create it (for SET).
//
// Returns error if XPath is invalid or cannot be parsed, or if requireFinalSegment=true and final segment doesn't exist.
func ensureXpathExists(rootNode *xmlquery.Node, xpath string, requireFinalSegment bool) (*xmlquery.Node, error) {
	// Parse XPath into segments
	segments, err := parseXpathSegments(xpath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XPath: %w", err)
	}

	// Start from root node
	currentNode := rootNode
	currentPath := ""

	// Walk through each segment
	for i, segment := range segments {
		// Build XPath up to current segment
		if currentPath == "" {
			currentPath = "/" + segment
		} else {
			currentPath = currentPath + "/" + segment
		}

		// Try to find this segment
		nextNode := xmlquery.FindOne(currentNode, segment)

		if nextNode == nil {
			// Check if this is the final segment and we require it to exist
			isFinalSegment := (i == len(segments)-1)
			if isFinalSegment && requireFinalSegment {
				// Final segment doesn't exist and is required - return error
				return nil, pangoerrors.NewErrObjectNotFound(xpath)
			}

			// Segment doesn't exist - create it
			elementName, attrs, err := extractPredicateAttributes(segment)
			if err != nil {
				return nil, fmt.Errorf("failed to parse segment '%s': %w", segment, err)
			}

			// Create new element node
			newNode := &xmlquery.Node{
				Type: xmlquery.ElementNode,
				Data: elementName,
				Attr: attrs,
			}

			// Add to current node
			xmlquery.AddChild(currentNode, newNode)
			currentNode = newNode
		} else {
			currentNode = nextNode
		}
	}

	return currentNode, nil
}
