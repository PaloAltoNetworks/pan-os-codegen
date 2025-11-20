package terraform_provider

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
`

const resourceCreateUpdateMovementRequiredTmpl = `
// We only manage a subset of PAN-OS object on the given list, so care
// has to be taken to calculate the order of those managed elements on the
// PAN-OS side.

// We filter all existing entries to end up with a list of entries that
// are in the plan. For every element of that list, we store its PAN-OS
// list index as StateIdx. Finally, the managedEntries index will serve
// as a way to check if managed entries are in order relative to each
// other.
var movementRequired bool
managedEntriesByName := make(map[string]*entryWithState, len(planEntriesByName))
idx := 0
for existingIdx, elt := range existing {
	if planEntry, found := planEntriesByName[elt.Name]; found {
		managedEntriesByName[elt.Name] = &entryWithState{
			Entry: existing[existingIdx],
			StateIdx: idx,
		}
		planEntry.Entry.Uuid = elt.Uuid
		planEntriesByName[elt.Name] = planEntry
	}
	idx++
}

// First, we check if managedEntries order matches planEntriesByName to check
// if all entries from the plan are properly ordered on the server.
var previousManagedEntry, previousPlannedEntry *entryWithState
for _, elt := range managedEntriesByName {
	// plannedEntriesByName is a map of entries from the plan indexed by their
	// name. If idx doesn't match StateIdx of the entry from the plan, the PAN-OS
	// object is out of order.
	plannedEntry := planEntriesByName[elt.Entry.Name]
	if plannedEntry.StateIdx != elt.StateIdx {
		movementRequired = true
		break
	}
	// If this is the first element we are comparing, store it for future reference
	// and continue. We will use it to calculate distance between two elements in
	// PAN-OS list.
	if previousManagedEntry == nil {
		previousManagedEntry = elt
		previousPlannedEntry = plannedEntry
		continue
	}

	serverDistance := elt.StateIdx - previousManagedEntry.StateIdx
	planDistance := plannedEntry.StateIdx - previousPlannedEntry.StateIdx

	// If the distance between previous and current object differs between
	// PAN-OS and the plan, we need to move objects around.
	if serverDistance != planDistance {
		movementRequired = true
		break
	}

	previousManagedEntry = elt
	previousPlannedEntry = plannedEntry
}

// If all entries are ordered properly, we check if their position matches what's
// requested.
if !movementRequired {
	existingEntriesByName := make(map[string]*entryWithState, len(existing))
	for idx, elt := range existing {
		existingEntriesByName[elt.Name] = &entryWithState{
			Entry: existing[idx],
			StateIdx: idx,
		}
	}

	positionWhere := {{ .State }}.Position.Where.ValueString()
	switch positionWhere {
	case "first":
		if existing[len({{ .Entries }})-1].Name != {{ .Entries }}[len({{ .Entries }})-1].Name {
			movementRequired = true

		}
	case "last":
		if existing[len({{ .Entries }})-1].Name != {{ .Entries }}[len({{ .Entries }})-1].Name {
			movementRequired = true
		}
	case "before":
		pivot := {{ .State }}.Position.Pivot.ValueString()
		directly := {{ .State }}.Position.Directly.ValueBool()
		if existingPivot, found := existingEntriesByName[pivot]; !found {
			resp.Diagnostics.AddError("failed to create move actions", fmt.Sprintf("pivot point '%s' missing from the server", pivot))
		} else if directly {
			if existingPivot.StateIdx == 0 {
				movementRequired = true
			} else if existing[existingPivot.StateIdx-1].Name != {{ .Entries }}[len({{ .Entries }})-1].Name {
				movementRequired = true
			}
		} else {
			if existingPivot.StateIdx == 0 {
				movementRequired = true
			}
		}
	case "after":
		pivot := {{ .State }}.Position.Pivot.ValueString()
		directly := {{ .State }}.Position.Directly.ValueBool()
		if existingPivot, found := existingEntriesByName[pivot]; !found {
			resp.Diagnostics.AddError("failed to create move actions", fmt.Sprintf("pivot point '%s' missing from the server", pivot))
		} else if directly {
			if existingPivot.StateIdx == len(existing)-1 {
				movementRequired = true
			} else if existing[existingPivot.StateIdx+1].Name != {{ .Entries }}[0].Name {
				movementRequired = true
			}
		} else {
			if existingPivot.StateIdx == len(existing)-1 {
				movementRequired = true
			}
		}
	}
}
`

const actionObj = `
var (
	_ action.ActionWithConfigure = &{{ structName }}{}
)

func New{{ structName }}() action.Action {
	return &{{ structName }}{}
}

type {{ structName }} struct {
	client *pango.Client
}

{{ RenderStructs }}

{{ RenderSchema }}

func (o *{{ structName }}) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (o *{{ structName }}) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = {{ structName }}Schema()
}

func (o *{{ structName }}) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	o.InvokeCustom(ctx, req, resp)
}

func (o *{{ structName }}) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData := req.ProviderData.(*ProviderData)
	o.client = providerData.Client
}
`

const resourceObj = `
{{- /* Begin */ -}}

{{- if IsEphemeral }}
// Generate Terraform Ephemeral object
var (
	_ ephemeral.EphemeralResource              = &{{ resourceStructName }}{}
        _ ephemeral.EphemeralResourceWithConfigure = &{{ resourceStructName }}{}
)
{{- else }}
// Generate Terraform Resource object
var (
	_ resource.Resource                = &{{ resourceStructName }}{}
	_ resource.ResourceWithConfigure   = &{{ resourceStructName }}{}
	_ resource.ResourceWithImportState = &{{ resourceStructName }}{}
)
{{- end }}


{{- if IsEphemeral }}
func New{{ resourceStructName }}() ephemeral.EphemeralResource {
	return &{{ resourceStructName }}{}
}
{{- else }}
func New{{ resourceStructName }}() resource.Resource {
  {{- if IsImportable }}
	if _, found := resourceFuncMap["panos{{ metaName }}"]; !found {
		resourceFuncMap["panos{{ metaName }}"] = resourceFuncs{
			CreateImportId: {{ structName }}ImportStateCreator,
		}
	}
  {{- end }}
	return &{{ resourceStructName }}{}
}
{{- end }}

type {{ resourceStructName }} struct {
	client *pango.Client
{{- if IsCustom }}
	custom *{{ structName }}Custom
{{- else if and IsEntry HasImports }}
	manager *sdkmanager.ImportableEntryObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, {{ resourceSDKName }}.ImportLocation, *{{ resourceSDKName }}.Service]
{{- else if IsEntry }}
	manager *sdkmanager.EntryObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service]
{{- else if IsUuid }}
	manager *sdkmanager.UuidObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service]
{{- else if IsConfig }}
	manager *sdkmanager.ConfigObjectManager[*{{ resourceSDKName }}.Config, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service]
{{- end }}
}

