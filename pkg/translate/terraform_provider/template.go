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

{{ RenderResourceStructs }}

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

{{ $ev := "" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
{{- else }}
  {{- $ev = "nil" }}
{{- end }}


type entryWithState struct {
	Entry    *{{ $resourceSDKStructName }}
	StateIdx int
}

{{- if .ResourceIsMap }}
var elements map[string]{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
entries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx := 0
for name, elt := range elements {
	var list_diags diag.Diagnostics
	var entry *{{ .resourceSDKName }}.{{ .EntryOrConfig }}
	entry, list_diags = elt.CopyToPango(ctx, {{ $ev }})
	resp.Diagnostics.Append(list_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	entries[idx] = entry
	idx++
}
{{- else }}
var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
entries := make([]*{{ $resourceSDKStructName }}, len(elements))

planEntriesByName := make(map[string]*entryWithState, len(elements))
for idx, elt := range elements {
	var list_diags diag.Diagnostics
	var entry *{{ $resourceSDKStructName }}
	entry, list_diags = elt.CopyToPango(ctx, {{ $ev }})
	resp.Diagnostics.Append(list_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planEntriesByName[elt.Name.ValueString()] = &entryWithState{
		Entry: entry,
		StateIdx: idx,
	}
	entries[idx] = entry
}
{{- end }}

svc := {{ .resourceSDKName }}.NewService(r.client)

// First, check if none of the entries from the plan already exist on the server
existing, err := svc.List(ctx, location, "get", "", "")
if err != nil && err.Error() != "Object not found" {
	resp.Diagnostics.AddError("sdk error while listing resources", err.Error())
	return
}

{{- if .ResourceIsMap }}
for _, elt := range existing {
	_, foundInPlan := elements[elt.Name]

	if foundInPlan {
		errorMsg := fmt.Sprintf("%s created outside of terraform", elt.Name)
		resp.Diagnostics.AddError("Conflict between plan and server data", errorMsg)
		return
	}
}
{{- else }}
for _, elt := range existing {
	_, foundInPlan := planEntriesByName[elt.Name]

	if foundInPlan {
		errorMsg := fmt.Sprintf("%s created outside of terraform", elt.Name)
		resp.Diagnostics.AddError("Conflict between plan and server data", errorMsg)
		return
	}
}
{{- end }}

specifier, _, err := {{ .resourceSDKName }}.Versioning(r.client.Versioning())
if err != nil {
	resp.Diagnostics.AddError("error while creating specifier", err.Error())
	return
}

updates := xmlapi.NewMultiConfig(len(elements))

for _, elt := range entries {
	path, err := location.XpathWithEntryName(r.client.Versioning(), elt.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create xpath for existing entry", err.Error())
		return
	}

	xmlEntry, err := specifier(*elt)
	if err != nil {
		resp.Diagnostics.AddError("Failed to transform Entry into XML document", err.Error())
		return
	}

	updates.Add(&xmlapi.Config{
		Action:  "edit",
		Xpath:   util.AsXpath(path),
		Element: xmlEntry,
		Target:  r.client.GetTarget(),
	})
}

if len(updates.Operations) > 0 {
	if _, _, _, err := r.client.MultiConfig(ctx, updates, false, nil); err != nil {
		resp.Diagnostics.AddError("error updating entries", err.Error())
		return
	}
}

existing, err = svc.List(ctx, location, "get", "", "")
if err != nil && err.Error() != "Object not found" {
	resp.Diagnostics.AddError("sdk error while listing resources", err.Error())
	return
}

{{- if and .Exhaustive (not .ResourceIsMap) }}
// We manage the entire list of PAN-OS objects, so the order of entries
// from the plan is compared against all existing PAN-OS objects.
var movementRequired bool
for idx, elt := range existing {
	if planEntriesByName[elt.Name].StateIdx != idx {
		movementRequired = true
	}
	planEntriesByName[elt.Name].Entry.Uuid = elt.Uuid
}
{{- else if and (not .Exhaustive) (not .ResourceIsMap) }}
// We only manage a subset of PAN-OS object on the given list, so care
// has to be taken to calculate the order of those managed elements on the
// PAN-OS side.

// We filter all existing entries to end up with a list of entries that
// are in the plan. For every element of that list, we store its PAN-OS
// list index as StateIdx. Finally, the managedEntries index will serve
// as a way to check if managed entries are in order relative to each
// other.
var movementRequired bool
managedEntries := make([]*entryWithState, len(entries))
for idx, elt := range existing {
	if planEntry, found := planEntriesByName[elt.Name]; found {
		planEntry.Entry.Uuid = elt.Uuid
		managedEntries = append(managedEntries, &entryWithState{
			Entry: &elt,
			StateIdx: idx,
		})
	}
}

var previousManagedEntry, previousPlannedEntry *entryWithState
for idx, elt := range managedEntries {
	// plannedEntriesByName is a map of entries from the plan indexed by their
	// name. If idx doesn't match StateIdx of the entry from the plan, the PAN-OS
	// object is out of order.
	plannedEntry := planEntriesByName[elt.Entry.Name]
	if plannedEntry.StateIdx != idx {
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

{{- end }}

{{- if and .Exhaustive (not .ResourceIsMap) }}
if movementRequired {
	entries := make([]{{ $resourceSDKStructName }}, len(planEntriesByName))
	for _, elt := range planEntriesByName {
		entries[elt.StateIdx] = *elt.Entry
	}
	trueValue := true
	err = svc.MoveGroup(ctx, location, rule.Position{First: &trueValue}, entries)
	if err != nil {
		resp.Diagnostics.AddError("Failed to reorder entries", err.Error())
		return
	}
}
{{- else if and (not .Exhaustive) (not .ResourceIsMap) }}
if movementRequired {
	entries := make([]{{ $resourceSDKStructName }}, len(managedEntries))
	for _, elt := range managedEntries {
		entries[elt.StateIdx] = *elt.Entry
	}
	trueValue := true
	err = svc.MoveGroup(ctx, location, rule.Position{First: &trueValue}, entries)
	if err != nil {
		resp.Diagnostics.AddError("Failed to reorder entries", err.Error())
		return
	}
}
{{- end }}

{{- if and (not .Exhaustive) .ResourceIsMap }}
for _, elt := range existing {
	if _, found := elements[elt.Name]; !found {
		continue
	}
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, &elt, {{ $ev }})
	resp.Diagnostics.Append(copy_diags...)
	elements[elt.Name] = object
}

if resp.Diagnostics.HasError() {
	return
}

var map_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, map_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), elements)
resp.Diagnostics.Append(map_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- else }}
objects := make([]{{ $resourceTFStructName }}, len(planEntriesByName))
for idx, elt := range existing {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, &elt, {{ $ev }})
	resp.Diagnostics.Append(copy_diags...)
	objects[idx] = object
}

if resp.Diagnostics.HasError() {
	return
}
var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

{{- if .HasEncryptedResources }}
	{
		copy_diags := state.CopyFromPango(ctx, create, &ev)
		resp.Diagnostics.Append(copy_diags...)
	}
	ev_map, ev_diags := types.MapValueFrom(ctx, types.StringType, ev)
        state.EncryptedValues = ev_map
        resp.Diagnostics.Append(ev_diags...)
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

	// Create the service.
	svc := {{ .resourceSDKName }}.NewService(r.client)

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

const resourceReadManyFunction = `
{{- $structName := "" }}
{{- if eq .ResourceOrDS "DataSource" }}
  {{ $structName = .dataSourceStructName }}
{{- else }}
  {{ $structName = .resourceStructName }}
{{- end }}
{{- $resourceSDKStructName := printf "%s.%s" .resourceSDKName .EntryOrConfig }}
{{- $resourceTFStructName := printf "%s%sObject" $structName .ListAttribute.CamelCase }}
// {{ $resourceSDKStructName }}
// {{ $resourceTFStructName }}

{{- $stateName := "" }}
{{- if eq .ResourceOrDS "DataSource" }}
  {{- $stateName = "Config" }}
{{- else }}
  {{- $stateName = "State" }}
{{- end }}
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

svc := {{ .resourceSDKName }}.NewService(o.client)

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "state.Location" "location" }}

existing, err := svc.List(ctx, location, "get", "", "")
if err != nil && err.Error() != "Object not found" {
	resp.Diagnostics.AddError("sdk error during read", err.Error())
	return
}

{{ $ev := "" }}
{{- if .HasEncryptedResources }}
  {{- $ev = "&ev" }}
ev := make(map[string]types.String, len(state.EncryptedValues.Elements()))
resp.Diagnostics.Append(savestate.EncryptedValues.ElementsAs(ctx, &ev, false)...)
if resp.Diagnostics.HasError() {
	return
}
{{- else }}
  {{- $ev = "nil" }}
{{- end }}

{{- if and .Exhaustive .ResourceIsMap }}
elements = make(map[string]{{ $resourceTFStructName }}, len(existing))
for idx, elt := range existing {
	var object {{ $resourceTFStructName }}
	object.CopyFromPango(ctx, &elt, {{ $ev }})
	elements[object.Name] = object
}
{{- else if and .Exhaustive (not .ResourceIsMap) }}
// For resources that take sole ownership of a given list, Read()
// will return all existing entries from the server.
objects := make([]{{ $resourceTFStructName }}, len(existing))
for idx, elt := range existing {
	var object {{ $resourceTFStructName }}
	object.CopyFromPango(ctx, &elt, {{ $ev }})
	objects[idx] = object
}
{{- else if and (not .Exhaustive) .ResourceIsMap }}
elements := make(map[string]{{ $resourceTFStructName }})
resp.Diagnostics.Append(state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)...)
if resp.Diagnostics.HasError() {
	return
}

objects := make(map[string]{{ $resourceTFStructName }})
for _, elt := range existing {
	if _, found := elements[elt.Name]; !found {
		continue
	}
	var object {{ $resourceTFStructName }}
	resp.Diagnostics.Append(object.CopyFromPango(ctx, &elt, {{ $ev }})...)
	if resp.Diagnostics.HasError() {
		return
	}
	objects[elt.Name] = object
}
{{- else if and (not .Exhaustive) (not .ResourceIsMap) }}
// For resources that only manage their own items in the list, Read()
// must only objects that are already part of the state.
var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
stateObjectsByName := make(map[string]*{{ $resourceTFStructName }}, len(elements))
for _, elt := range elements {
	stateObjectsByName[elt.Name.ValueString()] = &elt
}
objects := make([]{{ .structName }}{{ .ResourceOrDS }}{{ .ListAttribute.CamelCase }}Object, len(state.{{ .ListAttribute.CamelCase }}.Elements()))
for idx, elt := range existing {
	if _, found := stateObjectsByName[elt.Name]; !found {
		continue
	}
	var object {{ .structName }}{{ .ResourceOrDS }}{{ .ListAttribute.CamelCase }}Object
	object.CopyFromPango(ctx, &elt, nil)
	objects[idx] = object
}
{{- else }}
panic("Unsupported combination of .Exhaustive and .ResourceIsMap" }})
{{- end }}


{{- if .ResourceIsMap }}
var map_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, map_diags = types.MapValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(map_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- else }}
var list_diags diag.Diagnostics
state.{{ .ListAttribute.CamelCase }}, list_diags = types.ListValueFrom(ctx, state.getTypeFor("{{ .ListAttribute.Underscore }}"), objects)
resp.Diagnostics.Append(list_diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}

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

	{{ RenderLocationsPangoToState "loc.Location" "state.Location" }}

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

svc := {{ .resourceSDKName }}.NewService(r.client)

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
	var list_diags diag.Diagnostics
	var entry *{{ $resourceSDKStructName }}
	entry, list_diags = elt.CopyToPango(ctx, nil)
	resp.Diagnostics.Append(list_diags...)
	if resp.Diagnostics.HasError() {
		 return
	}
	entry.Name = name
	stateEntries[idx] = entry
	idx++
}

plan.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
planEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
idx = 0
for name, elt := range elements {
	var list_diags diag.Diagnostics
	var entry *{{ $resourceSDKStructName }}
	entry, list_diags = elt.CopyToPango(ctx, nil)
	resp.Diagnostics.Append(list_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry.Name = name
	planEntries[idx] = entry
	idx++
}

type entryState string
const entryUnknown entryState = "unknown"
const entryMissing entryState = "missing"
const entryOutdated entryState = "outdated"
const entryRenamed entryState = "renamed"
const entryOk entryState = "ok"

type entryWithState struct {
	Entry    *{{ $resourceSDKStructName }}
	State    entryState
	NewName  string
}

stateEntriesByName := make(map[string]entryWithState, len(stateEntries))
for _, elt := range stateEntries {
	stateEntriesByName[elt.Name] = entryWithState{
		Entry:    elt,
	}
}

planEntriesByName := make(map[string]entryWithState, len(planEntries))
for _, elt := range planEntries {
	planEntriesByName[elt.Name] = entryWithState{
		Entry:    elt,
	}
}


renamedEntries := make(map[string]bool)
processedStateEntries := make(map[string]*entryWithState)

findMatchingStateEntry := func(entry *{{ $resourceSDKStructName }}) (*{{ $resourceSDKStructName }}, bool) {
	var found *{{ $resourceSDKStructName }}

	for _, elt := range stateEntriesByName {
		if {{ .resourceSDKName }}.SpecMatches(entry, elt.Entry) {
			found = elt.Entry
			break
		}
	}

	if found == nil {
		return nil, false
	}

	// If matched entry already exists in the plan, this is not a rename
	// but adding a missing entry.
	if _, ok := planEntriesByName[found.Name]; ok {
		return nil, false
	}

	return found, true
}


for _, elt := range planEntries {
	var processedEntry *entryWithState

	if stateElt, found := stateEntriesByName[elt.Name]; !found {
		// If given plan entry is not found in state, check if there is another
		// entry that matches it without name. If so, this plan entry is a rename.
		// Keep the renamedEntry Index, and set its state to entryRename.
		if stateEntry, found := findMatchingStateEntry(elt); found {
			if _, found := renamedEntries[stateEntry.Name]; found {
				resp.Diagnostics.AddError("Failed to generate update actions", "Entry name swapped between entries")
				return
			}
			processedEntry = &entryWithState{
				Entry:    stateEntry,
				State:    entryRenamed,
				NewName:  elt.Name,
			}
			renamedEntries[elt.Name] = true
		} else {
			processedEntry = &entryWithState{
				Entry:    elt,
				State:    entryMissing,
			}
		}

		// If there is already a processed entry with state entryMissing, it means
		// we've encountered a new entry with the name matching renamedEntry old name.
		// It will have state entryOutdated because its spec didn't match spec of the
		// entry about to be renamed.
		// Change its state to entryMissing instead, and update its index to match
		// index from the plan.
		if previousEntry, found := processedStateEntries[processedEntry.Entry.Name]; found {
			if previousEntry.State != entryOutdated && previousEntry.State != entryMissing {
				resp.Diagnostics.AddError(
					"failed to create a list of entries to process",
					fmt.Sprintf("previousEntry.State '%s' != entryOutdated", previousEntry.State))
				return
			}
		}
		processedStateEntries[processedEntry.Entry.Name] = processedEntry
	} else {
		processedEntry = &entryWithState{
			Entry:    elt,
			State: entryMissing,
		}

		if !{{ .resourceSDKName }}.SpecMatches(elt, stateElt.Entry) {
			processedEntry.State = entryOutdated
		}

		processedStateEntries[elt.Name] = processedEntry
	}

}

existing, err := svc.List(ctx, location, "get", "", "")
if err != nil && err.Error() != "Object not found" {
	resp.Diagnostics.AddError("sdk error while listing resources", err.Error())
	return
}

updates := xmlapi.NewMultiConfig(len(planEntries))

// Iterate over all existing entries as returned from the server, comparing
// them to processedEntries.
for _, existingElt := range existing {
	path, err := location.XpathWithEntryName(r.client.Versioning(), existingElt.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create xpath for existing entry", err.Error())
	}

	_, foundInState := stateEntriesByName[existingElt.Name]
        _, foundInRenamed := renamedEntries[existingElt.Name]
	_, foundInPlan := planEntriesByName[existingElt.Name]

	if !foundInState && (foundInRenamed || foundInPlan) {
		errorMsg := fmt.Sprintf("Resource '%s' created outside of terraform", existingElt.Name)
		resp.Diagnostics.AddError("Conflict between Terraform and PAN-OS", errorMsg)
		return
	}

	// If the existing entry name matches new name for the renamed entry,
	// we delete it before adding Renamed commands.
	if _, found := renamedEntries[existingElt.Name]; found {
		updates.Add(&xmlapi.Config{
			Action: "delete",
			Xpath:  util.AsXpath(path),
			Target: r.client.GetTarget(),
		})
		continue
	}

	processedElt, found := processedStateEntries[existingElt.Name]
	if !found {
		// If existing entry is not found in the processedEntries map, it's not
		// entry we are managing and it should be deleted.
		updates.Add(&xmlapi.Config{
			Action: "delete",
			Xpath:  util.AsXpath(path),
			Target: r.client.GetTarget(),
		})
	} else {
		// XXX: If entry from the plan is in process of being renamed, and its content
		// differs from what exists on the server we should switch its state to entryOutdated
		// instead.
		if processedElt.State == entryRenamed {
			continue
		}

		if !{{ .resourceSDKName }}.SpecMatches(processedElt.Entry, &existingElt) {
			processedElt.State = entryOutdated
		} else {
			processedElt.State = entryOk
		}
	}
}

specifier, _, err := {{ .resourceSDKName }}.Versioning(r.client.Versioning())
if err != nil {
	resp.Diagnostics.AddError("error while creating specifier", err.Error())
	return
}

for _, elt := range processedStateEntries {
	path, err := location.XpathWithEntryName(r.client.Versioning(), elt.Entry.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create xpath for existing entry", err.Error())
		return
	}

	xmlEntry, err := specifier(*elt.Entry)
	if err != nil {
		resp.Diagnostics.AddError("Failed to transform Entry into XML document", err.Error())
		return
	}

	switch elt.State {
	case entryMissing, entryOutdated:
		tflog.Debug(ctx, "HERE5-1 Missing or Outdated", map[string]any{"Name": elt.Entry.Name, "State": elt.State})
		updates.Add(&xmlapi.Config{
			Action:  "edit",
			Xpath:   util.AsXpath(path),
			Element: xmlEntry,
			Target:  r.client.GetTarget(),
		})
	case entryRenamed:
		tflog.Debug(ctx, "HERE5-1 Renamed", map[string]any{"Name": elt.Entry.Name, "State": elt.State})
		updates.Add(&xmlapi.Config{
			Action:  "rename",
			Xpath:   util.AsXpath(path),
			NewName: elt.NewName,
			Target:  r.client.GetTarget(),
		})

		// If existing entry is found in our plan with state entryRenamed,
		// we move entry in processedEntries from old name to the new name,
		// indicating it has been renamed.
		// This is used later when we assign uuids to all entries.
		delete(processedStateEntries, elt.Entry.Name)
		elt.Entry.Name = elt.NewName
		processedStateEntries[elt.NewName] = elt
	case entryUnknown:
		tflog.Debug(ctx, "HERE5-1 Unknown", map[string]any{"Name": elt.Entry.Name, "State": elt.State})
	case entryOk:
		tflog.Debug(ctx, "HERE5-1 OK", map[string]any{"Name": elt.Entry.Name, "State": elt.State})
		// Nothing to do for entries that have no changes
	}
}

if len(updates.Operations) > 0 {
	if _, _, _, err := r.client.MultiConfig(ctx, updates, false, nil); err != nil {
		resp.Diagnostics.AddError("error updating entries", err.Error())
		return
	}
}

objects := make(map[string]*{{ $resourceTFStructName }}, len(processedStateEntries))
for _, elt := range processedStateEntries {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, elt.Entry, nil)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects[elt.Entry.Name] = &object
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

svc := {{ .resourceSDKName }}.NewService(r.client)

var location {{ .resourceSDKName }}.Location
{{ RenderLocationsStateToPango "plan.Location" "location" }}

var elements []{{ $resourceTFStructName }}
state.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
stateEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var list_diags diag.Diagnostics
	var entry *{{ $resourceSDKStructName }}
	entry, list_diags = elt.CopyToPango(ctx, nil)
	resp.Diagnostics.Append(list_diags...)
	if resp.Diagnostics.HasError() {
		 return
	}
	stateEntries[idx] = entry
}

plan.{{ .ListAttribute.CamelCase }}.ElementsAs(ctx, &elements, false)
planEntries := make([]*{{ $resourceSDKStructName }}, len(elements))
for idx, elt := range elements {
	var list_diags diag.Diagnostics
	var entry *{{ $resourceSDKStructName }}
	entry, list_diags = elt.CopyToPango(ctx, nil)
	resp.Diagnostics.Append(list_diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	planEntries[idx] = entry
}

type entryState string
const entryUnknown entryState = "unknown"
const entryMissing entryState = "missing"
const entryOutdated entryState = "outdated"
const entryRenamed entryState = "renamed"
const entryOk entryState = "ok"

type entryWithState struct {
	Entry    *{{ $resourceSDKStructName }}
	State    entryState
	StateIdx int
	NewName  string
}

stateEntriesByName := make(map[string]*entryWithState, len(stateEntries))
for idx, elt := range stateEntries {
	stateEntriesByName[elt.Name] = &entryWithState{
		Entry:    elt,
		StateIdx: idx,
	}
}

planEntriesByName := make(map[string]*entryWithState, len(planEntries))
for idx, elt := range planEntries {
	planEntriesByName[elt.Name] = &entryWithState{
		Entry:    elt,
		StateIdx: idx,
	}
}

findMatchingStateEntry := func(entry *{{ $resourceSDKStructName }}) (*{{ $resourceSDKStructName }}, bool) {
	var found *{{ $resourceSDKStructName }}

	for _, elt := range stateEntriesByName {
		entry.Uuid = elt.Entry.Uuid
		if {{ .resourceSDKName }}.SpecMatches(entry, elt.Entry) {
			found = elt.Entry
			break
		}
	}
	entry.Uuid = nil

	if found == nil {
		return nil, false
	}

	// If matched entry already exists in the plan, this is not a rename
	// but adding a missing entry.
	if _, ok := planEntriesByName[found.Name]; ok {
		return nil, false
	}

	return found, true
}

renamedEntries := make(map[string]bool)
processedStateEntries := make(map[string]*entryWithState)

for idx, elt := range planEntries {
	var processedEntry *entryWithState

	if stateElt, found := stateEntriesByName[elt.Name]; !found {
		// If given plan entry is not found in state, check if there is another
		// entry that matches it without name. If so, this plan entry is a rename.
		// Keep the renamedEntry Index, and set its state to entryRename.
		if renamedEntry, found := findMatchingStateEntry(elt); found {
			if _, found := renamedEntries[renamedEntry.Name]; found {
				resp.Diagnostics.AddError("Failed to generate update actions", "Entry name swapped between entries")
				return
			}
			processedEntry = &entryWithState{
				Entry:    renamedEntry,
				State:    entryRenamed,
				StateIdx: stateEntriesByName[renamedEntry.Name].StateIdx,
				NewName:  elt.Name,
			}
			renamedEntries[elt.Name] = true
		} else {
			processedEntry = &entryWithState{
				Entry:    elt,
				State:    entryMissing,
				StateIdx: idx,
			}
		}

		// If there is already a processed entry with state entryMissing, it means
		// we've encountered a new entry with the name matching renamedEntry old name.
		// It will have state entryOutdated because its spec didn't match spec of the
		// entry about to be renamed.
		// Change its state to entryMissing instead, and update its index to match
		// index from the plan.
		if previousEntry, found := processedStateEntries[processedEntry.Entry.Name]; found {
			if previousEntry.State != entryOutdated {
				resp.Diagnostics.AddError(
					"failed to create a list of entries to process",
					fmt.Sprintf("previousEntry.State '%s' != entryOutdated", previousEntry.State))
				return
			}
		}
		processedStateEntries[processedEntry.Entry.Name] = processedEntry
	} else {
		processedEntry = &entryWithState{
			Entry:    elt,
			StateIdx: idx,
		}

		elt.Uuid = stateElt.Entry.Uuid
		if {{ .resourceSDKName }}.SpecMatches(elt, stateElt.Entry) {
			processedEntry.State = entryOk
		} else {
			processedEntry.State = entryOutdated
		}
		elt.Uuid = nil

		processedStateEntries[elt.Name] = processedEntry
	}

}

existing, err := svc.List(ctx, location, "get", "", "")
if err != nil && err.Error() != "Object not found" {
	resp.Diagnostics.AddError("sdk error while listing resources", err.Error())
	return
}

updates := xmlapi.NewMultiConfig(len(planEntries))

// Iterate over all existing entries as returned from the server, comparing
// them to processedEntries.
for _, existingElt := range existing {
	path, err := location.XpathWithEntryName(r.client.Versioning(), existingElt.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create xpath for existing entry", err.Error())
	}

	_, foundInState := stateEntriesByName[existingElt.Name]
        _, foundInRenamed := renamedEntries[existingElt.Name]
	_, foundInPlan := planEntriesByName[existingElt.Name]

	if !foundInState && (foundInRenamed || foundInPlan) {
		errorMsg := fmt.Sprintf("%s created outside of terraform", existingElt.Name)
		resp.Diagnostics.AddError("Conflict between plan and PAN-OS data", errorMsg)
		return
	}

	// If the existing entry name matches new name for the renamed entry,
	// we delete it before adding Renamed commands.
	if _, found := renamedEntries[existingElt.Name]; found {
		updates.Add(&xmlapi.Config{
			Action: "delete",
			Xpath:  util.AsXpath(path),
			Target: r.client.GetTarget(),
		})
		continue
	}

	processedElt, found := processedStateEntries[existingElt.Name]
{{- if .Exhaustive }}
	if !found {
		// If existing entry is not found in the processedEntries map, it's not
		// entry we are managing and it should be deleted.
		updates.Add(&xmlapi.Config{
			Action: "delete",
			Xpath:  util.AsXpath(path),
			Target: r.client.GetTarget(),
		})
		continue
	}
{{- end }}
	if found && processedElt.Entry.Uuid != nil && *processedElt.Entry.Uuid == *existingElt.Uuid {
		// XXX: If entry from the plan is in process of being renamed, and its content
		// differs from what exists on the server we should switch its state to entryOutdated
		// instead.
		if processedElt.State == entryRenamed {
			continue
		}

		if !{{ .resourceSDKName }}.SpecMatches(processedElt.Entry, &existingElt) {
			processedElt.State = entryOutdated
		} else {
			processedElt.State = entryOk
		}
	}
}

specifier, _, err := {{ .resourceSDKName }}.Versioning(r.client.Versioning())
if err != nil {
	resp.Diagnostics.AddError("error while creating specifier", err.Error())
	return
}

for _, elt := range processedStateEntries {
	path, err := location.XpathWithEntryName(r.client.Versioning(), elt.Entry.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create xpath for existing entry", err.Error())
		return
	}

	xmlEntry, err := specifier(*elt.Entry)
	if err != nil {
		resp.Diagnostics.AddError("Failed to transform Entry into XML document", err.Error())
		return
	}

	switch elt.State {
	case entryMissing, entryOutdated:
		updates.Add(&xmlapi.Config{
			Action:  "edit",
			Xpath:   util.AsXpath(path),
			Element: xmlEntry,
			Target:  r.client.GetTarget(),
		})
	case entryRenamed:
		updates.Add(&xmlapi.Config{
			Action:  "rename",
			Xpath:   util.AsXpath(path),
			NewName: elt.NewName,
			Target:  r.client.GetTarget(),
		})

		// If existing entry is found in our plan with state entryRenamed,
		// we move entry in processedEntries from old name to the new name,
		// indicating it has been renamed.
		// This is used later when we assign uuids to all entries.
		delete(processedStateEntries, elt.Entry.Name)
		elt.Entry.Name = elt.NewName
		processedStateEntries[elt.NewName] = elt
	case entryOk:
		// Nothing to do for entries that have no changes
	}
}

if len(updates.Operations) > 0 {
	if _, _, _, err := r.client.MultiConfig(ctx, updates, false, nil); err != nil {
		resp.Diagnostics.AddError("error updating entries", err.Error())
		return
	}
}

existing, err = svc.List(ctx, location, "get", "", "")
if err != nil && err.Error() != "Object not found" {
	resp.Diagnostics.AddError("sdk error while listing resources", err.Error())
	return
}

var movementRequired bool
for idx, elt := range existing {
	if processedStateEntries[elt.Name].StateIdx != idx {
		movementRequired = true
	}
	processedStateEntries[elt.Name].Entry.Uuid = elt.Uuid
}

if movementRequired {
	entries := make([]{{ $resourceSDKStructName }}, len(processedStateEntries))
	for _, elt := range processedStateEntries {
		entries[elt.StateIdx] = *elt.Entry
	}

	var finalOrder []string
	for _, elt := range entries {
		finalOrder = append(finalOrder, elt.Name)
	}

	trueValue := true
	err = svc.MoveGroup(ctx, location, rule.Position{First: &trueValue}, entries)
	if err != nil {
		resp.Diagnostics.AddError("Failed to reorder entries", err.Error())
		return
	}
}

objects := make([]*{{ $resourceTFStructName }}, len(processedStateEntries))
for _, elt := range processedStateEntries {
	var object {{ $resourceTFStructName }}
	copy_diags := object.CopyFromPango(ctx, elt.Entry, nil)
	resp.Diagnostics.Append(copy_diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects[elt.StateIdx] = &object
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

updates := xmlapi.NewMultiConfig(len(elements))

{{- if .Exhausitive }}
existing, err := svc.List(ctx, location, "get", "", "")
if err != nil {
	resp.Diagnostics.AddError("sdk error while listing entries", err.Error())
}
for _, elt := range existing {
	path, err := location.XpathWithEntryName(r.client.Versioning(), elt.Name)
	if err != nil {
		resp.Diagnostics.AddError("sdk error while creating xpath", err.Error())
	}
	updates.Add(&xmlapi.Config{
		Action: "delete",
		Xpath:  util.AsXpath(path),
		Target: r.client.GetTarget(),
	})
}
{{- else if .ResourceIsMap }}
for name, _ := range elements {
	path, err := location.XpathWithEntryName(r.client.Versioning(), name)
	if err != nil {
		resp.Diagnostics.AddError("sdk error while creating xpath", err.Error())
	}
	updates.Add(&xmlapi.Config{
		Action: "delete",
		Xpath:  util.AsXpath(path),
		Target: r.client.GetTarget(),
	})
}
{{- else }}
for _, elt := range elements {
	path, err := location.XpathWithEntryName(r.client.Versioning(), elt.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("sdk error while creating xpath", err.Error())
	}
	updates.Add(&xmlapi.Config{
		Action: "delete",
		Xpath:  util.AsXpath(path),
		Target: r.client.GetTarget(),
	})
}
{{- end }}

if len(updates.Operations) > 0 {
	if _, _, _, err := r.client.MultiConfig(ctx, updates, false, nil); err != nil {
		resp.Diagnostics.AddError("error updating entries", err.Error())
		return
	}
}
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
