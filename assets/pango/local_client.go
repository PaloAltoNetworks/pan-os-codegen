package pango

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
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

// Compile-time check that LocalXmlClient implements PangoClient interface
var _ util.PangoClient = (*LocalXmlClient)(nil)

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
// The client supports full CRUD operations (create, read, update, delete) with
// optional auto-save functionality. File loading can be deferred using the Setup()
// method pattern.
//
// Key Features:
//   - Deferred file loading via Setup()
//   - Explicit file I/O control (LoadFromFile, SaveToFile)
//   - Optional auto-save mode (saves after each operation)
//   - MultiConfig transaction support (atomic batched operations)
//   - Thread-safe operations with RWMutex
//
// Example usage:
//
//	client, err := pango.NewLocalXmlClient("/path/to/running-config.xml")
//	if err != nil {
//	    panic(err)
//	}
//
//	if err := client.Setup(); err != nil {
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
	strictMode       bool // Error on unsupported operations vs silent error
	checkEnvironment bool // Whether we should check PANOS_ environment variables

	// File I/O and auto-save fields
	filepath string // Current file path for auto-save and deferred loading
	autoSave bool   // Auto-save mode enabled

	// Concurrency control
	mu sync.RWMutex // Protects rootNode access

	// Logger
	logger *categoryLogger
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

// WithLoggerInfo configures LocalXmlClient logging
func WithLoggingInfo(loggingInfo LoggingInfo) LocalClientOption {
	return func(c *LocalXmlClient) error {
		logger, err := SetupLogger(loggingInfo, c.checkEnvironment)
		if err != nil {
			return err
		}
		c.logger = logger
		return nil
	}
}

func WithCheckEnvironment(checkEnvironment bool) LocalClientOption {
	return func(c *LocalXmlClient) error {
		c.checkEnvironment = checkEnvironment
		return nil
	}
}

// WithAutoSave enables or disables auto-save mode.
// When enabled, the client automatically saves changes to the file
// after each CRUD operation (set/edit/delete).
//
// During multiconfig operations (between Op and Commit), auto-save
// is deferred until all operations complete successfully.
//
// Default: false (disabled)
//
// Example:
//
//	client, err := pango.NewLocalXmlClient(
//	    "/path/to/config.xml",
//	    pango.WithAutoSave(true),
//	)
func WithAutoSave(enabled bool) LocalClientOption {
	return func(c *LocalXmlClient) error {
		c.autoSave = enabled
		return nil
	}
}

// NewLocalXmlClient creates a client that operates on local XML configuration
// without connecting to a live PAN-OS device.
//
// The filepath is stored but the file is NOT loaded until Setup() is called.
// This enables lazy initialization and deferred loading patterns.
//
// Parameters:
//   - filepath: Path to XML configuration file (absolute or relative)
//   - opts: Optional configuration (version, hostname, strict mode, auto-save)
//
// Example usage with deferred loading:
//
//	client, err := pango.NewLocalXmlClient("/path/to/config.xml")
//	if err != nil {
//	    return err
//	}
//
//	// Perform other initialization...
//
//	// Load file when ready
//	if err := client.Setup(); err != nil {
//	    return err
//	}
//
//	// Client ready for operations
//	svc := address.NewService(client)
//
// Example with auto-save:
//
//	client, err := pango.NewLocalXmlClient(
//	    "/path/to/config.xml",
//	    pango.WithAutoSave(true),
//	)
//	if err != nil {
//	    return err
//	}
//
//	if err := client.Setup(); err != nil {
//	    return err
//	}
//
//	// Changes automatically saved to file
//	svc.Create(ctx, entry)  // Auto-saved!
func NewLocalXmlClient(filepath string, opts ...LocalClientOption) (*LocalXmlClient, error) {
	// Validate filepath is not empty
	if filepath == "" {
		return nil, fmt.Errorf("filepath cannot be empty")
	}

	// Initialize client
	client := &LocalXmlClient{
		filepath:   filepath, // Store filepath for later loading
		autoSave:   false,    // Auto-save off by default
		systemInfo: make(map[string]string),
	}

	// Apply options (can override defaults like autoSave)
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// LoadFromFile loads XML configuration from the specified file.
// This replaces any existing in-memory configuration.
//
// The filepath is stored and will be used for auto-save operations if enabled.
// Both absolute and relative paths are supported.
//
// Example:
//
//	err := client.LoadFromFile("/path/to/running-config.xml")
//	if err != nil {
//	    return fmt.Errorf("failed to load config: %w", err)
//	}
func (c *LocalXmlClient) LoadFromFile(filepath string) error {
	// Read file
	xmlBytes, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("XML file not found: %s", filepath)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied reading file: %s", filepath)
		}
		return fmt.Errorf("failed to read XML file %s: %w", filepath, err)
	}

	// Parse XML
	doc, err := xmlquery.Parse(bytes.NewReader(xmlBytes))
	if err != nil {
		return fmt.Errorf("failed to parse XML from %s: %w", filepath, err)
	}

	// Find config root
	configNode := xmlquery.FindOne(doc, "/config")
	if configNode == nil {
		return fmt.Errorf("invalid PAN-OS XML: missing <config> root element in %s", filepath)
	}

	// Detect version
	var ver version.Number
	if versionAttr := configNode.SelectAttr("detail-version"); versionAttr != "" {
		ver, err = version.New(versionAttr)
		if err != nil {
			return fmt.Errorf("failed to parse version from %s: %w", filepath, err)
		}
	}

	// Detect device type
	var devType deviceType
	panoramaNode := xmlquery.FindOne(doc, "/config/panorama")
	if panoramaNode != nil {
		devType = deviceTypePanorama
	} else {
		devType = deviceTypeFirewall
	}

	// Update state (thread-safe)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.rootNode = doc
	c.version = ver
	c.deviceType = devType
	c.filepath = filepath

	return nil
}

