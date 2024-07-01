package terraform_provider

const locationStructTemplate = `
{{- range .Fields }}
	{{ .Name }} {{ .Type }} ` + "`tfsdk:\"{{ .TagName }}\"`" + `
{{- end }}`

const resourceModelNestedStructTemplate = `
type {{ .structName }}Object struct {
	{{- range $pName, $pParam := $.Spec.Params -}}
		{{- structItems $pName $pParam -}}
	{{- end}}
	{{- range $pName, $pParam := $.Spec.OneOf -}}
		{{- structItems $pName $pParam  -}}
	{{- end}}
}
`

const resourceEntryConfigTemplate = `
{{- range .Entries }}
	{{- if eq .Type "list" }}
	resp.Diagnostics.Append(state.{{ .Name }}.ElementsAs(ctx, &obj.{{ .Name }}, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{- else }}
	obj.{{ .Name }} = state.{{ .Name }}.Value{{ CamelCaseType .Type }}Pointer()
	{{- end -}}
{{- end }}
`

const resourceTemplateSchemaLocationAttribute = `
			"location": rsschema.SingleNestedAttribute{
				Description: "The location of this object.",
				Required:    true,
				Attributes: map[string]rsschema.Attribute{
					"device_group": rsschema.SingleNestedAttribute{
						Description: "(Panorama) In the given device group. One of the following must be specified: ` + "`" + `device_group` + "`" + `, ` + "`" + `from_panorama` + "`" + `, ` + "`" + `shared` + "`" + `, or ` + "`" + `vsys` + "`" + `.",
						Optional:    true,
						Attributes: map[string]rsschema.Attribute{
							"name": rsschema.StringAttribute{
								Description: "The device group name.",
								Required:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"panorama_device": rsschema.StringAttribute{
								Description: "The Panorama device. Default: ` + `localhost.localdomain` + `.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("localhost.localdomain"),
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"from_panorama": rsschema.BoolAttribute{
						Description: "(NGFW) Pushed from Panorama. This is a read-only location and only suitable for data sources. One of the following must be specified: ` + "`" + `device_group` + "`" + `, ` + "`" + `from_panorama` + "`" + `, ` + "`" + `shared` + "`" + `, or ` + "`" + `vsys` + "`" + `.",
						Optional:    true,
						Validators: []validator.Bool{
							boolvalidator.ExactlyOneOf(
								path.MatchRoot("location").AtName("from_panorama"),
								path.MatchRoot("location").AtName("device_group"),
								path.MatchRoot("location").AtName("vsys"),
								path.MatchRoot("location").AtName("shared"),
							),
						},
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"shared": rsschema.BoolAttribute{
						Description: "(NGFW and Panorama) Located in shared. One of the following must be specified:` + "`" + `device_group` + "`" + `, ` + "`" + `from_panorama` + "`" + `, ` + "`" + `shared` + "`" + `, or ` + "`" + `vsys` + "`" + `.",
						Optional:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"vsys": rsschema.SingleNestedAttribute{
						Description: "(NGFW) In the given vsys. One of the following must be specified:` + "`" + `device_group` + "`" + `, ` + "`" + `from_panorama` + "`" + `, ` + "`" + `shared` + "`" + `, or ` + "`" + `vsys` + "`" + `.",
						Optional:    true,
						Attributes: map[string]rsschema.Attribute{
							"name": rsschema.StringAttribute{
								Description: "The vsys name.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("vsys1"),
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"ngfw_device": rsschema.StringAttribute{
								Description: "The NGFW device.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("localhost.localdomain"),
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"tfid": rsschema.StringAttribute{
				Description: "The Terraform ID.",
				Computed:    true,
			},
`

