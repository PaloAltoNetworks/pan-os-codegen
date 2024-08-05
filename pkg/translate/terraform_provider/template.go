package terraform_provider

const locationStructFields = `
{{- range .Fields }}
	{{ .Name }} {{ .Type }} ` + "`tfsdk:\"{{ .TagName }}\"`" + `
{{- end }}`

const resourceModelNestedStruct = `
type {{ .structName }}Object struct {
{{- if .HasEntryName }}
	Name types.String ` + "`" + `tfsdk:"name"` + "`" + `
{{- end }}
	{{- range $pName, $pParam := $.Spec.Params -}}
		{{- structItems $pName $pParam -}}
	{{- end}}
	{{- range $pName, $pParam := $.Spec.OneOf -}}
		{{- structItems $pName $pParam  -}}
	{{- end}}
}
`

const resourceConfigEntry = `
{{- range .Entries }}
	{{- if eq .Type "list" }}
	resp.Diagnostics.Append(state.{{ .Name }}.ElementsAs(ctx, &obj.{{ .Name }}, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{- else }}
	// {{ .Type }}
		{{- if eq .Type "object" }}
	// obj.{{ .Name }} = copy{{ .Name }}FromTerraformToPango(state.{{ .Name }})
		{{- else }}
	obj.{{ .Name }} = state.{{ .Name }}.Value{{ CamelCaseType .Type }}Pointer()
		{{- end }}
	{{- end -}}
{{- end }}
`

const resourceSchemaLocationAttribute = `
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

const resourceObj = `
{{- /* Begin */ -}}

// Generate Terraform Resource object
var (
	_ resource.Resource                = &{{ resourceStructName }}{}
	_ resource.ResourceWithConfigure   = &{{ resourceStructName }}{}
	_ resource.ResourceWithImportState = &{{ resourceStructName }}{}
)

func New{{ resourceStructName }}() resource.Resource {
	return &{{ resourceStructName }}{}
}

type {{ resourceStructName }} struct {
	client *pango.Client
}


type {{ resourceStructName }}Tfid struct {
	{{ CreateTfIdStruct }}
}


func (o *{{ resourceStructName }}Tfid) IsValid() error {
{{- if .HasEntryName }}
	if o.Name == "" {
		return fmt.Errorf("name is unspecified")
	}
{{- end }}
	return o.Location.IsValid()
}


func {{ resourceStructName }}LocationSchema() rsschema.Attribute {
	return {{ structName }}LocationSchema()
}

type {{ resourceStructName }}Model struct {
		{{ CreateTfIdResourceModel }}
{{- if .HasEntryName }}
		Name types.String` + "`" + `tfsdk:"name"` + "`" + `
{{- end }}
        {{- range $pName, $pParam := $.Spec.Params}}
            {{- ParamToModelResource $pName $pParam resourceStructName -}}
        {{- end}}
        {{- range $pName, $pParam := $.Spec.OneOf}}
            {{- ParamToModelResource $pName $pParam resourceStructName -}}
        {{- end}}

	{{- if .HasEncryptedResources }}
		EncryptedValues types.Map ` + "`" + `tfsdk:"encrypted_values"` + "`" + `
	{{- end }}
}

{{- range $pName, $pParam := $.Spec.Params}}
	{{ ModelNestedStruct $pName $pParam resourceStructName }}
{{- end}}
{{- range $pName, $pParam := $.Spec.OneOf}}
	{{ ModelNestedStruct $pName $pParam resourceStructName }}
{{- end}}

func (r *{{ resourceStructName }}) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

// <ResourceSchema>
{{ RenderResourceSchema }}

func (r *{{ resourceStructName }}) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = {{ resourceStructName }}Schema()
}

// </ResourceSchema>

func (r *{{ resourceStructName }}) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*pango.Client)
}

{{ RenderCopyToPangoFunctions }}

{{ RenderCopyFromPangoFunctions }}

func (r *{{ resourceStructName }}) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	{{ ResourceCreateFunction resourceStructName serviceName}}
}

func (o *{{ resourceStructName }}) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	{{ ResourceReadFunction resourceStructName serviceName}}
}


func (r *{{ resourceStructName }}) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	{{ ResourceUpdateFunction resourceStructName serviceName}}
}

func (r *{{ resourceStructName }}) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	{{ ResourceDeleteFunction resourceStructName serviceName}}
}