{{- if HasLocations }}
func {{ resourceStructName }}LocationSchema() rsschema.Attribute {
	return {{ structName }}LocationSchema()
}
{{- end }}

{{ RenderResourceStructs }}

func (o *{{ resourceStructName }}) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
{{- if HasPosition }}
	{

	var resource {{ resourceStructName }}Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &resource)...)
	if resp.Diagnostics.HasError() {
		return
	}

        if !resource.Position.IsUnknown() {
		var positionAttribute TerraformPositionObject
		resp.Diagnostics.Append(resource.Position.As(ctx, &positionAttribute, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		positionAttribute.ValidateConfig(resp)
	}

	}
{{- end }}

{{- if IsUuid }}
	{
	var resource {{ resourceStructName }}Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &resource)...)
	if resp.Diagnostics.HasError() {
		return
	}
  {{ $resourceTFStructName := printf "%s%sObject" resourceStructName ListAttribute.CamelCase }}
	entries := make(map[string]struct{})
	duplicated := make(map[string]struct{})

	var elements []types.Object
	resp.Diagnostics.Append(resource.{{ ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, elt := range elements {
		var typedElt {{ $resourceTFStructName }}
		resp.Diagnostics.Append(elt.As(ctx, &typedElt, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		if typedElt.Name.IsUnknown() {
			continue
		}

		entry := typedElt.Name.ValueString()
		if _, found := entries[entry]; found {
			duplicated[entry] = struct{}{}
		}
		entries[entry] = struct{}{}
	}

	var _ = strings.Join([]string{"a", "b"}, ",")

	if len(duplicated) > 0 {
		var entries []string
		for elt := range duplicated {
			entries = append(entries, fmt.Sprintf("'%s'", elt))
		}
		resp.Diagnostics.AddError("Failed to validate resource", fmt.Sprintf("Non-unique entry names in the list: %s", strings.Join(entries, ",")))
		return
	}

	}
{{- end }}
}

// <ResourceSchema>
{{ RenderResourceSchema }}

func (o *{{ resourceStructName }}) Metadata(ctx context.Context, req {{ tfresourcepkg }}.MetadataRequest, resp *{{ tfresourcepkg }}.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (o *{{ resourceStructName }}) Schema(_ context.Context, _ {{ tfresourcepkg }}.SchemaRequest, resp *{{ tfresourcepkg }}.SchemaResponse) {
	resp.Schema = {{ resourceStructName }}Schema()
}

// </ResourceSchema>

func (o *{{ resourceStructName }}) Configure(ctx context.Context, req {{ tfresourcepkg }}.ConfigureRequest, resp *{{ tfresourcepkg }}.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData := req.ProviderData.(*ProviderData)
	o.client = providerData.Client

{{- if IsCustom }}
	custom, err := New{{ structName }}Custom(providerData)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	o.custom = custom
{{- else if and IsEntry HasImports }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(o.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}

	batchSize := providerData.MultiConfigBatchSize
	o.manager =  sdkmanager.NewImportableEntryObjectManager(o.client, {{ resourceSDKName }}.NewService(o.client), batchSize, specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsEntry }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(o.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	batchSize := providerData.MultiConfigBatchSize
	o.manager =  sdkmanager.NewEntryObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service](o.client, {{ resourceSDKName }}.NewService(o.client), batchSize, specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsUuid }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(o.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	batchSize := providerData.MultiConfigBatchSize
	o.manager =  sdkmanager.NewUuidObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service](o.client, {{ resourceSDKName }}.NewService(o.client), batchSize, specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsConfig }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(o.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	o.manager =  sdkmanager.NewConfigObjectManager(o.client, {{ resourceSDKName }}.NewService(o.client), specifier)
{{- end }}
}

{{ RenderModelAttributeTypesFunction }}

{{ RenderCopyToPangoFunctions }}

{{ RenderCopyFromPangoFunctions }}

{{ RenderXpathComponentsGetter }}

{{- if FunctionSupported "Create" }}
func (o *{{ resourceStructName }}) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	{{ ResourceCreateFunction resourceStructName serviceName}}
}
{{- end }}

{{- if FunctionSupported "Read" }}
func (o *{{ resourceStructName }}) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	{{ ResourceReadFunction resourceStructName serviceName}}
}
{{- end }}


{{- if FunctionSupported "Update" }}
func (o *{{ resourceStructName }}) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	{{ ResourceUpdateFunction resourceStructName serviceName}}
}
{{- end }}

{{- if FunctionSupported "Delete" }}
func (o *{{ resourceStructName }}) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	{{ ResourceDeleteFunction resourceStructName serviceName}}
}
{{- end }}

{{- if FunctionSupported "Open" }}
func (o *{{ resourceStructName }}) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	{{ ResourceOpenFunction resourceStructName serviceName}}
}
{{- end }}

{{- if FunctionSupported "Renew" }}
func (o *{{ resourceStructName }}) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	{{ ResourceRenewFunction resourceStructName serviceName}}
}
{{- end }}

{{- if FunctionSupported "Close" }}
func (o *{{ resourceStructName }}) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	{{ ResourceCloseFunction resourceStructName serviceName}}
}
{{- end }}

{{ RenderImportStateStructs }}

{{ RenderImportStateMarshallers }}

{{ RenderImportStateCreator }}

func (o *{{ resourceStructName }}) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	{{ ResourceImportStateFunction }}
}