// saveToFileInternal serializes and saves the XML document without acquiring locks.
// CRITICAL: Caller MUST hold write lock before calling this method.
// This is an internal helper for auto-save integration within locked operations.
//
// Uses atomic write pattern (temp file + rename) to prevent corruption.
//
// Internal usage only - external callers should use SaveToFile() instead.
func (c *LocalXmlClient) saveToFileInternal(filepath string) error {
	if filepath == "" {
		return fmt.Errorf("filepath cannot be empty")
	}

	if c.rootNode == nil {
		return fmt.Errorf("cannot save: client not initialized")
	}

	// Serialize rootNode to XML string
	// Note: c.rootNode access is safe because caller holds lock
	xmlStr := c.rootNode.OutputXML(false)

	// Atomic write pattern: write to temp file, then rename
	dir := path.Dir(filepath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*.xml")
	if err != nil {
		return fmt.Errorf("failed to create temp file in %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write XML content
	if _, err := tmpFile.WriteString(xmlStr); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Fsync before rename (ensure data on disk)
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil // Prevent defer cleanup

	// Atomic rename
	if err := os.Rename(tmpPath, filepath); err != nil {
		os.Remove(tmpPath) // Clean up
		return fmt.Errorf("failed to rename temp file to %s: %w", filepath, err)
	}

	return nil
}

// SaveToFile saves the current in-memory configuration to the specified file.
// Uses atomic write pattern (temp file + rename) to prevent corruption.
//
// The filepath is stored and will be used for subsequent auto-save operations.
//
// Example:
//
//	err := client.SaveToFile("/path/to/output-config.xml")
//	if err != nil {
//	    return fmt.Errorf("failed to save config: %w", err)
//	}
func (c *LocalXmlClient) SaveToFile(filepath string) error {
	// Acquire read lock for serialization
	c.mu.RLock()
	err := c.saveToFileInternal(filepath)
	c.mu.RUnlock()

	if err != nil {
		return err
	}

	// Update filepath (requires write lock)
	c.mu.Lock()
	c.filepath = filepath
	c.mu.Unlock()

	return nil
}

// Setup loads the XML configuration file specified in the constructor.
// This enables deferred loading patterns where the client is created
// but file I/O is delayed until Setup() is called.
//
// This method must be called before performing any CRUD operations.
// Calling Setup() multiple times will reload the file from disk.
//
// Example:
//
//	client, err := pango.NewLocalXmlClient("/path/to/config.xml")
//	if err != nil {
//	    return err
//	}
//
//	// Perform other initialization...
//
//	if err := client.Setup(); err != nil {
//	    return fmt.Errorf("failed to load config: %w", err)
//	}
//
//	// Client ready for operations
//	svc := address.NewService(client)
func (c *LocalXmlClient) Setup() error {
	// Get filepath (thread-safe)
	c.mu.RLock()
	filepath := c.filepath
	c.mu.RUnlock()

	// Validate filepath is set
	if filepath == "" {
		return fmt.Errorf("cannot setup: filepath not set (should have been set in constructor)")
	}

	// Load file using existing LoadFromFile method
	return c.LoadFromFile(filepath)
}

// Initialize is a no-op for LocalXmlClient.
// All initialization work is performed in Setup() which loads the XML file.
// This method exists for PangoClient interface compatibility with API client.
//
// For API mode clients, Initialize performs post-Setup authentication and
// system info retrieval. For LocalXmlClient, this is not applicable.
func (c *LocalXmlClient) Initialize(ctx context.Context) error {
	// No-op: all initialization done in Setup()
	return nil
}

// SetupLocalInspection returns an unsupported operation error.
// This method is specific to API client's local inspection mode and is not
// applicable to LocalXmlClient which already operates on local XML files.
//
// To use LocalXmlClient, use the normal constructor pattern:
//
//	client, err := pango.NewLocalXmlClient("/path/to/config.xml")
//	if err != nil {
//	    return err
//	}
//	if err := client.Setup(); err != nil {
//	    return err
//	}
func (c *LocalXmlClient) SetupLocalInspection(schema, panosVersion string) error {
	return ErrUnsupportedOperation
}

// SetAutoSave enables or disables auto-save mode at runtime.
// When enabled, the client automatically saves changes to the file
// after each CRUD operation (set/edit/delete).
//
// During multiconfig operations, auto-save is deferred until all
// operations complete successfully.
//
// This can be called at any time to toggle auto-save behavior.
//
// Example:
//
//	client.SetAutoSave(true)   // Enable auto-save
//	client.SetAutoSave(false)  // Disable auto-save
func (c *LocalXmlClient) SetAutoSave(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.autoSave = enabled
}

// GetAutoSave returns whether auto-save mode is currently enabled.
//
// Example:
//
//	if client.GetAutoSave() {
//	    fmt.Println("Auto-save is enabled")
//	}
func (c *LocalXmlClient) GetAutoSave() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.autoSave
}

// GetFilepath returns the current file path for this client.
// This is the path that will be used for Setup() to load the configuration,
// and for auto-save operations (if enabled).
// The path may be updated by LoadFromFile() or SaveToFile() operations.
func (c *LocalXmlClient) GetFilepath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.filepath
}

// getLogger returns the logger or a noop logger if logger is nil
func (c *LocalXmlClient) getLogger() *categoryLogger {
	if c.logger != nil {
		return c.logger
	}
	// Return a noop logger that discards all logging
	return newCategoryLogger(slog.New(discardHandler{}), 0)
}

func (c *LocalXmlClient) GetXmlDocument() string {
	return c.rootNode.OutputXML(true)
}

// autoSaveIfEnabled saves the configuration to file if auto-save is enabled.
//
// CRITICAL: Caller MUST hold write lock before calling.
// This is called internally after successful CRUD operations (set/edit/delete)
// within Communicate() which already holds the write lock.
//
// It performs a no-op (returns nil) if auto-save is disabled or filepath is not set.
//
// Note: This method is for single operations only. MultiConfig() handles
// auto-save internally using a different pattern to avoid lock conflicts.
func (c *LocalXmlClient) autoSaveIfEnabled() error {
	// Skip if auto-save disabled
	if !c.autoSave {
		return nil
	}

	// Validate filepath is set
	if c.filepath == "" {
		return fmt.Errorf("auto-save enabled but filepath not set")
	}

	// Save using internal method (assumes caller holds lock)
	return c.saveToFileInternal(c.filepath)
}

// isInitialized checks if the client has been initialized via Setup().
// CRITICAL: Caller MUST hold either read or write lock before calling.
// This is an internal helper used by CRUD operations within Communicate().
func (c *LocalXmlClient) isInitialized() bool {
	return c.rootNode != nil
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
		return c.handleRead(ctx, config, strip, ans)

	case "set":
		return c.handleSet(ctx, config, strip, ans)

	case "edit":
		return c.handleEdit(ctx, config, strip, ans)

	case "delete":
		return c.handleDelete(ctx, config, strip, ans)

	case "rename":
		return c.handleRename(ctx, config, strip, ans)

	case "move":
		return c.handleMove(ctx, config, strip, ans)

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
func (c *LocalXmlClient) handleRead(ctx context.Context, config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	start := time.Now()
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)
	timingLogger := logger.WithLogCategory(LogCategoryTimings)

	// Operation start
	opLogger.DebugContext(ctx, "Starting operation",
		"operation", "read",
		"xpath", config.Xpath,
	)

	defer func() {
		// Operation end
		opLogger.DebugContext(ctx, "Completed operation",
			"operation", "read",
			"xpath", config.Xpath,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if !c.isInitialized() {
		return nil, nil, fmt.Errorf("client not initialized: call Setup() before performing operations")
	}

	// Phase: prepare (parse and validate XPath)
	phaseStart := time.Now()
	xpath := config.Xpath
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "read",
		"phase", "prepare",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: execute (XPath evaluation and node traversal)
	phaseStart = time.Now()
	nodes := xmlquery.Find(c.rootNode, xpath)
	if len(nodes) == 0 {
		// Return empty but valid XML - matches real API behavior
		// Callers can still unmarshal to get empty slice while checking IsObjectNotFound
		emptyXml := c.formatResponse([]*xmlquery.Node{}, strip, false)
		return emptyXml, c.mockHttpResponse(404), pangoerrors.ObjectNotFound()
	}

	// Check if XPath selects attributes (e.g., /@name)
	attributeQuery := len(nodes) > 0 && nodes[0].Type == xmlquery.AttributeNode

	// If XPath selects attributes, return parent elements instead
	// This is needed for ListWithXpath which expects entry elements, not attributes
	if attributeQuery {
		var parentNodes []*xmlquery.Node
		for _, node := range nodes {
			if node.Parent != nil {
				parentNodes = append(parentNodes, node.Parent)
			}
		}
		nodes = parentNodes
	}

	if len(nodes) == 0 {
		// Return empty but valid XML after attribute query processing
		emptyXml := c.formatResponse([]*xmlquery.Node{}, strip, false)
		return emptyXml, c.mockHttpResponse(404), pangoerrors.ObjectNotFound()
	}

	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "read",
		"phase", "execute",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: finalize (serialize result)
	phaseStart = time.Now()
	// Format response (attributeQuery=true means only output element tags with attributes)
	responseXml := c.formatResponse(nodes, strip, attributeQuery)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "read",
		"phase", "finalize",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	return responseXml, c.mockHttpResponse(200), nil
}

// handleSet executes SET operation (create or overwrite)
func (c *LocalXmlClient) handleSet(ctx context.Context, config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	start := time.Now()
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)
	timingLogger := logger.WithLogCategory(LogCategoryTimings)

	// Operation start
	opLogger.DebugContext(ctx, "Starting operation",
		"operation", "set",
		"xpath", config.Xpath,
	)

	defer func() {
		// Operation end
		opLogger.DebugContext(ctx, "Completed operation",
			"operation", "set",
			"xpath", config.Xpath,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if !c.isInitialized() {
		return nil, nil, fmt.Errorf("client not initialized: call Setup() before performing operations")
	}

	// Phase: prepare (parse element XML)
	phaseStart := time.Now()
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
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "set",
		"phase", "prepare",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: execute (create/replace node)
	phaseStart = time.Now()
	// Ensure parent path exists, creating all intermediate elements if needed
	// For SET, we create the full path including the final container element
	parent, err := ensureXpathExists(c.rootNode, config.Xpath, false)
	if err != nil {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("failed to ensure XPath exists: %w", err))
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
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "set",
		"phase", "execute",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: finalize (serialize result)
	phaseStart = time.Now()
	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{newElement}, strip, false)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "set",
		"phase", "finalize",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Auto-save if enabled (operation succeeded, save to disk)
	if err := c.autoSaveIfEnabled(); err != nil {
		return responseXml, c.mockHttpResponse(200), fmt.Errorf("operation succeeded but auto-save failed: %w", err)
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// extractXpathTarget parses the final segment of an XPath to extract element name and attributes
// Example: ".../entry[@name='foo'][@type='bar']" -> ("entry", {"name": "foo", "type": "bar"})
func extractXpathTarget(xpath string) (elementName string, attributes map[string]string) {
	attributes = make(map[string]string)

	// Get last segment after final /
	lastSlash := strings.LastIndex(xpath, "/")
	if lastSlash == -1 {
		return xpath, attributes
	}
	segment := xpath[lastSlash+1:]

	// Extract element name (before [ or end of string)
	bracketPos := strings.Index(segment, "[")
	if bracketPos == -1 {
		return segment, attributes
	}
	elementName = segment[:bracketPos]

	// Extract attributes from predicates [@attr='value']
	predicates := segment[bracketPos:]
	for {
		atPos := strings.Index(predicates, "@")
		if atPos == -1 {
			break
		}
		predicates = predicates[atPos+1:]

		eqPos := strings.Index(predicates, "=")
		if eqPos == -1 {
			break
		}
		attrName := predicates[:eqPos]

		// Find quoted value
		quoteStart := strings.IndexAny(predicates, "\"'")
		if quoteStart == -1 {
			break
		}
		quote := predicates[quoteStart]
		quoteEnd := strings.Index(predicates[quoteStart+1:], string(quote))
		if quoteEnd == -1 {
			break
		}
		attrValue := predicates[quoteStart+1 : quoteStart+1+quoteEnd]

		attributes[attrName] = attrValue
		predicates = predicates[quoteStart+1+quoteEnd+1:]
	}

	return elementName, attributes
}

// getAttrValue returns the value of an attribute by name, or empty string if not found
func getAttrValue(node *xmlquery.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Name.Local == attrName {
			return attr.Value
		}
	}
	return ""
}

// nodeMatchesTarget checks if a node matches the target element with required attributes
func nodeMatchesTarget(node *xmlquery.Node, targetElement string, targetAttrs map[string]string) bool {
	if node.Data != targetElement {
		return false
	}

	// Check all required attributes match
	for attrName, expectedValue := range targetAttrs {
		found := false
		for _, attr := range node.Attr {
			if attr.Name.Local == attrName && attr.Value == expectedValue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// handleEdit executes EDIT operation (merge fields)
func (c *LocalXmlClient) handleEdit(ctx context.Context, config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	start := time.Now()
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)
	timingLogger := logger.WithLogCategory(LogCategoryTimings)

	// Operation start
	opLogger.DebugContext(ctx, "Starting operation",
		"operation", "edit",
		"xpath", config.Xpath,
	)

	defer func() {
		// Operation end
		opLogger.DebugContext(ctx, "Completed operation",
			"operation", "edit",
			"xpath", config.Xpath,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if !c.isInitialized() {
		return nil, nil, fmt.Errorf("client not initialized: call Setup() before performing operations")
	}

	// Phase: prepare (parse element XML)
	phaseStart := time.Now()
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Ensure target path exists, creating intermediate container elements AND the entry if needed
	// Per PAN-OS API contract: EDIT creates missing containers and the entry itself if it doesn't exist
	target, err := ensureXpathExists(c.rootNode, config.Xpath, false)
	if err != nil {
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
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "edit",
		"phase", "prepare",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: execute (merge node attributes/children)
	phaseStart = time.Now()

	// Extract XPath target info to detect wrapper matching
	targetElement, targetAttrs := extractXpathTarget(config.Xpath)

	// Determine what to merge: wrapper's children or rootElement's children
	var childrenToMerge []*xmlquery.Node
	firstChild := rootElement.FirstChild
	var nextSibling *xmlquery.Node
	if firstChild != nil {
		nextSibling = firstChild.NextSibling
	}

	// Check if Element is a single wrapper that matches XPath target
	if firstChild != nil && firstChild.Type == xmlquery.ElementNode && nextSibling == nil &&
	   nodeMatchesTarget(firstChild, targetElement, targetAttrs) &&
	   nodeMatchesTarget(target, targetElement, targetAttrs) {
		// Element wrapper matches XPath target - extract wrapper's children to avoid duplication
		// Example: XPath=.../entry[@name='X'], Element=<entry name="X"><content/></entry>
		// Extract <content/> instead of adding entire <entry>
		for child := firstChild.FirstChild; child != nil; child = child.NextSibling {
			childrenToMerge = append(childrenToMerge, child)
		}
	} else {
		// Normal case: merge all children from <root>
		for child := rootElement.FirstChild; child != nil; child = child.NextSibling {
			childrenToMerge = append(childrenToMerge, child)
		}
	}

	// Merge children into target
	for _, newChild := range childrenToMerge {
		if newChild.Type != xmlquery.ElementNode {
			continue
		}

		// Find matching child in target
		// For entry elements, match by name attribute; for other elements, match by element name only
		found := false
		for targetChild := target.FirstChild; targetChild != nil; targetChild = targetChild.NextSibling {
			if targetChild.Type != xmlquery.ElementNode || targetChild.Data != newChild.Data {
				continue
			}

			// For entry elements, also check name attribute
			if newChild.Data == "entry" {
				newChildName := getAttrValue(newChild, "name")
				targetChildName := getAttrValue(targetChild, "name")
				if newChildName == "" || targetChildName == "" || newChildName != targetChildName {
					continue
				}
			}

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

		if !found {
			// Add new field - need to copy the node
			nodeCopy := copyNode(newChild)
			xmlquery.AddChild(target, nodeCopy)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "edit",
		"phase", "execute",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: finalize (serialize result)
	phaseStart = time.Now()
	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{target}, strip, false)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "edit",
		"phase", "finalize",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Auto-save if enabled (operation succeeded, save to disk)
	if err := c.autoSaveIfEnabled(); err != nil {
		return responseXml, c.mockHttpResponse(200), fmt.Errorf("operation succeeded but auto-save failed: %w", err)
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleDelete executes DELETE operation (remove element)
func (c *LocalXmlClient) handleDelete(ctx context.Context, config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	start := time.Now()
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)
	timingLogger := logger.WithLogCategory(LogCategoryTimings)

	// Operation start
	opLogger.DebugContext(ctx, "Starting operation",
		"operation", "delete",
		"xpath", config.Xpath,
	)

	defer func() {
		// Operation end
		opLogger.DebugContext(ctx, "Completed operation",
			"operation", "delete",
			"xpath", config.Xpath,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if !c.isInitialized() {
		return nil, nil, fmt.Errorf("client not initialized: call Setup() before performing operations")
	}

	// Phase: prepare (validate request)
	phaseStart := time.Now()
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "delete",
		"phase", "prepare",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: execute (find and remove node)
	phaseStart = time.Now()
	// Find target element
	target := xmlquery.FindOne(c.rootNode, config.Xpath)
	if target == nil {
		return nil, nil, pangoerrors.NewErrObjectNotFound(config.Xpath)
	}

	// Remove from tree
	xmlquery.RemoveFromTree(target)
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "delete",
		"phase", "execute",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: finalize (serialize result)
	phaseStart = time.Now()
	// Format empty response (element was deleted)
	responseXml := c.formatResponse([]*xmlquery.Node{}, strip, false)

	// Unmarshal if ans struct provided (though response is empty)
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "delete",
		"phase", "finalize",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Auto-save if enabled (operation succeeded, save to disk)
	if err := c.autoSaveIfEnabled(); err != nil {
		return responseXml, c.mockHttpResponse(200), fmt.Errorf("operation succeeded but auto-save failed: %w", err)
	}

	return responseXml, c.mockHttpResponse(200), nil
}

// handleRename executes RENAME operation (change @name attribute)
func (c *LocalXmlClient) handleRename(ctx context.Context, config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	start := time.Now()
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)
	timingLogger := logger.WithLogCategory(LogCategoryTimings)

	// Operation start
	opLogger.DebugContext(ctx, "Starting operation",
		"operation", "rename",
		"xpath", config.Xpath,
		"new_name", config.NewName,
	)

	defer func() {
		// Operation end
		opLogger.DebugContext(ctx, "Completed operation",
			"operation", "rename",
			"xpath", config.Xpath,
			"new_name", config.NewName,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if !c.isInitialized() {
		return nil, nil, fmt.Errorf("client not initialized: call Setup() before performing operations")
	}

	// Phase: prepare (validate names)
	phaseStart := time.Now()
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
	}

	// Get new name
	newName := config.NewName
	if newName == "" {
		return nil, nil, pangoerrors.NewErrInvalidXpath(config.Xpath,
			fmt.Errorf("new name is required for rename operation"))
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "rename",
		"phase", "prepare",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: execute (find node and update name attribute)
	phaseStart = time.Now()
	// Find target element
	target := xmlquery.FindOne(c.rootNode, config.Xpath)
	if target == nil {
		return nil, nil, pangoerrors.NewErrObjectNotFound(config.Xpath)
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
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "rename",
		"phase", "execute",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: finalize (serialize result)
	phaseStart = time.Now()
	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{target}, strip, false)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "rename",
		"phase", "finalize",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	return responseXml, c.mockHttpResponse(200), nil
}

// handleMove executes MOVE operation (reorder element within parent)
func (c *LocalXmlClient) handleMove(ctx context.Context, config *xmlapi.Config, strip bool, ans any) ([]byte, *http.Response, error) {
	start := time.Now()
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)
	timingLogger := logger.WithLogCategory(LogCategoryTimings)

	// Operation start
	opLogger.DebugContext(ctx, "Starting operation",
		"operation", "move",
		"xpath", config.Xpath,
		"where", config.Where,
		"dst", config.Destination,
	)

	defer func() {
		// Operation end
		opLogger.DebugContext(ctx, "Completed operation",
			"operation", "move",
			"xpath", config.Xpath,
			"where", config.Where,
			"dst", config.Destination,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if !c.isInitialized() {
		return nil, nil, fmt.Errorf("client not initialized: call Setup() before performing operations")
	}

	// Phase: prepare (validate parameters)
	phaseStart := time.Now()
	// Validate XPath
	if err := c.validateXpath(config.Xpath); err != nil {
		return nil, nil, err
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
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "move",
		"phase", "prepare",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: execute (find nodes and reorder)
	phaseStart = time.Now()
	// Find target element
	target := xmlquery.FindOne(c.rootNode, config.Xpath)
	if target == nil {
		return nil, nil, pangoerrors.NewErrObjectNotFound(config.Xpath)
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
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "move",
		"phase", "execute",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	// Phase: finalize (serialize result)
	phaseStart = time.Now()
	// Format response
	responseXml := c.formatResponse([]*xmlquery.Node{target}, strip, false)

	// Unmarshal if ans struct provided
	if ans != nil {
		if err := xml.Unmarshal(responseXml, ans); err != nil {
			return responseXml, c.mockHttpResponse(200), fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	timingLogger.DebugContext(ctx, "Operation phase",
		"operation", "move",
		"phase", "finalize",
		"duration_ms", time.Since(phaseStart).Milliseconds(),
	)

	return responseXml, c.mockHttpResponse(200), nil
}

// formatResponse constructs a PAN-OS API-compatible XML response
// When attributeQuery is true, only outputs element tags with attributes (self-closing),
// not the full element tree. This is used for XPath attribute queries like /@name.
func (c *LocalXmlClient) formatResponse(nodes []*xmlquery.Node, strip bool, attributeQuery bool) []byte {
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

	// Serialize each node
	for _, node := range nodes {
		if attributeQuery {
			// For attribute queries, only output the element tag with attributes (self-closing)
			// e.g., <entry name="test-addr-00000"/>
			buf.WriteString("<")
			buf.WriteString(node.Data)
			for _, attr := range node.Attr {
				buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, attr.Value))
			}
			buf.WriteString("/>")
		} else {
			// Normal mode: output full element tree
			buf.WriteString(node.OutputXML(true))
		}
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
	logger := c.getLogger()
	opLogger := logger.WithLogCategory(LogCategoryOp)

	// Validate input
	if mc == nil || len(mc.Operations) == 0 {
		opLogger.DebugContext(ctx, "MultiConfig: empty operation batch")
		// Return empty successful response for empty batch
		mcResp := &xmlapi.MultiConfigResponse{
			Status:  "success",
			Code:    20,
			Results: []xmlapi.MultiConfigResponseElement{},
		}
		return []byte("<response status=\"success\"></response>"), c.mockHttpResponse(200), mcResp, nil
	}

	opLogger.InfoContext(ctx, "MultiConfig: starting batch",
		"operation_count", len(mc.Operations))

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
	opLogger.DebugContext(ctx, "MultiConfig: creating working copy")
	workingCopy := copyNode(c.rootNode)

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

		opLogger.InfoContext(ctx, "MultiConfig: executing operation",
			"index", i,
			"action", config.Action,
			"xpath", config.Xpath)

		// Execute operation on working copy
		err := c.executeOperationOnDoc(ctx, workingCopy, config)
		if err != nil {
			opLogger.ErrorContext(ctx, "MultiConfig: operation failed",
				"index", i,
				"action", config.Action,
				"xpath", config.Xpath,
				"error", err)
			// Operation failed - discard working copy and return error with index
			return nil, nil, nil, pangoerrors.NewErrOperationFailed(i, err)
		}

		opLogger.DebugContext(ctx, "MultiConfig: operation succeeded",
			"index", i,
			"action", config.Action,
			"xpath", config.Xpath)
	}

	// All operations succeeded - commit working copy to backing document
	opLogger.InfoContext(ctx, "MultiConfig: all operations succeeded, committing changes")
	c.commitWorkingCopy(workingCopy)

	// Auto-save if enabled (still holding lock)
	if c.autoSave && c.filepath != "" {
		opLogger.DebugContext(ctx, "MultiConfig: auto-saving to file", "filepath", c.filepath)
		if err := c.saveToFileInternal(c.filepath); err != nil {
			// Return save-specific error
			// Operations succeeded but file write failed - user decides how to proceed
			return nil, nil, nil, fmt.Errorf("operation succeeded but auto-save failed: %w", err)
		}
	}

	// Format response (empty result for multiconfig)
	responseXml := c.formatResponse([]*xmlquery.Node{}, false, false)

	// Create MultiConfigResponse
	mcResp := &xmlapi.MultiConfigResponse{
		Status:  "success",
		Code:    20,
		Results: []xmlapi.MultiConfigResponseElement{},
	}

	return responseXml, c.mockHttpResponse(200), mcResp, nil
}

// ChunkedMultiConfig executes a multi-config operation in batches.
// This method splits the operations into chunks based on mc.BatchSize and
// executes each chunk sequentially via MultiConfig.
//
// Parameters:
//   - ctx: Context for cancellation
//   - mc: MultiConfig containing operations and BatchSize
//   - strict: Unused in local mode (preserved for interface compatibility)
//   - extras: Unused in local mode (preserved for interface compatibility)
//
// Returns:
//   - Array of ChunkedMultiConfigResponse, one per batch executed
//   - Error if any batch fails (subsequent batches will not execute)
//
// Note: Unlike API mode, local mode executes batches sequentially with full
// atomicity per batch. Each batch either succeeds completely or fails completely.
func (c *LocalXmlClient) ChunkedMultiConfig(ctx context.Context, mc *xmlapi.MultiConfig, strict bool, extras url.Values) ([]xmlapi.ChunkedMultiConfigResponse, error) {
	if mc.BatchSize == 0 {
		return nil, fmt.Errorf("cannot use ChunkedMultiConfig() with batchSize of 0")
	}

	if len(mc.Operations) == 0 {
		return nil, nil
	}

	var chunked [][]xmlapi.MultiConfigOperation
	for i := 0; i < len(mc.Operations); i += mc.BatchSize {
		end := i + mc.BatchSize
		if end > len(mc.Operations) {
			end = len(mc.Operations)
		}

		chunked = append(chunked, mc.Operations[i:end])
	}

	var response []xmlapi.ChunkedMultiConfigResponse
	for _, chunk := range chunked {
		updates := &xmlapi.MultiConfig{
			Operations: chunk,
		}

		data, resp, mcResp, err := c.MultiConfig(ctx, updates, strict, extras)
		if err != nil {
			return response, err
		}

		response = append(response, xmlapi.ChunkedMultiConfigResponse{
			Data:                data,
			HttpResponse:        resp,
			MultiConfigResponse: mcResp,
		})
	}

	return response, nil
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

// RetrieveSystemInfo returns an error (not supported in local mode)
// System information retrieval requires API connection to live PAN-OS device.
func (c *LocalXmlClient) RetrieveSystemInfo(ctx context.Context) error {
	return ErrUnsupportedOperation
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
//	workingCopy := copyNode(c.rootNode)
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
func (c *LocalXmlClient) executeOperationOnDoc(ctx context.Context, doc *xmlquery.Node, config *xmlapi.Config) error {
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
		_, _, err := tempClient.handleSet(ctx, config, false, nil)
		return err

	case "edit":
		_, _, err := tempClient.handleEdit(ctx, config, false, nil)
		return err

	case "delete":
		_, _, err := tempClient.handleDelete(ctx, config, false, nil)
		return err

	case "rename":
		_, _, err := tempClient.handleRename(ctx, config, false, nil)
		return err

	case "move":
		_, _, err := tempClient.handleMove(ctx, config, false, nil)
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