const resourceTemplateStr = `
{{- /* Begin */ -}}

// Generate Terraform Resource object

var (
	_ resource.Resource                = &{{ structName }}{}
	_ resource.ResourceWithConfigure   = &{{ structName }}{}
	_ resource.ResourceWithImportState = &{{ structName }}{}
)

func New{{ structName }}() resource.Resource {
	return &{{ structName }}{}
}

type {{ structName }} struct {
	client *pango.Client
}

type {{ structName }}Tfid struct {
	{{ CreateTfIdStruct }}
}

func (o *{{ structName }}Tfid) IsValid() error {
	if o.Name == "" {
		return fmt.Errorf("name is unspecified")
	}

	return o.Location.IsValid()
}

type {{ structName }}Location struct {
// TODO: Generate Location struct via function
	{{- CreateLocationStruct structName }}
}

type {{ structName }}VsysLocation struct {
// TODO: Generate Location struct via function
	{{- CreateLocationVsysStruct structName }}
}

type {{ structName }}DeviceGroupLocation struct {
// TODO: Generate Device Group struct via function
	{{- CreateLocationDeviceGroupStruct structName }}
}

type {{ structName }}Model struct {
		{{ CreateTfIdResourceModel }}
		Name types.String` + "`" + `tfsdk:"name"` + "`" + `
        {{- range $pName, $pParam := $.Spec.Params}}
            {{- ParamToModelResource $pName $pParam structName -}}
        {{- end}}
        {{- range $pName, $pParam := $.Spec.OneOf}}
            {{- ParamToModelResource $pName $pParam structName -}}
        {{- end}}
}

{{- range $pName, $pParam := $.Spec.Params}}
	{{ ModelNestedStruct $pName $pParam structName }}
{{- end}}
{{- range $pName, $pParam := $.Spec.OneOf}}
	{{ ModelNestedStruct $pName $pParam structName }}
{{- end}}

func (r *{{ structName }}) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (r *{{ structName }}) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
// TODO: Fill schema via function
	resp.Schema = rsschema.Schema{
		Description: "",
		Attributes: map[string]rsschema.Attribute{
	{{- ResourceSchemaLocationAttribute }}
	"name": rsschema.StringAttribute{
		Description: "The name of the resource.",
		Required:    true,
	},	
	{{- range $pName, $pParam := $.Spec.Params -}}
		{{ ResourceParamToSchema $pName $pParam }}
	{{- end }}
	{{- range $pName, $pParam := $.Spec.OneOf -}}
		{{ ResourceParamToSchema $pName $pParam }}
	{{- end }}
		},
	}
}

func (r *{{ structName }}) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*pango.Client)
}

func (r *{{ structName }}) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	{{ ResourceCreateFunction structName serviceName}}
}

func (r *{{ structName }}) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	{{ ResourceReadFunction structName serviceName}}
}

func (r *{{ structName }}) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	{{ ResourceUpdateFunction structName serviceName}}
}

func (r *{{ structName }}) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	{{ ResourceDeleteFunction structName serviceName}}
}

func (r *{{ structName }}) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tfid"), req, resp)
}

{{- /* Done */ -}}`

const resourceCreateTemplateStr = `
{{- /* Begin */ -}}
	var state {{ .structName }}Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Basic logging.
	tflog.Info(ctx, "performing resource create", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .structName }}",
		"function":      "Create",
		"name":          state.Name.ValueString(),
	})

	// Verify mode.
	if r.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(r.client)

	// Determine the location.
	loc := {{ .structName }}Tfid{Name: state.Name.ValueString()}
	// TODO: this needs to handle location structure for UUID style shared has nested structure type 
	if !state.Location.Shared.IsNull() && state.Location.Shared.ValueBool() {
		loc.Location.Shared = true
	}
	if !state.Location.FromPanorama.IsNull() && state.Location.FromPanorama.ValueBool() {
		loc.Location.FromPanoramaShared = true
	}
	if state.Location.Vsys != nil {
		loc.Location.Vsys = &{{ .serviceName }}.VsysLocation{}
		loc.Location.Vsys.NgfwDevice = state.Location.Vsys.NgfwDevice.ValueString()
	}
	if state.Location.DeviceGroup != nil {
		loc.Location.DeviceGroup = &{{ .serviceName }}.DeviceGroupLocation{}
		loc.Location.DeviceGroup.DeviceGroup = state.Location.DeviceGroup.Name.ValueString()
		loc.Location.DeviceGroup.PanoramaDevice = state.Location.DeviceGroup.PanoramaDevice.ValueString()
	}
	if err := loc.IsValid(); err != nil {
		resp.Diagnostics.AddError("Invalid location", err.Error())
		return
	}

	// Load the desired config.
	var obj {{ .resourceSDKName }}.Entry
	obj.Name = state.Name.ValueString()
	{{- range $pName, $pParam := $.paramSpec.Params }}
		{{ LoadConfigToEntry $pParam $pName }}
	{{- end }}
	{{- range $pName, $pParam := $.paramSpec.OneOf }}
		{{ LoadConfigToEntry $pParam $pName }}
	{{- end }}

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(state.Timeouts.Create))
		defer cancel()
	*/

	// Perform the operation.
	create, err := svc.Create(ctx, loc.Location, obj)
	if err != nil {
		resp.Diagnostics.AddError("Error in create", err.Error())
		return
	}

	// Tfid handling.
	tfid, err := EncodeLocation(&loc)
	if err != nil {
		resp.Diagnostics.AddError("Error creating tfid", err.Error())
		return
	}

	// Save the state.
	state.Tfid = types.StringValue(tfid)
	state.Name = types.StringValue(create.Name)

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

{{- /* Done */ -}}
`