{{- /* Done */ -}}`

const resourceCreateEntryListFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{ $resourceTFStructName := printf "%s%sObject" .structName .ListAttribute.CamelCase }}

var state {{ .structName }}Model
resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource create", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Create",
})

{{ RenderEncryptedValuesInitialization }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

type entryWithState struct {
	Entry    *{{ $resourceSDKStructName }}
	StateIdx int
}

{{ if eq .PluralType "map" }}
var elements map[string]{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx := 0
for name, elt := range elements {
	var entry *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	entries[idx] = entry
	idx++
}
{{ else if or (eq .PluralType "list") (eq .PluralType "set") }}
var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var entry *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = elt.Name.ValueString()
	entries[idx] = entry
}
{{- end }}

components, err := state.resourceXpathParentComponents()
if err != nil {
	resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
	return
}

created, err := o.manager.CreateMany(ctx, location, components, entries)
if err != nil {
	resp.Diagnostics.AddError("Failed to create new entries", err.Error())
	return
}

{{ if eq .PluralType "map" }}
for _, elt := range created {
	if _, found := elements[elt.Name]; !found {
		continue
	}
	var object {{ $resourceTFStructName }}
	object.name = elt.Name
	resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, elt, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	elements[elt.Name] = object
}

var map_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, map_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), elements)
resp.Diagnostics.Append(map_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{ else if or (eq .PluralType "list") (eq .PluralType "set") }}
elementsByName := make(map[string]int)
for idx, elt := range elements {
	elementsByName[elt.Name.ValueString()] = idx
}

for _, elt := range created {
	idx, found := elementsByName[elt.Name]
	if !found {
		continue
	}

	var object {{ $resourceTFStructName }}
	resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, elt, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	elements[idx] = object
}

var list_diags diag.Diagnostics
  {{ if eq .PluralType "list" }}
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), elements)
  {{ else if eq .PluralType "set" }}
state.{{ .ListAttribute.CamelCase }}, list_diags = types.SetValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), elements)
  {{- end }}
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceCreateManyFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{ $resourceTFStructName := printf "%s%sObject" .structName .ListAttribute.CamelCase }}

var state {{ .structName }}Model
resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource create", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Create",
})

{{ RenderEncryptedValuesInitialization }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries[idx] = entry
}

components, err := state.resourceXpathParentComponents()
if err != nil {
	resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
	return
}

{{- if .Exhaustive }}
processed, err := o.manager.CreateMany(ctx, location, components, entries, sdkmanager.Exhaustive, movement.PositionFirst{})
if err != nil {
	resp.Diagnostics.AddError("Error during CreateMany() call", err.Error())
	return
}
{{- else }}
var positionAttribute TerraformPositionObject
resp.Diagnostics.Append(state.Position.As(ctx, &positionAttribute, basetypes.ObjectAsOptions{})...)
if resp.Diagnostics.HasError() {
	return
}
position := positionAttribute.CopyToPango()
processed, err := o.manager.CreateMany(ctx, location, components, entries, sdkmanager.NonExhaustive, position)
if err != nil {
	resp.Diagnostics.AddError("Error during CreateMany() call", err.Error())
	return
}
{{- end }}
objects := make([]{{ $resourceTFStructName }}, len(processed))
for idx, elt := range processed {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, o.client, nil, elt, ev)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[idx] = object
}

var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}

{{ RenderEncryptedValuesFinalizer }}

resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

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
	if o.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	{{ RenderEncryptedValuesInitialization }}

	// Determine the location.

	var location {{ .resourceSDKName }}.Location
	{{ RenderLocationsStateToPango "state.Location" "location" }}

	if err := location.IsValid(); err != nil {
		resp.Diagnostics.AddError("Invalid location", err.Error())
		return
	}

	// Load the desired config.
	var obj *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
	resp.Diagnostics.Append(state.CopyToPango(ctx, o.client, nil, &obj, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(state.Timeouts.Create))
		defer cancel()
	*/

	// Perform the operation.

	components, err := state.resourceXpathParentComponents()
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
		return
	}

{{- if .HasImports }}
	var importLocation {{ .resourceSDKName }}.ImportLocation
	{{ RenderImportLocationAssignment "state.Location" "importLocation" }}
	created, err := o.manager.Create(ctx, location, components, obj)
	if err != nil {
		resp.Diagnostics.AddError("Error in create", err.Error())
		return
	}

	if importLocation != nil {
		err = o.manager.ImportToLocations(ctx, location, []{{ .resourceSDKName }}.ImportLocation{importLocation}, obj.Name)
		if err != nil {
			resp.Diagnostics.AddError("Failed to import resource into location", err.Error())
			return
		}
	}
{{- else }}
	created, err := o.manager.Create(ctx, location, components, obj)
	if err != nil {
		resp.Diagnostics.AddError("Error in create", err.Error())
		return
	}
{{- end }}


	resp.Diagnostics.Append(state.CopyFromPango(ctx, o.client, nil, created, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}

	{{ RenderEncryptedValuesFinalizer }}

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

{{- /* Done */ -}}
`

const resourceReadEntryListFunction = `
{{- $structName := "" }}
{{- if eq .ResourceOrDS "DataSource" }}
  {{ $structName = .dataSourceStructName }}
{{- else }}
  {{ $structName = .resourceStructName }}
{{- end }}
{{- $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{- $resourceTFStructName := printf "%s%sObject" $structName .ListAttribute.CamelCase }}

{{- $stateName := "" }}
{{- if eq .ResourceOrDS "DataSource" }}
  {{- $stateName = "Config" }}
{{- else }}
  {{- $stateName = "State" }}
{{- end -}}



var state {{ .structName }}{{ .ResourceOrDS }}Model

resp.Diagnostics.Append(req.{{ $stateName }}.Get(ctx, &state)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource create", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Create",
})

{{ RenderEncryptedValuesInitialization }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

{{- if eq .PluralType "map" }}
elements := make(map[string]{{ $resourceTFStructName }})
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if len(elements) == 0 || resp.Diagnostics.HasError() {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, 0, len(elements))
for name, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	entries = append(entries, entry)
}

{{- else if or (eq .PluralType "list") (eq .PluralType "set") }}
var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, 0, len(elements))
for _, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries = append(entries, entry)
}
{{- end }}

components, err := state.resourceXpathParentComponents()
if err != nil {
	resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
	return
}

readEntries, err := o.manager.ReadMany(ctx, location, components)
if err != nil {
	if errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.State.RemoveResource(ctx)
	} else {
		resp.Diagnostics.AddError("Failed to read entries from the server", err.Error())
	}
	return
}

{{- if eq .PluralType "map" }}
entriesByName := make(map[string]*{{ $resourceSDKStructName }})
for _, elt := range entries {
	entriesByName[elt.EntryName()] = elt
}

objects := make(map[string]{{ $resourceTFStructName }})
for _, elt := range readEntries {
	if _, found := entriesByName[elt.EntryName()]; !found {
		continue
	}

	var object {{ $resourceTFStructName }}
	object.name = elt.Name
	resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, elt, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[elt.Name] = object
}
{{- else if or (eq .PluralType "list") (eq .PluralType "set") }}
objects := make([]{{ $resourceTFStructName }}, len(readEntries))
for idx, elt := range readEntries {
	var object {{ $resourceTFStructName }}
	resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, elt, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[idx] = object
}
{{- end }}

{{ if eq .PluralType "map" }}
var map_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, map_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(map_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- else if eq .PluralType "list" }}
var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- else if eq .PluralType "set" }}
var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.SetValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

{{- if .ResourceXpathVariablesWithChecks }}
{{ AttributesFromXpathComponents "state" }}
{{- end }}

{{ RenderEncryptedValuesFinalizer }}

resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceReadManyFunction = `
{{- $structName := "" }}
{{- if eq .ResourceOrDS "DataSource" }}
  {{ $structName = .dataSourceStructName }}
{{- else }}
  {{ $structName = .resourceStructName }}
{{- end }}
{{- $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{- $resourceTFStructName := printf "%s%sObject" $structName .ListAttribute.CamelCase }}

{{- $stateName := "" }}
{{- if eq .ResourceOrDS "DataSource" }}
  {{- $stateName = "Config" }}
{{- else }}
  {{- $stateName = "State" }}
{{- end -}}

var state {{ .structName }}{{ .ResourceOrDS }}Model

resp.Diagnostics.Append(req.{{ $stateName }}.Get(ctx, &state)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource create", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Create",
})