func (r *{{ resourceStructName }}) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tfid"), req, resp)
}

{{- /* Done */ -}}`

const resourceCreateFunction = `
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
{{- if .HasEntryName }}
		"name":          state.Name.ValueString(),
{{- end }}
	})

	// Verify mode.
	if r.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(r.client)

	// Determine the location.
{{- if .HasEntryName }}
	loc := {{ .structName }}Tfid{Name: state.Name.ValueString()}
{{- else }}
	loc := {{ .structName }}Tfid{}
{{- end }}


	// TODO: this needs to handle location structure for UUID style shared has nested structure type
	{{ RenderLocationsStateToPango "state" }}

	if err := loc.IsValid(); err != nil {
		resp.Diagnostics.AddError("Invalid location", err.Error())
		return
	}

	// Load the desired config.
	var obj *{{ .resourceSDKName }}.{{ .EntryOrConfig }}

{{- if .HasEncryptedResources }}
	ev := make(map[string]types.String)
	obj, diags := state.CopyToPango(ctx, &ev)
{{- else }}
	obj, diags := state.CopyToPango(ctx, nil)
{{- end }}
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(state.Timeouts.Create))
		defer cancel()
	*/

	// Perform the operation.
	create, err := svc.Create(ctx, loc.Location, *obj)
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

{{- if .HasEncryptedResources }}
	{
		copy_diags := state.CopyFromPango(ctx, create, &ev)
		resp.Diagnostics.Append(copy_diags...)
	}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
{{- else }}
	{
		copy_diags := state.CopyFromPango(ctx, create, nil)
		resp.Diagnostics.Append(copy_diags...)
	}
{{- end }}

{{- if .HasEntryName }}
	state.Name = types.StringValue(create.Name)
{{- end }}

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

{{- /* Done */ -}}
`

const resourceReadFunction = `
{{- if eq .ResourceOrDS "DataSource" }}
	var savestate, state {{ .dataSourceStructName }}Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &savestate)...)
{{- else }}
	var savestate, state {{ .resourceStructName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &savestate)...)
{{- end }}
	if resp.Diagnostics.HasError() {
		return
	}

{{- if eq .ResourceOrDS "DataSource" }}
	var loc {{ .dataSourceStructName }}Tfid
  {{- if .HasEntryName }}
	loc.Name = *savestate.Name.ValueStringPointer()
  {{- end }}
	{{ RenderLocationsStateToPango "savestate" }}
{{- else }}
	var loc {{ .resourceStructName }}Tfid
	// Parse the location from tfid.
	if err := DecodeLocation(savestate.Tfid.ValueString(), &loc); err != nil {
		resp.Diagnostics.AddError("Error parsing tfid", err.Error())
		return
	}
{{- end }}

{{- if .HasEncryptedResources }}
	ev := make(map[string]types.String, len(savestate.EncryptedValues.Elements()))
	resp.Diagnostics.Append(savestate.EncryptedValues.ElementsAs(ctx, &ev, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
{{- end }}

	// Basic logging.
	tflog.Info(ctx, "performing resource read", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .resourceStructName }}",
		"function":      "Read",
{{- if .HasEntryName }}
		"name":          loc.Name,
{{- end }}
	})

	// Verify mode.
	if o.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(o.client)

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(savestate.Timeouts.Read))
		defer cancel()
	*/

	// Perform the operation.
{{- if .HasEntryName }}
	object, err := svc.Read(ctx, loc.Location, loc.Name, "get")
{{- else }}
	object, err := svc.Read(ctx, loc.Location, "get")
{{- end }}
	if err != nil {
		if IsObjectNotFound(err) {
{{- if eq .ResourceOrDS "DataSource" }}
			resp.Diagnostics.AddError("Error reading data", err.Error())
{{- else }}
			resp.State.RemoveResource(ctx)
{{- end }}
		} else {
			resp.Diagnostics.AddError("Error reading config", err.Error())
		}
		return
	}

{{- if .HasEncryptedResources }}
	copy_diags := state.CopyFromPango(ctx, object, &ev)
{{- else }}
	copy_diags := state.CopyFromPango(ctx, object, nil)
{{- end }}
	resp.Diagnostics.Append(copy_diags...)

	{{ RenderLocationsPangoToState "state" }}

	/*
			// Keep the timeouts.
		    // TODO: This won't work for state import.
			state.Timeouts = savestate.Timeouts
	*/

	// Save tfid to state.
	state.Tfid = savestate.Tfid

	// Save the answer to state.

