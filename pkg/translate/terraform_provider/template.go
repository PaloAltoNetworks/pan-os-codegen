package terraform_provider

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
{{- if IsCustom }}

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

{{- if not GoSDKSkipped }}
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
{{- end }}

func {{ resourceStructName }}LocationSchema() rsschema.Attribute {
	return {{ structName }}LocationSchema()
}

{{ RenderResourceStructs }}

func (r *{{ resourceStructName }}) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "{{ metaName }}"
}

func (r *{{ resourceStructName }}) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
{{- if HasPosition }}
	var resource {{ resourceStructName }}Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &resource)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource.Position.ValidateConfig(resp)
{{- end }}
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

{{- if IsCustom }}

{{- else if and IsEntry HasImports }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(r.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	r.manager =  sdkmanager.NewImportableEntryObjectManager(r.client, {{ resourceSDKName }}.NewService(r.client), specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsEntry }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(r.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	r.manager =  sdkmanager.NewEntryObjectManager(r.client, {{ resourceSDKName }}.NewService(r.client), specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsUuid }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(r.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	r.manager =  sdkmanager.NewUuidObjectManager(r.client, {{ resourceSDKName }}.NewService(r.client), specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsConfig }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(r.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	r.manager =  sdkmanager.NewConfigObjectManager(r.client, {{ resourceSDKName }}.NewService(r.client), specifier)
{{- end }}
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

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

{{ $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
{{- end }}


type entryWithState struct {
	Entry    *{{ $resourceSDKStructName }}
	StateIdx int
}

var elements map[string]{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
entries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx := 0
for name, elt := range elements {
	var entry *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	entries[idx] = entry
	idx++
}

created, err := r.manager.CreateMany(ctx, location, entries)
if err != nil {
	resp.Diagnostics.AddError("Failed to create new entries", err.Error())
	return
}

for _, elt := range created {
	if _, found := elements[elt.Name]; !found {
		continue
	}
	var object {{ $resourceTFStructName }}
	resp.Diagnostics.Append(object.CopyFromPango(ctx, elt, {{ $ev }})...)
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

{{- if .HasEncryptedResources }}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
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

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

{{ $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
{{- end }}


var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
entries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries[idx] = entry
}

{{- if .ResourceIsMap }}
processed, err := o.manager.CreateMany(ctx, location, entries)
if err != nil {
	resp.Diagnostics.AddError("Error during CreateMany() call", err.Error())
	return
}
{{- else if .Exhaustive }}
trueVal := true
processed, err := r.manager.CreateMany(ctx, location, entries, sdkmanager.Exhaustive, rule.Position{First: &trueVal})
if err != nil {
	resp.Diagnostics.AddError("Error during CreateMany() call", err.Error())
	return
}
{{- else }}
position := state.Position.CopyToPango()
processed, err := r.manager.CreateMany(ctx, location, entries, sdkmanager.NonExhaustive, position)
if err != nil {
	resp.Diagnostics.AddError("Error during CreateMany() call", err.Error())
	return
}
{{- end }}


{{- if .ResourceIsMap }}
objects := make(map[string]{{ $resourceTFStructName }}, len(processed))
for _, elt := range processed {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, elt, {{ $ev }})
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[elt.Name] = object
}

var map_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, map_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- else }}
objects := make([]{{ $resourceTFStructName }}, len(processed))
for idx, elt := range processed {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, elt, {{ $ev }})
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
{{- end }}

{{- if .HasEncryptedResources }}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
{{- end }}

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
	if r.client.Hostname == "" {
		resp.Diagnostics.AddError("Invalid mode error", InspectionModeError)
		return
	}

	// Determine the location.
{{- if .HasEntryName }}
	loc := {{ .structName }}Tfid{Name: state.Name.ValueString()}
{{- else }}
	loc := {{ .structName }}Tfid{}
{{- end }}


	// TODO: this needs to handle location structure for UUID style shared has nested structure type
	{{ RenderLocationsStateToPango "state.Location" "loc.Location" }}

	if err := loc.IsValid(); err != nil {
		resp.Diagnostics.AddError("Invalid location", err.Error())
		return
	}

	// Load the desired config.
	var obj *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
{{ $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{ $ev = "&ev" }}
	ev := make(map[string]types.String)
{{- end }}
	resp.Diagnostics.Append(state.CopyToPango(ctx, &obj, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
		// Timeout handling.
		ctx, cancel := context.WithTimeout(ctx, GetTimeout(state.Timeouts.Create))
		defer cancel()
	*/

	// Perform the operation.
{{- if .HasImports }}
	{{ RenderImportLocationAssignment "state.Location" }}
	created, err := r.manager.Create(ctx, loc.Location, []{{ .resourceSDKName }}.ImportLocation{location}, obj)
{{- else }}
	created, err := r.manager.Create(ctx, loc.Location, obj)
{{- end }}
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

	resp.Diagnostics.Append(state.CopyFromPango(ctx, created, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
{{- if .HasEncryptedResources }}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
{{- end }}

{{- if .HasEntryName }}
	state.Name = types.StringValue(created.Name)
{{- end }}

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

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

{{ $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
resp.Diagnostics.Append(savestate.EncryptedValues.ElementsAs(ctx, &ev, false)...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

elements := make(map[string]{{ $resourceTFStructName }})
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

entries := make([]*{{ $resourceSDKStructName }}, 0, len(elements))
for name, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	entries = append(entries, entry)
}

readEntries, err := o.manager.ReadMany(ctx, location, entries)
if err != nil {
	if errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.State.RemoveResource(ctx)
	} else {
		resp.Diagnostics.AddError("Failed to read entries from the server", err.Error())
	}
	return
}

objects := make(map[string]{{ $resourceTFStructName }})
for _, elt := range readEntries {
	var object {{ $resourceTFStructName }}
	resp.Diagnostics.Append(object.CopyFromPango(ctx, elt, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[elt.Name] = object
}

var map_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, map_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(map_diags...)
if resp.Diagnostics.HasError() {
	return
}

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

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

{{ $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
resp.Diagnostics.Append(savestate.EncryptedValues.ElementsAs(ctx, &ev, false)...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
entries := make([]*{{ $resourceSDKStructName }}, 0, len(elements))
for _, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
	entries = append(entries, entry)
}

{{ $exhaustive := "sdkmanager.NonExhaustive" }}
{{ if .Exhaustive }}
  {{ $exhaustive = "sdkmanager.Exhaustive" }}
{{- end }}
readEntries, err := o.manager.ReadMany(ctx, location, entries, {{ $exhaustive }})
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
	err := object.CopyFromPango(ctx, elt, {{ $ev }})
	resp.Diagnostics.Append(err...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects = append(objects, object)
}

var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}

resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
	{{ RenderLocationsStateToPango "savestate.Location" "loc.Location" }}
{{- else }}
	var loc {{ .resourceStructName }}Tfid
	// Parse the location from tfid.
	if err := DecodeLocation(savestate.Tfid.ValueString(), &loc); err != nil {
		resp.Diagnostics.AddError("Error parsing tfid", err.Error())
		return
	}
{{- end }}

{{ $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
	ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
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


	// Perform the operation.
{{- if .HasEntryName }}
	object, err := o.manager.Read(ctx, loc.Location, loc.Name)
{{- else }}
	object, err := o.manager.Read(ctx, loc.Location)
{{- end }}
	if err != nil {
		tflog.Warn(ctx, "KK: HERE3-1", map[string]any{"Error": err.Error()})
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

	copy_diags := state.CopyFromPango(ctx, object, {{ $ev }})
	resp.Diagnostics.Append(copy_diags...)

	/*
			// Keep the timeouts.
		    // TODO: This won't work for state import.
			state.Timeouts = savestate.Timeouts
	*/

	state.Location = savestate.Location
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

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "plan.Location" "location" }}

// Basic logging.
tflog.Info(ctx, "performing resource update", map[string]any{
	"resource_name": "panos_{{ UnderscoreName .structName }}",
	"function":      "Update",
})

var elements map[string]{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
stateEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx := 0
for name, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, nil)...)
	if resp.Diagnostics.HasError() {
		 return
	}
	entry.Name = name
	stateEntries[idx] = entry
	idx++
}

existing, err := r.manager.ReadMany(ctx, location, stateEntries)
if err != nil && !sdkerrors.IsObjectNotFound(err) {
	resp.Diagnostics.AddError("Error while reading entries from the server", err.Error())
	return
}

existingEntriesByName := make(map[string]*{{ $resourceSDKStructName }}, len(existing))
for _, elt := range existing {
	existingEntriesByName[elt.Name] = elt
}

plan.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
planEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx = 0
for name, elt := range elements {
	entry, _ := existingEntriesByName[name]
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry.Name = name
	planEntries[idx] = entry
	idx++
}

processed, err := r.manager.UpdateMany(ctx, location, stateEntries, planEntries)
if err != nil {
	resp.Diagnostics.AddError("Error while updating entries", err.Error())
	return
}

objects := make(map[string]*{{ $resourceTFStructName }}, len(processed))
for _, elt := range processed {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, elt, nil)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects[elt.Name] = &object
}

var list_diags diag.Diagnostics
plan.{{ .ListAttribute.CamelCase }}, list_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}

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

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "plan.Location" "location" }}

var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
stateEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var entry *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, nil)...)
	if resp.Diagnostics.HasError() {
		 return
	}
	stateEntries[idx] = entry
}

{{ $exhaustive := "sdkmanager.NonExhaustive" }}
{{- if .Exhaustive }}
  {{ $exhaustive = "sdkmanager.Exhaustive" }}
trueValue := true
position := rule.Position{First: &trueValue}
{{- else }}
position := state.Position.CopyToPango()
{{- end }}

existing, err := r.manager.ReadMany(ctx, location, stateEntries, {{ $exhaustive }})
if err != nil && !sdkerrors.IsObjectNotFound(err) {
	resp.Diagnostics.AddError("Error while reading entries from the server", err.Error())
	return
}

existingEntriesByName := make(map[string]*{{ $resourceSDKStructName }}, len(existing))
for _, elt := range existing {
	existingEntriesByName[elt.Name] = elt
}

plan.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
planEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	entry, _ := existingEntriesByName[elt.Name.ValueString()]
	resp.Diagnostics.Append(elt.CopyToPango(ctx, &entry, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}
	planEntries[idx] = entry
}

processed, err := r.manager.UpdateMany(ctx, location, stateEntries, planEntries, {{ $exhaustive }}, position)
if err != nil {
	resp.Diagnostics.AddError("Failed to udpate entries", err.Error())
}

objects := make([]*{{ $resourceTFStructName }}, len(processed))
for idx, elt := range processed {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, elt, nil)
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

{{- $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
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

{{- if .HasEntryName }}
	obj, err := r.manager.Read(ctx, loc.Location, loc.Name)
{{- else }}
	obj, err := r.manager.Read(ctx, loc.Location)
{{- end }}
	if err != nil {
		resp.Diagnostics.AddError("Error in update", err.Error())
		return
	}

	resp.Diagnostics.Append(plan.CopyToPango(ctx, &obj, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform the operation.
{{- if .HasEntryName }}
	updated, err := r.manager.Update(ctx, loc.Location, obj, loc.Name)
{{- else }}
	updated, err := r.manager.Update(ctx, loc.Location, obj)
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

	copy_diags := state.CopyFromPango(ctx, updated, {{ $ev }})
{{- if .HasEncryptedResources }}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
{{- end }}
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Done.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

{{- if .ResourceIsMap }}
elements := make(map[string]{{ $resourceTFStructName }}, len(state.{{ .ListAttribute.CamelCase }}.Elements()))
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
{{- else }}
var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
{{- end }}

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

var names []string
{{- if .ResourceIsMap }}
for name, _ := range elements {
	names = append(names, name)
}
{{- else }}
for _, elt := range elements {
	names = append(names, elt.Name.ValueString())
}
{{- end }}

{{- if .ResourceIsMap }}
err := r.manager.Delete(ctx, location, names)
{{- else if .Exhaustive }}
err := r.manager.Delete(ctx, location, names, sdkmanager.Exhaustive)
{{- else }}
err := r.manager.Delete(ctx, location, names, sdkmanager.NonExhaustive)
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

{{- if .HasEntryName }}
  {{- if .HasImports }}
	{{ RenderImportLocationAssignment "state.Location" }}
	err := r.manager.Delete(ctx, loc.Location, []{{ .resourceSDKName }}.ImportLocation{location}, []string{loc.Name}, sdkmanager.NonExhaustive)
  {{- else }}
	err := r.manager.Delete(ctx, loc.Location, []string{loc.Name})
  {{- end }}
	if err != nil && !errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
	}
{{- else }}

{{- $ev := "nil" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
	ev := make(map[string]types.String)
{{- end }}
	var obj *{{ $resourceSDKStructName }}
	resp.Diagnostics.Append(state.CopyToPango(ctx, &obj, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.manager.Delete(ctx, loc.Location, obj)
	if err != nil && errors.Is(err, sdkmanager.ErrObjectNotFound) {
		resp.Diagnostics.AddError("Error in delete", err.Error())
	}
{{- end }}
`

const commonTemplate = `
{{- RenderLocationStructs }}

{{- RenderLocationSchemaGetter }}

{{- RenderCustomCommonCode }}
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

{{- if not GoSDKSkipped }}
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
{{- end }}


{{ RenderDataSourceStructs }}

{{ RenderCopyToPangoFunctions }}

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

{{- if IsCustom }}

{{- else if and IsEntry HasImports }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	d.manager =  sdkmanager.NewImportableEntryObjectManager(d.client, {{ resourceSDKName }}.NewService(d.client), specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsEntry }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	d.manager =  sdkmanager.NewEntryObjectManager(d.client, {{ resourceSDKName }}.NewService(d.client), specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsUuid }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	d.manager =  sdkmanager.NewUuidObjectManager(d.client, {{ resourceSDKName }}.NewService(d.client), specifier, {{ resourceSDKName }}.SpecMatches)
{{- else if IsConfig }}
	specifier, _, err := {{ resourceSDKName }}.Versioning(d.client.Versioning())
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure SDK client", err.Error())
		return
	}
	d.manager =  sdkmanager.NewConfigObjectManager(d.client, {{ resourceSDKName }}.NewService(d.client), specifier)
{{- end }}
}

func (o *{{ dataSourceStructName }}) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	{{ DataSourceReadFunction dataSourceStructName serviceName }}
}

{{- /* Done */ -}}
`

const providerFile = `
{{- /* Begin */ -}}
package provider

// Note:  This file is automatically generated.  Manually made changes
// will be overwritten when the provider is generated.
{{ renderImports }}
{{ renderCustomImports }}
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

func (p *PanosProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewAddressValueFunction,
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
