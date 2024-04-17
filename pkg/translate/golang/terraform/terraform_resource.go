package terraform

import (
	"fmt"
	_ "sort"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

func Resource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, "Resource", false)

	resourcetemplateFunctions := template.FuncMap{
		"metaName":   func() string { return metaName },
		"structName": func() string { return structName },
	}
	resourceTemplate := template.Must(
		template.New(
			fmt.Sprintf("terraform-resource-%s", tfName),
		).Funcs(
			resourcetemplateFunctions,
		).Parse(`
{{- /* Begin */ -}}
// Resource object

var (
	_ resource.Resource                = &{{ structName }}ObjectResource{}
	_ resource.ResourceWithConfigure   = &{{ structName }}ObjectResource{}
	_ resource.ResourceWithImportState = &{{ structName }}ObjectResource{}
)

func New{{ structName }}Resource() resource.Resource {
	return &{{ structName }}Resource{}
}

// ExampleResource defines the resource implementation.
type {{ structName }}Resource struct {
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

func (r *ExampleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (r *ExampleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
// TODO: Fill schema via function
}

func (r *ExampleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExampleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

func (r *ExampleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data {{ structName }}EntryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExampleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data {{ structName }}EntryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExampleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data {{ structName }}EntryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *ExampleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tfid"), req, resp)
}

{{- /* Done */ -}}`,
		),
	)
	var renderedTemplate strings.Builder
	if err := resourceTemplate.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	terraformProvider.Code = renderedTemplate
	terraformProvider.Resources = append(terraformProvider.Resources, structName)
	return nil
}