const resourceReadTemplateStr = `
	var savestate, state {{ .structName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &savestate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the location from tfid.
	var loc {{ .structName }}Tfid
	if err := DecodeLocation(savestate.Tfid.ValueString(), &loc); err != nil {
		resp.Diagnostics.AddError("Error parsing tfid", err.Error())
		return
	}

	// Basic logging.
	tflog.Info(ctx, "performing resource read", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .structName }}",
		"function":      "Read",
		"name":          loc.Name,
	})

	// Verify mode.
	if r.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(r.client)

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(savestate.Timeouts.Read))
		defer cancel()
	*/

	// Perform the operation.
	_, err := svc.Read(ctx, loc.Location, loc.Name, "get")
	if err != nil {
		if IsObjectNotFound(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading config", err.Error())
		}
		return
	}

	// Save location to state.
	if loc.Location.Shared {
		state.Location.Shared = types.BoolValue(true)
	}
	if loc.Location.FromPanoramaShared {
		state.Location.FromPanorama = types.BoolValue(true)
	}
	if loc.Location.Vsys != nil {
		state.Location.Vsys = &{{ .structName }}VsysLocation{}
		state.Location.Vsys.NgfwDevice = types.StringValue(loc.Location.Vsys.NgfwDevice)
	}
	if loc.Location.DeviceGroup != nil {
		state.Location.DeviceGroup = &{{ .structName }}DeviceGroupLocation{}
		state.Location.DeviceGroup.PanoramaDevice = types.StringValue(loc.Location.DeviceGroup.PanoramaDevice)
	}

	/*
			// Keep the timeouts.
		    // TODO: This won't work for state import.
			state.Timeouts = savestate.Timeouts
	*/

	// Save tfid to state.
	state.Tfid = savestate.Tfid

	// Save the answer to state.


	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceUpdateTemplateStr = `
	var plan, state {{ .structName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var loc {{ .structName }}Tfid
	if err := DecodeLocation(state.Tfid.ValueString(), &loc); err != nil {
		resp.Diagnostics.AddError("Error parsing tfid", err.Error())
		return
	}

	// Basic logging.
	tflog.Info(ctx, "performing resource update", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .structName }}",
		"function":      "Update",
		"tfid":          state.Tfid.ValueString(),
	})

	// Verify mode.
	if r.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(r.client)

	// Load the desired config.
	var obj {{ .resourceSDKName }}.Entry

	if resp.Diagnostics.HasError() {
		return
	}

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(plan.Timeouts.Update))
		defer cancel()
	*/

	// Perform the operation.
	_, err := svc.Update(ctx, loc.Location, obj, loc.Name)
	if err != nil {
		resp.Diagnostics.AddError("Error in update", err.Error())
		return
	}

	// Save the location.
	state.Location = plan.Location

	/*
		// Keep the timeouts.
		state.Timeouts = plan.Timeouts
	*/

	// Save the tfid.
	loc.Name = obj.Name
	tfid, err := EncodeLocation(&loc)
	if err != nil {
		resp.Diagnostics.AddError("error creating tfid", err.Error())
		return
	}
	state.Tfid = types.StringValue(tfid)

	// Save the state.


	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceDeleteTemplateStr = `
	var state {{ .structName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the location from tfid.
	var loc {{ .structName }}Tfid
	if err := DecodeLocation(state.Tfid.ValueString(), &loc); err != nil {
		resp.Diagnostics.AddError("error parsing tfid", err.Error())
		return
	}

	// Basic logging.
	tflog.Info(ctx, "performing resource delete", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .structName }}",
		"function":      "Delete",
		"name":          loc.Name,
	})

	// Verify mode.
	if r.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(r.client)

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(state.Timeouts.Delete))
		defer cancel()
	*/

	// Perform the operation.
	if err := svc.Delete(ctx, loc.Location, loc.Name); err != nil && !IsObjectNotFound(err) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
	}
`

