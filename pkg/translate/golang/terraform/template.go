package terraform

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
	client *pango.XmlApiClient
}

type {{ structName }}Tfid struct {
	//TODO: Generate tfid struct via function
}

func (o *{{ structName }}Tfid) IsValid() error {
	if o.Name == "" {
		return fmt.Errorf("name is unspecified")
	}

	return o.Location.IsValid()
}

type {{ structName }}Location struct {
// TODO: Generate Location struct via function
}

type {{ structName }}VsysLocation struct {
// TODO: Generate Location struct via function
}

type {{ structName }}DeviceGroupLocation struct {
// TODO: Generate Device Group struct via function
}

type {{ structName }}Model struct {
// TODO: Entry model struct via function
}

func (r *{{ structName }}) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (r *{{ structName }}) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
// TODO: Fill schema via function
}

func (r *{{ structName }}) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*pango.XmlApiClient)
	
	//TODO: There should be some error handling
	//if !ok {
	//	resp.Diagnostics.AddError(
	//		"Unexpected Resource Configure Type",
	//		fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	//	)
	//
	//	return
	//}
	//
	//r.client = client
}

func (r *{{ structName }}) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data {{ structName }}Model

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *{{ structName }}) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data {{ structName }}Model

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *{{ structName }}) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data {{ structName }}Model

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *{{ structName }}) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data {{ structName }}Model

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *{{ structName }}) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tfid"), req, resp)
}

{{- /* Done */ -}}`

const dataSourceTemplateStr = `
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
    client *pango.XmlApiClient
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

	d.client = req.ProviderData.(*pango.XmlApiClient)
}

func (d *{{ structName }}) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data {{ structName }}Model

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

{{- /* Done */ -}}
`
const providerFileTemplateStr = `
{{- /* Begin */ -}}
package provider

// Note:  This file is automatically generated.  Manually made changes
// will be overwritten when the provider is generated.

{{ renderImports "terraform_provider_file" structSDKLocation }}

{{- /* Done */ -}}
`