{{ RenderEncryptedValuesInitialization }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() || len(elements) == 0 {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, 0, len(elements))
for _, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries = append(entries, entry)
}

{{ $exhaustive := "sdkmanager.NonExhaustive" }}
{{- if .Exhaustive }}
  {{ $exhaustive = "sdkmanager.Exhaustive" }}
position := movement.PositionFirst{}
{{- else }}
var position movement.Position
var positionAttribute TerraformPositionObject
if !state.Position.IsNull() && !state.Position.IsUnknown() {
	resp.Diagnostics.Append(state.Position.As(ctx, &positionAttribute, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	position = positionAttribute.CopyToPango()
}
{{- end }}

{{- if .Exhaustive }}
readEntries, _, err := o.manager.ReadMany(ctx, location, entries, {{ $exhaustive }}, position)
{{- else }}
readEntries, movementRequired, err := o.manager.ReadMany(ctx, location, entries, {{ $exhaustive }}, position)
{{- end }}
if err != nil {
	if errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.State.RemoveResource(ctx)
	} else {
		resp.Diagnostics.AddError("Failed to read entries from the server", err.Error())
	}
	return
}

var objects []{{ $resourceTFStructName }}
for _, elt := range readEntries {
	var object {{ $resourceTFStructName }}
	err := object.CopyFromPango(ctx, o.client, nil, elt, ev)
	resp.Diagnostics.Append(err...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects = append(objects, object)
}

{{- if not .Exhaustive }}
if movementRequired {
	state.Position = types.ObjectNull(positionAttribute.AttributeTypes())
}
{{- end }}

var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}

{{ RenderEncryptedValuesFinalizer }}

resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceReadFunction = `
{{- if eq .ResourceOrDS "DataSource" }}
	var state {{ .dataSourceStructName }}Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
{{- else }}
	var state {{ .resourceStructName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
{{- end }}
	if resp.Diagnostics.HasError() {
		return
	}

	{{ RenderEncryptedValuesInitialization }}

	var location {{ .resourceSDKName }}.Location
	{{ RenderLocationsStateToPango "state.Location" "location" }}

	// Basic logging.
	tflog.Info(ctx, "performing resource read", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .resourceStructName }}",
		"function":      "Read",
{{- if .HasEntryName }}
		"name":          state.Name.ValueString(),
{{- end }}
	})

	components, err := state.resourceXpathParentComponents()
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
		return
	}

{{- if .HasEntryName }}
	object, err := o.manager.Read(ctx, location, components, state.Name.ValueString())
{{- else }}
	object, err := o.manager.Read(ctx, location, components)
{{- end }}
	if err != nil {
		if errors.Is(err, sdkmanager.ErrObjectNotFound) {
{{- if eq .ResourceOrDS "DataSource" }}
			resp.Diagnostics.AddError("Error reading data", err.Error())
{{- else }}
			resp.State.RemoveResource(ctx)
{{- end }}
		} else {
			resp.Diagnostics.AddError("Error reading entry", err.Error())
		}
		return
	}

	copy_diags := state.CopyFromPango(ctx, o.client, nil, object, ev)
	resp.Diagnostics.Append(copy_diags...)

	/*
			// Keep the timeouts.
		    // TODO: This won't work for state import.
			state.Timeouts = state.Timeouts
	*/

	state.Location = state.Location

{{ AttributesFromXpathComponents "state" }}

{{ RenderEncryptedValuesFinalizer }}

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
`

const resourceUpdateEntryListFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{ $resourceTFStructName := printf "%s%sObject" .structName .ListAttribute.CamelCase }}

var state, plan {{ .structName }}Model
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource create", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Create",
})

{{ RenderEncryptedValuesInitialization }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "plan.Location" "location" }}

// Basic logging.
tflog.Info(ctx, "performing resource update", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Update",
})

{{ if eq .PluralType "map" }}
var elements map[string]{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

stateEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx := 0
for name, elt := range elements {
	var entry *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	stateEntries[idx] = entry
	idx++
}
{{ else if or (eq .PluralType "list") (eq .PluralType "set") }}
var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

stateEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		 return
	}
	stateEntries[idx] = entry
}
{{- end }}

stateEntriesByName := make(map[string]*{{ $resourceSDKStructName }}, len(stateEntries))
for _, elt := range stateEntries {
	stateEntriesByName[elt.Name] = elt
}

components, err := state.resourceXpathParentComponents()
if err != nil {
	resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
	return
}

existing, err := o.manager.ReadMany(ctx, location, components)
if err != nil && !errors.Is(err, sdkmanager.ErrObjectNotFound) {
	resp.Diagnostics.AddError("Error while reading entries from the server", err.Error())
	return
}

filtered := make([]*{{ $resourceSDKStructName }}, 0, len(stateEntries))
for _, elt := range existing {
	if _, found := stateEntriesByName[elt.EntryName()]; found {
		filtered = append(filtered, elt)
	}
}

existingEntriesByName := make(map[string]*{{ $resourceSDKStructName }}, len(filtered))
for _, elt := range filtered {
	existingEntriesByName[elt.Name] = elt
}

resp.Diagnostics.Append(plan.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

{{ if eq .PluralType "map" }}
planEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx = 0
for name, elt := range elements {
	entry, _ := existingEntriesByName[name]
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry.Name = name
	planEntries[idx] = entry
	idx++
}
{{ else if or (eq .PluralType "list") (eq .PluralType "set") }}
var planEntries []*{{ $resourceSDKStructName }}
for _, elt := range elements {
	existingEntry, _ := existingEntriesByName[elt.Name.ValueString()]
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &existingEntry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planEntries = append(planEntries, existingEntry)
}
{{- end }}

processed, err := o.manager.UpdateMany(ctx, location, components, stateEntries, planEntries)
if err != nil {
	resp.Diagnostics.AddError("Error while updating entries", err.Error())
	return
}

{{- if eq .PluralType "map" }}
objects := make(map[string]*{{ $resourceTFStructName }}, len(processed))
for _, elt := range processed {
	var object {{ $resourceTFStructName }}
	object.name = elt.Name
	copy_diags := object.CopyFromPango(ctx, o.client, nil, elt, ev)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[elt.Name] = &object
}
{{- else if or (eq .PluralType "list") (eq .PluralType "set") }}
objects := make([]*{{ $resourceTFStructName }}, len(processed))
for idx, elt := range processed {
	var object {{ $resourceTFStructName }}
	resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, elt, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects[idx] = &object
}
{{- end }}

var list_diags diag.Diagnostics
{{ if eq .PluralType "map" }}
plan.{{ .ListAttribute.CamelCase }}, list_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
{{ else if eq .PluralType "list" }}
plan.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
{{ else if eq .PluralType "set" }}
plan.{{ .ListAttribute.CamelCase }}, list_diags = types.SetValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
{{- end }}
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}

