package terraform

const resourceTemplateStr = `
{{- /* Begin */ -}}
// GenerateTerraformResource object

var (
	_ resource.GenerateTerraformResource                = &{{ structName }}ObjectResource{}
	_ resource.ResourceWithConfigure   = &{{ structName }}ObjectResource{}
	_ resource.ResourceWithImportState = &{{ structName }}ObjectResource{}
)

func New{{ structName }}GenerateTerraformResource() resource.GenerateTerraformResource {
	return &{{ structName }}GenerateTerraformResource{}
}

// ExampleResource defines the resource implementation.
type {{ structName }}GenerateTerraformResource struct {
	client *pango.XmlApiClient
}

type {{ structName }}ObjectLocation struct {
// TODO: Location struct via function
}

func (o *{{ structName }}ObjectLocation) IsValid() error {
	if o.Name == "" {
		return fmt.Errorf("name is unspecified")
	}

	return o.Location.IsValid()
}

type {{ structName }}EntryModel struct {
// TODO: Entry model struct via function
}

type nestedLocationModel struct {
// TODO: Nested Location model struct via function
}

type nestedVsysLocation struct {
// TODO: Nested Vsys Location model struct via function
}

type nestedDeviceGroupLocation struct {
// TODO: Nested DeviceGroup Location model struct via function
}

func (r *{{ structName }}GenerateTerraformResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (r *{{ structName }}GenerateTerraformResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
// TODO: Fill schema via function
}

func (r *{{ structName }}GenerateTerraformResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*pango.XmlApiClient)
	
	//TODO: There should be some error handling
	//if !ok {
	//	resp.Diagnostics.AddError(
	//		"Unexpected GenerateTerraformResource Configure Type",
	//		fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	//	)
	//
	//	return
	//}
	//
	//r.client = client
}

func (r *{{ structName }}GenerateTerraformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data {{ structName }}EntryModel

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

func (r *{{ structName }}GenerateTerraformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data {{ structName }}EntryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *{{ structName }}GenerateTerraformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data {{ structName }}EntryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *{{ structName }}GenerateTerraformResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data {{ structName }}EntryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *{{ structName }}GenerateTerraformResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tfid"), req, resp)
}

{{- /* Done */ -}}`