{{- if .HasEncryptedResources }}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
{{- end }}


	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceUpdateFunction = `
	var plan, state {{ .structName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

{{- if .HasEncryptedResources }}
	ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
	resp.Diagnostics.Append(state.EncryptedValues.ElementsAs(ctx, &ev, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
{{- end }}

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

{{- if .HasEncryptedResources }}
	obj, copy_diags := plan.CopyToPango(ctx, &ev)
{{- else }}
	obj, copy_diags := plan.CopyToPango(ctx, nil)
{{- end }}
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(plan.Timeouts.Update))
		defer cancel()
	*/

	// Perform the operation.
{{- if .HasEntryName }}
	updated, err := svc.Update(ctx, loc.Location, *obj, loc.Name)
{{- else }}
	updated, err := svc.Update(ctx, loc.Location, *obj)
{{- end }}
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
{{- if .HasEntryName }}
	loc.Name = obj.Name
{{- end }}
	tfid, err := EncodeLocation(&loc)
	if err != nil {
		resp.Diagnostics.AddError("error creating tfid", err.Error())
		return
	}
	state.Tfid = types.StringValue(tfid)

{{- if .HasEncryptedResources }}
	copy_diags = state.CopyFromPango(ctx, updated, &ev)
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
{{- else }}
	copy_diags = state.CopyFromPango(ctx, updated, nil)
{{- end }}
	resp.Diagnostics.Append(copy_diags...)

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceDeleteFunction = `
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
{{- if .HasEntryName }}
		"name":          loc.Name,
{{- end }}
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
{{- if .HasEntryName }}
	if err := svc.Delete(ctx, loc.Location, loc.Name); err != nil && !IsObjectNotFound(err) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
	}
{{- else }}

{{- if .HasEncryptedResources }}
	ev := make(map[string]types.String)
	obj, diags := state.CopyToPango(ctx, &ev)
{{- else }}
	obj, diags := state.CopyToPango(ctx, nil)
{{- end }}
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	if err := svc.Delete(ctx, loc.Location, *obj); err != nil && !IsObjectNotFound(err) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
	}
{{- end }}
`

const commonTemplate = `
{{- RenderLocationStructs }}

{{- RenderLocationSchemaGetter }}
`

const dataSourceSingletonObj = `
{{- /* Begin */ -}}

// Generate Terraform Data Source object.
var (
    _ datasource.DataSource = &{{ dataSourceStructName }}{}
    _ datasource.DataSourceWithConfigure = &{{ dataSourceStructName }}{}
)

func New{{ dataSourceStructName }}() datasource.DataSource {
    return &{{ dataSourceStructName }}{}
}

type {{ dataSourceStructName }} struct {
    client *pango.Client
}

type {{ dataSourceStructName }}Filter struct {
//TODO: Generate Data Source filter via function
}

type {{ dataSourceStructName }}Tfid struct {
	{{ CreateTfIdStruct }}
}

func (o *{{ dataSourceStructName }}Tfid) IsValid() error {
{{- if .HasEntryName }}
	if o.Name == "" {
		return fmt.Errorf("name is unspecified")
	}
{{- end }}
	return o.Location.IsValid()
}

{{ RenderDataSourceStructs }}

{{ RenderCopyFromPangoFunctions }}

{{ RenderDataSourceSchema }}

func {{ dataSourceStructName }}LocationSchema() rsschema.Attribute {
	return {{ structName }}LocationSchema()
}

// Metadata returns the data source type name.
func (d *{{ dataSourceStructName }}) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

// Schema defines the schema for this data source.
func (d *{{ dataSourceStructName }}) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = {{ dataSourceStructName }}Schema()
}

// Configure prepares the struct.
func (d *{{ dataSourceStructName }}) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*pango.Client)
}

func (o *{{ dataSourceStructName }}) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	{{ DataSourceReadFunction dataSourceStructName serviceName }}
}

{{- /* Done */ -}}
`

const dataSourceListObj = ``

const providerFile = `
{{- /* Begin */ -}}
package provider

// Note:  This file is automatically generated.  Manually made changes
// will be overwritten when the provider is generated.
{{ renderImports }}
{{ renderCode }}

{{- /* Done */ -}}
`
const provider = `
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