{{ RenderEncryptedValuesFinalizer }}

resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
`

const resourceUpdateManyFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{ $resourceTFStructName := printf "%s%sObject" .structName .ListAttribute.CamelCase }}

var state, plan {{ .structName }}Model
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource create", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Create",
})

{{ RenderEncryptedValuesInitialization }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "plan.Location" "location" }}

var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}
stateEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		 return
	}
	stateEntries[idx] = entry
}

{{ $exhaustive := "sdkmanager.NonExhaustive" }}
{{- if .Exhaustive }}
  {{ $exhaustive = "sdkmanager.Exhaustive" }}
position := movement.PositionFirst{}
{{- else }}
var positionAttribute TerraformPositionObject
resp.Diagnostics.Append(plan.Position.As(ctx, &positionAttribute, basetypes.ObjectAsOptions{})...)
if resp.Diagnostics.HasError() {
	return
}
position := positionAttribute.CopyToPango()
{{- end }}

existing, _, err := o.manager.ReadMany(ctx, location, stateEntries, {{ $exhaustive }}, position)
if err != nil && !errors.Is(err, sdkmanager.ErrObjectNotFound) {
	resp.Diagnostics.AddError("Error while reading entries from the server", err.Error())
	return
}

existingEntriesByName := make(map[string]*{{ $resourceSDKStructName }}, len(existing))
for _, elt := range existing {
	existingEntriesByName[elt.Name] = elt
}

resp.Diagnostics.Append(plan.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

planEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	entry, _ := existingEntriesByName[elt.Name.ValueString()]
	resp.Diagnostics.Append(elt.CopyToPango(ctx, o.client, nil, &entry, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	planEntries[idx] = entry
}

components, err := state.resourceXpathParentComponents()
if err != nil {
	resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
	return
}

processed, err := o.manager.UpdateMany(ctx, location, components, stateEntries, planEntries, {{ $exhaustive }}, position)
if err != nil {
	resp.Diagnostics.AddError("Failed to udpate entries", err.Error())
}

objects := make([]*{{ $resourceTFStructName }}, len(processed))
for idx, elt := range processed {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, o.client, nil, elt, ev)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[idx] = &object
}

var list_diags diag.Diagnostics
plan.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}

{{ RenderEncryptedValuesFinalizer }}

resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
`

const resourceUpdateFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}

	var plan, state {{ .structName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	{{ RenderEncryptedValuesInitialization }}

	var location {{ .resourceSDKName }}.Location
	{{ RenderLocationsStateToPango "state.Location" "location" }}

	// Basic logging.
	tflog.Info(ctx, "performing resource update", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .structName }}",
		"function":      "Update",
	})

	// Verify mode.
	if o.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	components, err := state.resourceXpathParentComponents()
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
		return
	}

{{- if .HasEntryName }}
	var obj *{{ .resourceSDKName }}.Entry
	if state.Name.ValueString() != plan.Name.ValueString() {
		obj, err = o.manager.Read(ctx, location, components, state.Name.ValueString())
	} else {
		obj, err = o.manager.Read(ctx, location, components, plan.Name.ValueString())
	}
{{- else }}
	obj, err := o.manager.Read(ctx, location, components)
{{- end }}
	if err != nil {
		resp.Diagnostics.AddError("Error in update", err.Error())
		return
	}

	resp.Diagnostics.Append(plan.CopyToPango(ctx, o.client, nil, &obj, ev)...)
	if resp.Diagnostics.HasError() {
		return
	}

	components, err = plan.resourceXpathParentComponents()
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
		return
	}

{{ if .HasEntryName }}
	// If name differs between plan and state, we need to set old name for the object
	// before calling SDK Update() function to properly handle rename + edit cycle.
	var newName string
	if state.Name.ValueString() != plan.Name.ValueString() {
		newName = plan.Name.ValueString()
		obj.Name = state.Name.ValueString()
	}

	updated, err := o.manager.Update(ctx, location, components, obj, newName)
{{ else }}
	updated, err := o.manager.Update(ctx, location, components, obj)
{{ end }}
	if err != nil {
		resp.Diagnostics.AddError("Error in update", err.Error())
		return
	}

	/*
		// Keep the timeouts.
		state.Timeouts = plan.Timeouts
	*/

	copy_diags := plan.CopyFromPango(ctx, o.client, nil, updated, ev)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	{{ RenderEncryptedValuesFinalizer }}

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
`

const resourceDeleteManyFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{ $resourceTFStructName := printf "%s%sObject" .structName .ListAttribute.CamelCase }}

var state {{ .structName }}Model
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
if resp.Diagnostics.HasError() {
	return
}

// Basic logging.
tflog.Info(ctx, "performing resource delete", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Delete",
})

{{- if eq .PluralType "map" }}
elements := make(map[string]{{ $resourceTFStructName }}, len(state.{{ .ListAttribute.CamelCase }}.Elements()))
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}
{{- else if or (eq .PluralType "list") (eq .PluralType "set") }}
var elements []{{ $resourceTFStructName }}
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

var names []string
{{- if eq .PluralType "map" }}
for name, _ := range elements {
	names = append(names, name)
}
{{- else if or (eq .PluralType "list") (eq .PluralType "set") }}
for _, elt := range elements {
	names = append(names, elt.Name.ValueString())
}
{{- end }}

components, err := state.resourceXpathParentComponents()
if err != nil {
	resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
	return
}

{{- if eq .Exhaustive "exhaustive" }}
err = o.manager.Delete(ctx, location, components, names, sdkmanager.Exhaustive)
{{- else if eq .Exhaustive "non-exhaustive" }}
err = o.manager.Delete(ctx, location, components, names, sdkmanager.NonExhaustive)
{{- else }}
err = o.manager.Delete(ctx, location, components, names)
{{- end }}
if err != nil {
	resp.Diagnostics.AddError("error while deleting entries", err.Error())
	return
}
`