const dataSourceSingletonTemplateStr = `
{{- /* Begin */ -}}

// Generate Terraform Data Source object.
var (
    _ datasource.DataSource = &{{ structName }}{}
    _ datasource.DataSourceWithConfigure = &{{ structName }}{}
)

func New{{ structName }}() datasource.DataSource {
    return &{{ structName }}{}
}

type {{ structName }} struct {
    client *pango.Client
}

type {{ structName }}Model struct {
//TODO: Generate Data Source model via function
}

type {{ structName }}Filter struct {
//TODO: Generate Data Source filter via function
}

type {{ structName }}Location struct {
//TODO: Generate Data Source Location via function
}

type {{ structName }}SharedLocation struct {
//TODO: Generate Data Source Location shared via function
}

type {{ structName }}VsysLocation struct {
//TODO: Generate Data Source Location vsys via function
}

type {{ structName }}DeviceGroupLocation struct {
//TODO: Generate Data Source Location Device Group via function
}

type {{ structName }}Entry struct {
// TODO: Entry model struct via function
}

// Metadata returns the data source type name.
func (d *{{ structName }}) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

// Schema defines the schema for this data source.
func (d *{{ structName }}) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = dsschema.Schema{
//TODO: generate schema via function
    }
}

// Configure prepares the struct.
func (d *{{ structName }}) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*pango.Client)
}

func (d *{{ structName }}) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data {{ structName }}Model

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

{{- /* Done */ -}}
`

const dataSourceListTemplatetStr = ``

const providerFileTemplateStr = `
{{- /* Begin */ -}}
package provider

// Note:  This file is automatically generated.  Manually made changes
// will be overwritten when the provider is generated.
{{ renderImports }}
{{ renderCode }}

{{- /* Done */ -}}
`
const providerTemplateStr = `
{{- /* Begin */ -}}
package provider

{{ RenderImports }}

// Ensure the provider implementation interface is sound.
var (
	_ provider.Provider = &PanosProvider{}
)

// PanosProvider is the provider implementation.
type PanosProvider struct {
	version string
}

// PanosProviderModel maps provider schema data to a Go type.
type PanosProviderModel struct {
{{- range $pName, $pParam := ProviderParams }}
{{ ParamToModelBasic $pName $pParam }}
{{- end }}
}

// Metadata returns the provider type name.
func (p *PanosProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "panos"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *PanosProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider to interact with Palo Alto Networks PAN-OS.",
		Attributes: map[string]schema.Attribute{
{{- range $pName, $pParam := ProviderParams }}
{{ ParamToSchemaProvider $pName $pParam }}
{{- end }}
		},
	}
}

// Configure prepares the provider.
func (p *PanosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring the provider client...")

	var config PanosProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var con *sdk.Client

	if config.ConfigFile.ValueStringPointer() != nil {
		tflog.Info(ctx, "Configuring client for local inspection mode")
		con = &sdk.Client{}
		if err := con.SetupLocalInspection(config.ConfigFile.ValueString(), config.PanosVersion.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting up local inspection mode", err.Error())
			return
		}
	} else {
		tflog.Info(ctx, "Configuring client for API mode")
		con = &sdk.Client{
			Hostname:        config.Hostname.ValueString(),
			Username:        config.Username.ValueString(),
			Password:        config.Password.ValueString(),
			ApiKey:          config.ApiKey.ValueString(),
			Protocol:        config.Protocol.ValueString(),
			Port:            int(config.Port.ValueInt64()),
			Target:          config.Target.ValueString(),
			ApiKeyInRequest: config.ApiKeyInRequest.ValueBool(),
			// Headers from AdditionalHeaders
			SkipVerifyCertificate: config.SkipVerifyCertificate.ValueBool(),
			AuthFile:              config.AuthFile.ValueString(),
			CheckEnvironment:      true,
			//Agent:            fmt.Sprintf("Terraform/%s Provider/scm Version/%s", req.TerraformVersion, p.version),
		}

		if err := con.Setup(); err != nil {
			resp.Diagnostics.AddError("Provider parameter value error", err.Error())
			return
		}

		//con.HttpClient.Transport = sdkapi.NewTransport(con.HttpClient.Transport, con)

		if err := con.Initialize(ctx); err != nil {
			resp.Diagnostics.AddError("Initialization error", err.Error())
			return
		}
	}

	resp.DataSourceData = con
	resp.ResourceData = con

	// Done.
	tflog.Info(ctx, "Configured client", map[string]any{"success": true})
}

// DataSources defines the data sources for this provider.
func (p *PanosProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTfidDataSource,
{{- range $fnName := DataSources }}
        New{{ $fnName }},
{{- end }}
	}
}

// Resources defines the data sources for this provider.
func (p *PanosProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
{{- range $fnName := Resources }}
        New{{ $fnName }},
{{- end }}
	}
}

// New is a helper function to get the provider implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PanosProvider{
			version: version,
		}
	}
}

{{- /* Done */ -}}`
