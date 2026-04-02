package util

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/PaloAltoNetworks/pango/plugin"
	"github.com/PaloAltoNetworks/pango/version"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

// PangoClient interface.
type PangoClient interface {
 	// Setup initializes the client and establishes necessary connections or loads
 	// required data. This method should be called after client creation and before
 	// performing any CRUD operations.
 	//
 	// For LocalXmlClient: loads the XML file from the filepath specified in constructor
 	// For ApiClient: establishes connection to PAN-OS device and retrieves system info
 	//
 	// Example:
 	//
 	//	client, err := pango.NewLocalXmlClient("/path/to/config.xml")
 	//	if err != nil {
 	//	    return err
 	//	}
 	//
 	//	if err := client.Setup(); err != nil {
 	//	    return fmt.Errorf("setup failed: %w", err)
 	//	}
 	//
 	//	// Client ready for operations
 	Setup() error
 
+	// Initialize performs post-Setup initialization (API mode: retrieves system info).
+	// For LocalXmlClient: no-op (all initialization done in Setup)
+	// For ApiClient: retrieves system information and performs authentication
+	Initialize(context.Context) error
+
+	// SetupLocalInspection configures client for local inspection mode (API client only).
+	// For LocalXmlClient: returns unsupported operation error
+	// For ApiClient: loads PAN-OS config and version information from provided schema
+	SetupLocalInspection(schema, panosVersion string) error
+
>>>>>>> conflict 1 of 1 ends
	// Basics.
	Versioning() version.Number
	Plugins(context.Context) ([]plugin.Info, error)
	GetTarget() string
	IsPanorama() (bool, error)
	IsFirewall() (bool, error)
	Clock(context.Context) (time.Time, error)
	RetrieveSystemInfo(context.Context) error

	// Local inspection mode functions.
	ReadFromConfig(context.Context, []string, bool, any) ([]byte, error)

	// Communication functions.
	StartJob(context.Context, PangoCommand) (uint, []byte, *http.Response, error)
	Communicate(context.Context, PangoCommand, bool, any) ([]byte, *http.Response, error)

	// Polling functions.
	WaitForJob(context.Context, uint, time.Duration, any) error
	WaitForLogs(context.Context, uint, time.Duration, any) ([]byte, error)

	// Specialized communication functions around specific XPI API commands.
	MultiConfig(context.Context, *xmlapi.MultiConfig, bool, url.Values) ([]byte, *http.Response, *xmlapi.MultiConfigResponse, error)
	ChunkedMultiConfig(context.Context, *xmlapi.MultiConfig, bool, url.Values) ([]xmlapi.ChunkedMultiConfigResponse, error)
	ImportFile(context.Context, *xmlapi.Import, []byte, string, string, bool, any) ([]byte, *http.Response, error)
	ExportFile(context.Context, *xmlapi.Export, any) (string, []byte, *http.Response, error)

	// Operational functions in use by one or more resources / data sources / namespaces.
	GenerateApiKey(context.Context, string, string) (string, error)
	RequestPasswordHash(context.Context, string) (string, error)
	GetTechSupportFile(context.Context) (string, []byte, error)
}