const resourceDeleteFunction = `
{{ $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}

	var state {{ .structName }}Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Basic logging.
	tflog.Info(ctx, "performing resource delete", map[string]any{
		"resource_name": "panos_{{ UnderscoreName .structName }}",
		"function":      "Delete",
{{- if .HasEntryName }}
		"name":          state.Name.ValueString(),
{{- end }}
	})

	// Verify mode.
	if o.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	var location {{ .resourceSDKName }}.Location
	{{ RenderLocationsStateToPango "state.Location" "location" }}

{{- if .HasEntryName }}
	components, err := state.resourceXpathParentComponents()
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
		return
	}
  {{- if .HasImports }}
	var importLocation {{ .resourceSDKName }}.ImportLocation
	{{ RenderImportLocationAssignment "state.Location" "importLocation" }}
	if importLocation != nil {
		err = o.manager.UnimportFromLocations(ctx, location, []{{ .resourceSDKName }}.ImportLocation{importLocation}, state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Error in delete", err.Error())
		return
	}
	err = o.manager.Delete(ctx, location, []{{ .resourceSDKName }}.ImportLocation{importLocation}, components, []string{state.Name.ValueString()})
  {{- else }}
	err = o.manager.Delete(ctx, location, components, []string{state.Name.ValueString()})
  {{- end }}
	if err != nil && !errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
		return
	}
{{- else }}
	components, err := state.resourceXpathParentComponents()
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource xpath", err.Error())
		return
	}

	existing, err := o.manager.Read(ctx, location, components)
	if err != nil {
		resp.Diagnostics.AddError("Error while deleting resource", err.Error())
		return
	}

	var obj {{ $resourceSDKStructName }}
	obj.Misc = existing.Misc

	err = o.manager.Delete(ctx, location, &obj)
	if err != nil && !errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
		return
	}
{{- end }}
`

const renderImportStateStructsTmpl = `
{{- range .Specs }}
type {{ .StructName }} struct {
  {{- range .Fields }}
	{{ .Name }} {{ .Type }} {{ .Tags }}
  {{- end }}
}
{{- end }}
`

const locationMarshallersTmpl = `
{{- define "renderMarshallerField" }}
  {{- if eq .Type "types.Object" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_object,
  {{- else if eq .Type "types.List" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_list,
  {{- else }}
	{{ .Name.CamelCase }}: o.{{ .Name.CamelCase }}.Value{{ .Type | CamelCaseName }}Pointer(),
  {{- end }}
{{- end }}

{{- define "renderShadowStructField" }}
  {{- if eq .Type "types.Object" }}
    {{ .Name.CamelCase }} *{{ .StructName }} {{ .Tags }}
  {{- else if eq .Type "types.List" }}
    {{ .Name.CamelCase }} {{ .StructName }} {{ .Tags }}
  {{- else }}
    {{ .Name.CamelCase }} *{{ .Type }} {{ .Tags }}
  {{- end }}
{{- end }}

{{- define "renderUnmarshallerField" }}
  {{- if eq .Type "types.Object" }}
    o.{{ .Name.CamelCase }} = {{ .Name.LowerCamelCase }}_object
  {{- else if eq .Type "types.List" }}
    o.{{ .Name.CamelCase }} = {{ .Name.LowerCamelCase }}_list
  {{- else }}
    o.{{ .Name.CamelCase }} = types.{{ .Type | CamelCaseName }}PointerValue(shadow.{{ .Name.CamelCase }})
  {{- end }}
{{- end }}

{{- define "renderPangoObjectConversion" -}}
  {{- if eq .Type "types.Object" }}
	var {{ .Name.LowerCamelCase }}_object types.Object
	{
		var diags_tmp diag.Diagnostics
		{{ .Name.LowerCamelCase }}_object, diags_tmp = types.ObjectValueFrom(context.TODO(), shadow.{{ .Name.CamelCase }}.AttributeTypes(), shadow.{{ .Name.CamelCase }})
		if diags_tmp.HasError() {
			return NewDiagnosticsError("Failed to unmarshal JSON document into {{ .Name.Underscore }}", diags_tmp.Errors())
		}
	}
  {{- else if eq .Type "types.List" }}
	var {{ .Name.LowerCamelCase }}_list types.List
	{
		var diags_tmp diag.Diagnostics
		{{ .Name.LowerCamelCase }}_list, diags_tmp = types.ListValueFrom(context.TODO(), types.StringType, shadow.{{ .Name.CamelCase }})
		if diags_tmp.HasError() {
			return NewDiagnosticsError("Failed to unmarshal JSON document into {{ .Name.Underscore }}", diags_tmp.Errors())
		}
	}
  {{- end }}
{{- end }}

{{- define "renderTfObjectConversion" -}}
  {{- if eq .Type "types.Object" }}
	var {{ .Name.LowerCamelCase }}_object *{{ .StructName }}
        {
		diags := o.{{ .Name.CamelCase }}.As(context.TODO(), &{{ .Name.LowerCamelCase }}_object, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, NewDiagnosticsError("Failed to marshal {{ .Name.Underscore }} into JSON document", diags.Errors())
		}
        }
  {{- else if eq .Type "types.List" }}
	var {{ .Name.LowerCamelCase }}_list {{ .StructName }}
	{
		diags := o.{{ .Name.CamelCase }}.ElementsAs(context.TODO(), &{{ .Name.LowerCamelCase }}_list, false)
		if diags.HasError() {
			return nil, NewDiagnosticsError("Failed to marshal {{ .Name.Underscore }} into JSON document", diags.Errors())
		}
	}
  {{- end }}
{{- end }}

{{- range .Specs }}
  {{- $spec := . }}
func (o {{ .StructName }}) MarshalJSON() ([]byte, error) {
	type shadow struct {
  {{- range .Fields }}
    {{- template "renderShadowStructField" . }}
  {{- end }}
	}

  {{-  range .Fields }}
    {{- template "renderTfObjectConversion" . }}
  {{- end }}

	obj := shadow{
  {{- range .Fields }}
    {{- template "renderMarshallerField" . }}
  {{- end }}
	}

	return json.Marshal(obj)
}

func (o *{{ .StructName }}) UnmarshalJSON(data []byte) error {
	var shadow struct {
  {{- range .Fields }}
    {{- template "renderShadowStructField" . }}
  {{- end }}
	}

	err := json.Unmarshal(data, &shadow)
	if err != nil {
		return err
	}
  {{- range .Fields }}
    {{- template "renderPangoObjectConversion" . }}
  {{- end }}

  {{- range .Fields }}
    {{- template "renderUnmarshallerField" . }}
  {{- end }}

	return nil
}
{{- end }}
`

const resourceImportStateCreatorTmpl = `
func {{ .FuncName }}(ctx context.Context, resource types.Object) ([]byte, error) {
	attrs := resource.Attributes()
	if attrs == nil {
		return nil, fmt.Errorf("Object has no attributes")
	}

	locationAttr, ok := attrs["location"]
	if !ok {
		return nil, fmt.Errorf("location attribute missing")
	}

	var location types.Object
	switch value := locationAttr.(type) {
	case types.Object:
		location = value
	default:
		return nil, fmt.Errorf("location attribute expected to be an object")
	}

{{- if eq .ResourceType "entry" }}
	nameAttr, ok := attrs["name"]
	if !ok {
		return nil, fmt.Errorf("name attribute missing")
	}

	var name types.String
	switch value := nameAttr.(type) {
	case types.String:
		name = value
	default:
		return nil, fmt.Errorf("name attribute expected to be a string")
	}

	importStruct := {{ .StructNamePrefix }}ImportState{
		Location: location,
		Name: name,
	}
{{- else if eq .ResourceType "entry-plural" }}
  {{- if .HasParent }}
	parentAttr, ok := attrs["{{ .ParentAttribute.Underscore }}"]
	if !ok {
		return nil, fmt.Errorf("{{ .ParentAttribute.Underscore }} attribute missing")
	}

	var parent types.String
	switch value := parentAttr.(type) {
	case types.String:
		parent = value
	default:
		return nil, fmt.Errorf("{{ .ParentAttribute.Underscore }} expected to be a map")
	}

	importStruct := {{ .StructNamePrefix }}ImportState{
		Location: location,
		{{ .ParentAttribute.CamelCase }}: parent,
	}
  {{- else }}
	itemsAttr, ok := attrs["{{ .ListAttribute.Underscore }}"]
	if !ok {
		return nil, fmt.Errorf("{{ .ListAttribute.Underscore }} attribute missing")
	}

	items := make(map[string]{{ .ListStructName }})
	switch value := itemsAttr.(type) {
	case types.Map:
		diags := value.ElementsAs(ctx, &items, false)
		if diags.HasError() {
			return nil, fmt.Errorf("Failed to convert {{ .ListAttribute.Underscore }} into a valid map: %s", diags.Errors())
		}
	default:
		return nil, fmt.Errorf("{{ .ListAttribute.Underscore }} expected to be a map")
	}

	var names []string
	for key := range items {
		names = append(names, key)
	}

	var namesObj types.List
	{
		var diags_err diag.Diagnostics
		namesObj, diags_err = types.ListValueFrom(ctx, types.StringType, names)
		if diags_err.HasError() {
			return nil, NewDiagnosticsError("Failed to generate a list of names for the import ID", diags_err.Errors())
		}
	}

	importStruct := {{ .StructNamePrefix }}ImportState{
		Location: location,
		Names: namesObj,
	}
  {{- end }}
{{- else if or (eq .ResourceType "uuid") }}
	itemsAttr, ok := attrs["{{ .ListAttribute.Underscore }}"]
	if !ok {
		return nil, fmt.Errorf("{{ .ListAttribute.Underscore }} attribute missing")
	}

	var items []*{{ .ListStructName }}
	switch value := itemsAttr.(type) {
	case types.List:
		diags := value.ElementsAs(ctx, &items, false)
		if diags.HasError() {
			return nil, fmt.Errorf("Invalid {{ .ListAttribute.Underscore }} attribute element type, expected list of valid objects")
		}
	default:
		return nil, fmt.Errorf("Invalid names attribute type, expected list of strings")
	}

	var names []string
	for _, elt := range items {
		names = append(names, elt.Name.ValueString())
	}

	var namesObject types.List
	namesObject, diags_tmp := types.ListValueFrom(ctx, types.StringType, names)
	if diags_tmp.HasError() {
		return nil, NewDiagnosticsError("Failed to generate import ID", diags_tmp.Errors())
	}

	importStruct := {{ .StructNamePrefix }}ImportState{
		Location: location,
		Names: namesObject,
	}
{{- else if (eq .ResourceType "uuid-plural") }}
	positionAttr, ok := attrs["position"]
	if !ok {
		return nil, fmt.Errorf("position attribute missing")
	}

	var position types.Object
	switch value := positionAttr.(type) {
	case types.Object:
		position = value
	default:
		return nil, fmt.Errorf("position attribute expected to be an object")
	}

	itemsAttr, ok := attrs["{{ .ListAttribute.Underscore }}"]
	if !ok {
		return nil, fmt.Errorf("{{ .ListAttribute.Underscore }} attribute missing")
	}

	var items []*{{ .ListStructName }}
	switch value := itemsAttr.(type) {
	case types.List:
		diags := value.ElementsAs(ctx, &items, false)
		if diags.HasError() {
			return nil, fmt.Errorf("Invalid {{ .ListAttribute.Underscore }} attribute element type, expected list of valid objects")
		}
	default:
		return nil, fmt.Errorf("Invalid names attribute type, expected list of strings")
	}

	var names []string
	for _, elt := range items {
		names = append(names, elt.Name.ValueString())
	}

	var namesObject types.List
	namesObject, diags_tmp := types.ListValueFrom(ctx, types.StringType, names)
	if diags_tmp.HasError() {
		return nil, NewDiagnosticsError("Failed to generate import ID", diags_tmp.Errors())
	}

	importStruct := {{ .StructNamePrefix }}ImportState{
		Location: location,
		Position: position,
		Names: namesObject,
	}
{{- end }}

	return json.Marshal(importStruct)
}
`

const resourceImportStateFunctionTmpl = `
	var obj {{ .StructName }}ImportState
	data, err := base64.StdEncoding.DecodeString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode Import ID", err.Error())
		return
	}

{{- if and .PluralType (not .HasParent) }}
	{{ RenderEncryptedValuesInitialization }}
{{- end }}

	err = json.Unmarshal(data, &obj)
	if err != nil {
		var diagsErr *DiagnosticsError
		if errors.As(err, &diagsErr) {
			resp.Diagnostics.Append(diagsErr.Diagnostics()...)
		} else {
			resp.Diagnostics.AddError("Failed to unmarshal Import ID", err.Error())
		}
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("location"), obj.Location)...)
	if resp.Diagnostics.HasError() {
		return
	}
{{- if .HasParent }}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("{{ .ParentAttribute.Underscore }}"), obj.{{ .ParentAttribute.CamelCase }})...)
  {{- if eq .PluralType "" }}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), obj.Name)...)
  {{- end }}
{{- else if eq .PluralType "map" }}
	names := make(map[string]*{{ .ListStructName }})

	var objectNames []string
	resp.Diagnostics.Append(obj.Names.ElementsAs(ctx, &objectNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, elt := range objectNames {
		object := &{{ .ListStructName }}{}
		resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, &{{ .PangoStructName }}{}, ev)...)
		if resp.Diagnostics.HasError() {
			return
		}
		names[elt] = object
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("{{ .ListAttribute.Underscore }}"), names)...)
{{- else if eq .PluralType "list" }}
  {{- if .HasPosition }}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("position"), obj.Position)...)
	if resp.Diagnostics.HasError() {
		return
	}
  {{- end }}

	var names []*{{ .ListStructName }}
	var objectNames []string
	resp.Diagnostics.Append(obj.Names.ElementsAs(ctx, &objectNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, elt := range objectNames {
		object := &{{ .ListStructName }}{}
		resp.Diagnostics.Append(object.CopyFromPango(ctx, o.client, nil, &{{ .PangoStructName }}{}, ev)...)
		if resp.Diagnostics.HasError() {
			return
		}
		object.Name = types.StringValue(elt)
		names = append(names, object)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("{{ .ListAttribute.Underscore }}"), names)...)
{{- else }}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), obj.Name)...)
{{- end -}}
`

const commonTemplate = `
{{- if HasLocations }}
{{- RenderLocationStructs }}

{{- RenderLocationSchemaGetter }}

{{- RenderLocationMarshallers }}

{{- RenderLocationAttributeTypes }}
{{- end }}
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

{{- if IsCustom }}
	custom *{{ structName }}Custom
{{- else if and IsEntry HasImports }}
	manager *sdkmanager.ImportableEntryObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, {{ resourceSDKName }}.ImportLocation, *{{ resourceSDKName }}.Service]
{{- else if IsEntry }}
	manager *sdkmanager.EntryObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service]
{{- else if IsUuid }}
	manager *sdkmanager.UuidObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service]
{{- else if IsConfig }}
	manager *sdkmanager.ConfigObjectManager[*{{ resourceSDKName }}.Config, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service]
{{- end }}
}

type {{ dataSourceStructName }}Filter struct {
//TODO: Generate Data Source filter via function
}

{{ RenderDataSourceStructs }}

{{ RenderModelAttributeTypesFunction }}

{{ RenderCopyToPangoFunctions }}

{{ RenderCopyFromPangoFunctions }}

{{ RenderXpathComponentsGetter }}

{{ RenderDataSourceSchema }}

{{- if HasLocations }}
func {{ dataSourceStructName }}LocationSchema() rsschema.Attribute {
	return {{ structName }}LocationSchema()
}
{{- end }}

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

	providerData := req.ProviderData.(*ProviderData)
	d.client = providerData.Client

{{- if IsCustom }}
	custom, err := New{{ structName }}Custom(providerData)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	d.custom = custom
{{- else if and IsEntry HasImports }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	batchSize := providerData.MultiConfigBatchSize
	d.manager =  sdkmanager.NewImportableEntryObjectManager(d.client, {{ resourceSDKName }}.NewService(d.client), batchSize, specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsEntry }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	batchSize := providerData.MultiConfigBatchSize
	d.manager =  sdkmanager.NewEntryObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service](d.client, {{ resourceSDKName }}.NewService(d.client), batchSize, specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsUuid }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	batchSize := providerData.MultiConfigBatchSize
	d.manager =  sdkmanager.NewUuidObjectManager[*{{ resourceSDKName }}.Entry, {{ resourceSDKName }}.Location, *{{ resourceSDKName }}.Service](d.client, {{ resourceSDKName }}.NewService(d.client), batchSize, specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsConfig }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	d.manager =  sdkmanager.NewConfigObjectManager(d.client, {{ resourceSDKName }}.NewService(d.client), specifier)
{{- end }}
}

{{- if FunctionSupported "Read" }}
func (o *{{ dataSourceStructName }}) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	{{ DataSourceReadFunction dataSourceStructName serviceName }}
}
{{- end }}

{{- /* Done */ -}}
`

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
	_ provider.ProviderWithFunctions = &PanosProvider{}
	_ provider.ProviderWithActions = &PanosProvider{}
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

type ProviderData struct {
	Client               *sdk.Client
	MultiConfigBatchSize int
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
		var logCategories sdk.LogCategory
		if !config.SdkLogCategories.IsNull() {
			categories := strings.Split(config.SdkLogCategories.ValueString(), ",")
			var err error
			logCategories, err = sdk.LogCategoryFromStrings(categories)
			if err != nil {
				resp.Diagnostics.AddError("Failed to configure Terraform provider", err.Error())
				return
			}
		}

		var logLevel slog.Level
		if !config.SdkLogLevel.IsNull() {
			levelStr := config.SdkLogLevel.ValueString()
			err := logLevel.UnmarshalText([]byte(levelStr))
			if err != nil {
				resp.Diagnostics.AddError("Failed to configure Terraform provider", fmt.Sprintf("Invalid Log Level: %s", levelStr))
			}
		} else {
			logLevel = slog.LevelInfo
		}

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
			Logging: sdk.LoggingInfo{
				LogLevel: logLevel,
				LogCategories: logCategories,
			},
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

	batchSize := config.MultiConfigBatchSize.ValueInt64()
	if batchSize == 0 {
		batchSize = 500
	} else if batchSize < 0 || batchSize > 10000 {
		resp.Diagnostics.AddError("Failed to configure Terraform provider", fmt.Sprintf("multi_config_batch_size must be between 1 and 10000, value: %d", batchSize))
		return
	}

	providerData := &ProviderData{
		Client: con,
		MultiConfigBatchSize: int(batchSize),
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
	resp.EphemeralResourceData = providerData

	// Done.
	tflog.Info(ctx, "Configured client", map[string]any{"success": true})
}

// DataSources defines the data sources for this provider.
func (p *PanosProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
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

func (p *PanosProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
{{- range $fnName := EphemeralResources }}
	New{{ $fnName }},
{{- end }}
	}
}

func (p *PanosProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
{{- range $fnName := Actions }}
	New{{ $fnName }},
{{- end }}
	}
}

func (p *PanosProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewAddressValueFunction,
		NewCreateImportIdFunction,
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

type CreateResourceIdFunc func(context.Context, types.Object) ([]byte, error)

type resourceFuncs struct {
	CreateImportId CreateResourceIdFunc
}

var resourceFuncMap = map[string]resourceFuncs{
{{- RenderResourceFuncMap }}
}

{{- /* Done */ -}}`

const resourceFuncMapTmpl = `
{{- range .Entries }}
	"{{ .Key }}": resourceFuncs{
		CreateImportId: {{ .StructName }}ImportStateCreator,
	},
{{- end }}
`
