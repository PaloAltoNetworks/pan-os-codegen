package terraform

import (
	"fmt"
	_ "sort"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/normalized"
)

func Resource(name string, namespaces map[string]*normalized.Namespace, schemas map[string]normalized.Item, spec *properties.Normalization) (string, string, *imports.Manager, error) {
	// Sanity checks.
	ns, ok := namespaces[name]
	if !ok {
		return "", "", nil, fmt.Errorf("namespace:%q not present", name)
	} else if ns.Create == nil || ns.Read == nil || ns.Update == nil || ns.Delete == nil {
		return "", "", nil, nil
	} else if ns.ShortName == "" {
		return "", "", nil, fmt.Errorf("ns.ShortName is empty")
	} else if !ns.Create.FunctionHasInput() {
		return "", "", nil, fmt.Errorf("ns:%q.Create does not have input", name)
	} else if spec.Name == "" {
		return "", "", nil, fmt.Errorf("spec.Name is not defined")
	} else if spec.Repository == "" {
		return "", "", nil, fmt.Errorf("repository is not defined")
	} else if spec.RepositoryShortName == "" {
		return "", "", nil, fmt.Errorf("repository_short_name is not defined")
	}

	var locMap map[string]int

	hasEncryptedValues := false
	if ns.Create.Request != nil {
		hev, err := ns.Create.Request.HasEncryptedItems(schemas)
		if err != nil {
			return "", "", nil, err
		}
		hasEncryptedValues = hev
	}

	tfName := ns.ModuleSuffix()
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, "Resource", false)
	modelName := naming.CamelCase("", tfName, "RsModel", false)
	newFuncName := naming.CamelCase("New", structName, "", true)
	namer := naming.NewNamer()

	// Add imports.
	manager := imports.NewManager()
	manager.AddStandardImport("context", "")
	manager.AddSdkImport(spec.Repository, "")
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")

	// Define the funcmap.
	fm := template.FuncMap{
		"MetaName":         func() string { return metaName },
		"StructName":       func() string { return structName },
		"ModelName":        func() string { return modelName },
		"NewFuncName":      func() string { return newFuncName },
		"RepoShortName":    func() string { return spec.RepositoryShortName },
		"ServiceShortName": func() string { return ns.ShortName },
		"ProviderName":     func() string { return spec.Name },
		"ResetVarNaming": func() string {
			namer.ResetVarNaming()
			return ""
		},
		"NewVarName":         func() string { return namer.NextVarName() },
		"CreateFunc":         func() *normalized.Function { return ns.Create },
		"ReadFunc":           func() *normalized.Function { return ns.Read },
		"UpdateFunc":         func() *normalized.Function { return ns.Update },
		"DeleteFunc":         func() *normalized.Function { return ns.Delete },
		"HasEncryptedValues": func() bool { return hasEncryptedValues },
		"Evccn":              func() string { return EncryptedValuesCamelCaseName },
		"LocMap":             func() map[string]int { return locMap },
		"HasInput":           func(v *normalized.Function) bool { return v.FunctionHasInput() },
		"AsTerraformModel": func() (string, error) {
			code, libs, err := AsTerraformModel(ns.Create, 'r', modelName, ns.ShortName, "tfid", schemas)
			manager.Merge(libs)
			return code, err
		},
		"AsTerraformSchema": func() (string, error) {
			code, libs, err := AsTerraformSchema(ns.Create, "rsschema", "tfid", nil, schemas)
			manager.Merge(libs)
			return code, err
		},
		"AsTerraformId": func() (string, error) {
			code, lm, libs, err := AsTerraformId(ns.Create, ns.Read, schemas)
			locMap = lm
			manager.Merge(libs)
			return code, err
		},
		"AsTerraformInput": func(theFunc *normalized.Function, tfType byte) (string, error) {
			code, libs, err := AsTerraformInput(theFunc, tfType, locMap, namer, schemas, spec)
			manager.Merge(libs)
			return code, err
		},
		"AsTerraformSaveState": func(theFunc *normalized.Function, tfType byte) (string, error) {
			code, libs, err := AsTerraformSaveState(theFunc, tfType, namer, modelName, ns.ShortName, "tfid", schemas, spec)
			manager.Merge(libs)
			return code, err
		},
		"AsInputFrom": func(item normalized.Item) (string, error) {
			code, libs, err := AsInputFrom(item, locMap)
			manager.Merge(libs)
			return code, err
		},
		"IsCreateOnlyParam": func(pp normalized.Item) (bool, error) {
			if ns.Create == nil {
				return false, fmt.Errorf("create is nil")
			}
			if ns.Update == nil {
				return false, fmt.Errorf("update is nil")
			}
			name := pp.GetInternalName()
			for _, cp := range ns.Update.PathParams {
				if cp.GetInternalName() == name {
					return false, nil
				}
			}
			for _, cp := range ns.Update.QueryParams {
				if cp.GetInternalName() == name {
					return false, nil
				}
			}
			return true, nil
		},
		"SaveParamUsingLocMap": func(i normalized.Item) (string, error) {
			code, libs, err := SaveParamUsingLocMap(i, "state", locMap)
			manager.Merge(libs)
			return code, err
		},
	}

	t := template.Must(
		template.New(
			fmt.Sprintf("terraform-resource-%s", tfName),
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
// Resource.
var (
    _ resource.Resource = &{{ StructName }}{}
    _ resource.ResourceWithConfigure = &{{ StructName }}{}
{{- if not HasEncryptedValues }}
    _ resource.ResourceWithImportState = &{{ StructName }}{}
{{- end }}

)

func {{ NewFuncName }}() resource.Resource {
    return &{{ StructName }}{}
}

type {{ StructName }} struct {
    client *{{ RepoShortName }}.Client
}

{{ AsTerraformModel }}

// Metadata returns the data source type name.
func (r *{{ StructName }}) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "{{ MetaName }}"
}

// Schema defines the schema for this data source.
func (r *{{ StructName }}) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = rsschema.Schema{
        Description: "Retrieves a config item.",

{{ AsTerraformSchema }}
    }
}

// Configure prepares the struct.
func (r *{{ StructName }}) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    r.client = req.ProviderData.(*scm.Client)
}

// Create resource.
{{- $create := CreateFunc }}
{{- ResetVarNaming }}
func (r *{{ StructName }}) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var state {{ ModelName }}
    resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }
{{- if HasEncryptedValues }}

    // Map containing encrypted and plain text values for encrypted params.
    ev := make(map[string] types.String)
{{- end }}

    // Basic logging.
    tflog.Info(ctx, "performing resource create", map[string] any{
        "resource_name": "{{ ProviderName }}{{ MetaName }}",
        "terraform_provider_function": "Create",
{{- range $item := $create.PathParams }}
{{ $item.TflogString }}
{{- end }}
{{- range $item := $create.QueryParams }}
{{ $item.TflogString }}
{{- end }}
{{- if ne $create.Request nil }}
{{ $create.Request.TflogString }}
{{- end }}
    })

    // Prepare to create the config.
    svc := {{ ServiceShortName }}.NewClient(r.client)
{{- if $create.FunctionHasInput }}

    // Prepare input for the API endpoint.
    input := {{ ServiceShortName }}.{{ $create.Name }}Input{}
{{ AsTerraformInput $create 'c' }}
{{- end }}

    // Perform the operation.
    ans, err := svc.{{ $create.Name }}(ctx
{{- if $create.FunctionHasInput }}, input
{{- end -}}
    )
    if err != nil {
        resp.Diagnostics.AddError("Error creating config", err.Error())
        return
    }

{{ AsTerraformId }}

    // Store the answer to state.
{{ AsTerraformSaveState $create 'c' }}
{{- if HasEncryptedValues }}
{{- $v1 := NewVarName }}
{{- $v2 := NewVarName }}

    {{ $v1 }}, {{ $v2 }} := types.MapValueFrom(ctx, types.StringType, ev)
    state.{{ Evccn }} = {{ $v1 }}
    resp.Diagnostics.Append({{ $v2 }}.Errors()...)
{{- end }}

    // Done.
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read performs Read for the struct.
{{- $read := ReadFunc }}
{{- ResetVarNaming }}
func (r *{{ StructName }}) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var savestate, state {{ ModelName }}
    resp.Diagnostics.Append(req.State.Get(ctx, &savestate)...)
    if resp.Diagnostics.HasError() {
        return
    }

    tfid := savestate.Tfid.ValueString()
    tokens := strings.Split(tfid, IdSeparator)
    if len(tokens) != {{ len LocMap }} {
        resp.Diagnostics.AddError("Error in resource ID format", "Expected {{ len LocMap }} tokens")
        return
    }
{{- if HasEncryptedValues }}

    // Map containing encrypted and plain text values for encrypted params.
    ev := make(map[string] types.String, len(savestate.{{ Evccn }}.Elements()))
    resp.Diagnostics.Append(savestate.{{ Evccn }}.ElementsAs(ctx, &ev, false).Errors()...)
    if resp.Diagnostics.HasError() {
        return
    }
{{- end }}

    // Basic logging.
    tflog.Info(ctx, "performing resource read", map[string] any{
        "terraform_provider_function": "Read",
        "resource_name": "{{ ProviderName }}{{ MetaName }}",
        "locMap": {{ printf "%#v" LocMap }},
        "tokens": tokens,
    })

    // Prepare to read the config.
    svc := {{ ServiceShortName }}.NewClient(r.client)
{{- if $read.FunctionHasInput }}

    // Prepare input for the API endpoint.
    input := {{ ServiceShortName }}.{{ $read.Name }}Input{}
{{- range $pp := $read.PathParams }}
{{ AsInputFrom $pp }}
{{- end }}
{{- range $pp := $read.QueryParams }}
{{ AsInputFrom $pp }}
{{- end }}
{{- end }}

    // Perform the operation.
    ans, err := svc.{{ $read.Name }}(ctx
{{- if .FunctionHasInput }}, input
{{- end -}}
    )
    if err != nil {
        if IsObjectNotFound(err) {
            resp.State.RemoveResource(ctx)
        } else {
            resp.Diagnostics.AddError("Error reading config", err.Error())
        }
        return
    }

    // Store the answer to state.
{{- range $pp := $create.PathParams }}
{{ SaveParamUsingLocMap $pp }}
{{- end }}
{{- range $pp := $create.QueryParams }}
{{ SaveParamUsingLocMap $pp }}
{{- end }}
    state.Tfid = savestate.Tfid
{{ AsTerraformSaveState $read 'r' }}
{{- if HasEncryptedValues }}
{{- $v1 := NewVarName }}
{{- $v2 := NewVarName }}

    {{ $v1 }}, {{ $v2 }} := types.MapValueFrom(ctx, types.StringType, ev)
    state.{{ Evccn }} = {{ $v1 }}
    resp.Diagnostics.Append({{ $v2 }}.Errors()...)
{{- end }}

    // Done.
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update performs the Update for the struct.
{{- $update := UpdateFunc }}
{{- ResetVarNaming }}
func (r *{{ StructName }}) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state {{ ModelName }}
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    tfid := state.Tfid.ValueString()
    tokens := strings.Split(tfid, IdSeparator)
    if len(tokens) != {{ len LocMap }} {
        resp.Diagnostics.AddError("Error in resource ID format", "Expected {{ len LocMap }} tokens")
        return
    }
{{- if HasEncryptedValues }}

    // Map containing encrypted and plain text values for encrypted params.
    ev := make(map[string] types.String, len(state.{{ Evccn }}.Elements()))
    resp.Diagnostics.Append(state.{{ Evccn }}.ElementsAs(ctx, &ev, false).Errors()...)
    if resp.Diagnostics.HasError() {
        return
    }
{{- end }}

    // Basic logging.
    tflog.Info(ctx, "performing resource update", map[string] any{
        "terraform_provider_function": "Update",
        "resource_name": "{{ ProviderName }}{{ MetaName }}",
        "tfid": state.Tfid.ValueString(),
    })

    // Prepare to update the config.
    svc := {{ ServiceShortName }}.NewClient(r.client)
{{- if $update.FunctionHasInput }}

    // Prepare input for the API endpoint.
    input := {{ ServiceShortName }}.{{ $update.Name }}Input{}
{{ AsTerraformInput $update 'u' }}
{{- end }}

    // Perform the operation.
    ans, err := svc.{{ $update.Name }}(ctx
{{- if $update.FunctionHasInput }}, input
{{- end -}}
    )
    if err != nil {
        if IsObjectNotFound(err) {
            resp.State.RemoveResource(ctx)
        } else {
            resp.Diagnostics.AddError("Error updating resource", err.Error())
        }
        return
    }

    // Store the answer to state.
    // Note: when supporting importing a resource, this will need to change to taking
    // values from the savestate.Tfid param and locMap.
{{ AsTerraformSaveState $update 'u' }}
{{- if HasEncryptedValues }}
{{- $v1 := NewVarName }}
{{- $v2 := NewVarName }}

    {{ $v1 }}, {{ $v2 }} := types.MapValueFrom(ctx, types.StringType, ev)
    state.{{ Evccn }} = {{ $v1 }}
    resp.Diagnostics.Append({{ $v2 }}.Errors()...)
{{- end }}

    // Done.
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete performs delete for the struct.
{{- $delete := DeleteFunc }}
{{- ResetVarNaming }}
func (r *{{ StructName }}) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var idType types.String
    resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tfid"), &idType)...)
    if resp.Diagnostics.HasError() {
        return
    }
    tfid := idType.ValueString()
    tokens := strings.Split(tfid, IdSeparator)
    if len(tokens) != {{ len LocMap }} {
        resp.Diagnostics.AddError("Error in resource ID format", "Expected {{ len LocMap }} tokens")
        return
    }

    // Basic logging.
    tflog.Info(ctx, "performing resource delete", map[string] any{
        "terraform_provider_function": "Delete",
        "resource_name": "{{ ProviderName }}{{ MetaName }}",
        "locMap": {{ printf "%#v" LocMap }},
        "tokens": tokens,
    })

    svc := {{ ServiceShortName }}.NewClient(r.client)
{{- if $delete.FunctionHasInput }}

    // Prepare input for the API endpoint.
    input := {{ ServiceShortName }}.{{ $delete.Name }}Input{}
{{- range $pp := $delete.PathParams }}
{{ AsInputFrom $pp }}
{{- end }}
{{- range $pp := $delete.QueryParams }}
{{ AsInputFrom $pp }}
{{- end }}
{{- end }}

    // Perform the operation.
    if _, err := svc.{{ $delete.Name }}(ctx
{{- if $delete.FunctionHasInput }}, input
{{- end -}}
    ); err != nil && !IsObjectNotFound(err) {
        resp.Diagnostics.AddError("Error in delete", err.Error())
    }
}
{{- if not HasEncryptedValues }}

func (r *{{ StructName }}) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("tfid"), req, resp)
}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, ns.Read)

	return newFuncName, b.String(), manager, err
}

func AsTerraformModel(fn *normalized.Function, tfType byte, prefix, defaultShortName, tfid string, schemas map[string]normalized.Item) (string, *imports.Manager, error) {
	merged, others, inputs, outputs, _, hasEncryptedValues, err := MergeIO(fn, tfType, tfid, schemas)
	if err != nil {
		return "", nil, err
	}

	if merged == nil {
		return "", nil, fmt.Errorf("merged is nil")
	} else if inputs == nil {
		return "", nil, fmt.Errorf("inputs is nil")
	} else if outputs == nil {
		return "", nil, fmt.Errorf("outputs is nil")
	}

	manager := imports.NewManager()

	fm := template.FuncMap{
		"Tfid":               func() string { return tfid },
		"AsModelLine":        func(i normalized.Item) (string, error) { return AsModelLine(i, prefix, defaultShortName, schemas) },
		"Prefix":             func() string { return prefix },
		"Merged":             func() *normalized.Object { return merged },
		"Others":             func() []*normalized.Object { return others },
		"HasEncryptedValues": func() bool { return hasEncryptedValues },
		"Evccn":              func() string { return EncryptedValuesCamelCaseName },
		"Evun":               func() string { return EncryptedValuesUnderscoreName },
		"Inputs": func() []string {
			ans := make([]string, 0, len(merged.Params))
			for _, name := range merged.OrderedParams(1) {
				if inputs[name] {
					ans = append(ans, name)
				}
			}
			return ans
		},
		"Outputs": func() []string {
			ans := make([]string, 0, len(merged.Params))
			for _, name := range merged.OrderedParams(1) {
				if outputs[name] {
					ans = append(ans, name)
				}
			}
			return ans
		},
		"IsAlsoInput": func(i normalized.Item) bool {
			_, ok := inputs[i.GetInternalName()]
			return ok
		},
		"ObjectModelName": func(cls *normalized.Object) (string, error) {
			return cls.TerraformModelType(prefix, defaultShortName, schemas)
		},
	}

	t := template.Must(
		template.New(
			"function-to-model",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $prefix := Prefix }}
{{- $merged := Merged }}
// {{ Prefix }} {{ $merged.Description }}
type {{ Prefix }} struct {
{{ AsModelLine (index $merged.Params Tfid) }}

    // Input.
{{- range $pname := Inputs }}
{{- $pp := index $merged.Params $pname }}
{{ AsModelLine $pp }}
{{- end }}

    // Output.
{{- if HasEncryptedValues }}
    {{ Evccn }} types.Map ` + "`tfsdk:\"{{ Evun }}\"`" + `
{{- end }}
{{- range $pname := Outputs }}
{{- $pp := index $merged.Params $pname }}
{{- if eq $pname "tfid" }}
{{- else if IsAlsoInput $pp }}
    // omit input: {{ $pp.GetUnderscoreName }}
{{- else }}
{{ AsModelLine $pp }}
{{- end }}
{{- end }}
}
{{- range $cls := Others }}
{{- end }}
{{- range $cls := Others }}

type {{ ObjectModelName $cls }} struct {
{{- range $pname := $cls.OrderedParams 1 }}
{{- $pp := index $cls.Params $pname }}
{{ AsModelLine $pp }}
{{- end }}
}
{{- end }}
{{- /* End */ -}}`,
		),
	)

	var b strings.Builder
	err = t.Execute(&b, nil)

	return b.String(), manager, err
}

func AsTerraformSchema(fn *normalized.Function, schemaPrefix, tfid string, suffixes map[string]string, schemas map[string]normalized.Item) (string, *imports.Manager, error) {
	var tfType byte
	switch schemaPrefix {
	case "dsschema":
		tfType = 'd'
	case "rsschema":
		tfType = 'r'
	}

	merged, _, inputs, outputs, forceNew, hasEncryptedValues, err := MergeIO(fn, tfType, tfid, schemas)
	if err != nil {
		return "", nil, err
	}

	evs := ""
	if hasEncryptedValues {
		evs = EncryptedValueSchema
	}

	return merged.AsTerraformSchema(schemaPrefix, evs, inputs, outputs, forceNew, suffixes, schemas)
}

func MergeIO(fn *normalized.Function, tfType byte, tfid string, schemas map[string]normalized.Item) (*normalized.Object, []*normalized.Object, map[string]bool, map[string]bool, map[string]bool, bool, error) {
	if fn == nil {
		return nil, nil, nil, nil, nil, false, fmt.Errorf("func is nil")
	} else if tfType != 'd' && tfType != 'r' {
		return nil, nil, nil, nil, nil, false, fmt.Errorf("tfType must be 'd' or 'r'")
	}

	rot := true
	merged := &normalized.Object{
		Description: "is the model.",
		ClassName:   "",
		//Params: make(map[string] normalized.Item),
		Params: map[string]normalized.Item{
			tfid: &normalized.String{
				Name:           tfid,
				ReadOnly:       &rot,
				Description:    "The Terraform ID.",
				UnderscoreName: tfid,
				CamelCaseName:  naming.CamelCase("", tfid, "", true),
			},
		},
	}

	forceNew := make(map[string]bool)
	inputs := make(map[string]bool)
	outputs := map[string]bool{
		tfid: true,
	}

	for _, p := range fn.PathParams {
		name := p.GetInternalName()
		if _, ok := merged.Params[name]; ok {
			return nil, nil, nil, nil, nil, false, fmt.Errorf("path param %q already in merged", name)
		}
		merged.Params[name] = p
		inputs[name] = true
		forceNew[name] = true
	}

	for _, p := range fn.QueryParams {
		name := p.GetInternalName()
		if _, ok := merged.Params[name]; ok {
			return nil, nil, nil, nil, nil, false, fmt.Errorf("path param %q already in merged", name)
		}
		merged.Params[name] = p
		inputs[name] = true
		forceNew[name] = true
	}

	otherMap := make(map[string]bool)
	others := make([]*normalized.Object, 0, 10)

	if fn.Request != nil {
		switch x := fn.Request.(type) {
		case *normalized.Object:
			merged.OneOf = append([]string(nil), x.OneOf...)

			list, err := x.GetObjects(schemas)
			if err != nil {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("Err in nsf.Input.GetObjects: %s", err)
			} else if len(list) == 0 {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("nsf.Input.GetObjects has 0 len")
			}

			// Merge the params, which should all be unique.
			for name, p := range list[0].Params {
				if _, ok := merged.Params[name]; ok {
					return nil, nil, nil, nil, nil, false, fmt.Errorf("body param %q already in merged", name)
				}
				merged.Params[name] = p
				inputs[name] = true
			}

			// Add the unique other classes found.
			for _, oo := range list[1:] {
				otherKey := fmt.Sprintf("%s.%s", oo.GetShortName(), oo.ClassName)
				if !otherMap[otherKey] {
					otherMap[otherKey] = true
					others = append(others, oo)
				}
			}
		default:
			// NOTE: I'm intentionally not resolving any ref here.  Sure hope a
			// a spec doesn't refer to a string.
			name := fn.Request.GetInternalName()
			if _, ok := merged.Params[name]; ok {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("default request body %q already in merged", name)
			}
			merged.Params[name] = fn.Request
			inputs[name] = true
			list, err := fn.Request.GetObjects(schemas)
			if err != nil {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("default request body %q.GetObjects err: %s", name, err)
			}
			others = append(others, list...)
		}
	}

	if fn.Request != nil && fn.Output != nil && fn.Request.GetReference() != "" && fn.Request.GetReference() == fn.Output.GetReference() {
		// If the function passed in is Create then don't add anything.
		// NOTE: If / when the query params are moved to the request body,
		// this logic will need to account for that.
		for name := range inputs {
			if !forceNew[name] {
				outputs[name] = true
			}
		}
	} else if fn.Output != nil {
		switch x := fn.Output.(type) {
		case *normalized.Object:
			list, err := x.GetObjects(schemas)
			if err != nil {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("Err in nsf.Output.GetObjects: %s", err)
			} else if len(list) == 0 {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("nsf.Output.GetObjects has 0 len")
			}

			// Merge the params.
			for name, p := range list[0].Params {
				if _, ok := merged.Params[name]; !ok {
					merged.Params[name] = p
				}
				outputs[name] = true
			}

			// Only add the class if it doesn't already exist.
			for _, oo := range list[1:] {
				otherKey := fmt.Sprintf("%s.%s", oo.GetShortName(), oo.ClassName)
				if !otherMap[otherKey] {
					otherMap[otherKey] = true
					others = append(others, oo)
				}
			}
		default:
			// NOTE: Again, intentionally not resolved any refs here, since it
			// is not an object...
			name := fn.Request.GetInternalName()
			if _, ok := merged.Params[name]; !ok {
				merged.Params[name] = fn.Request
			}
			outputs[name] = true
			list, err := fn.Request.GetObjects(schemas)
			if err != nil {
				return nil, nil, nil, nil, nil, false, fmt.Errorf("default output %q.GetObjects err: %s", name, err)
			}
			others = append(others, list...)
		}
	}

	hasEncryptedValues := false
	if tfType == 'r' && fn.Request != nil {
		encAns, err := fn.Request.HasEncryptedItems(schemas)
		if err != nil {
			return nil, nil, nil, nil, nil, false, fmt.Errorf("enc check err: %s", err)
		}
		if _, ok := merged.Params[EncryptedValuesUnderscoreName]; encAns && ok {
			return nil, nil, nil, nil, nil, false, fmt.Errorf("enc check err: %q already exists in merged params", EncryptedValuesUnderscoreName)
		}
		hasEncryptedValues = encAns
	}

	return merged, others, inputs, outputs, forceNew, hasEncryptedValues, nil
}

func AsInputFromTerraformId(fn *normalized.Function, locMap map[string]int, schemas map[string]normalized.Item) (string, *imports.Manager, error) {
	if fn == nil {
		return "", nil, fmt.Errorf("as input from terraform id fn is nil")
	} else if !fn.FunctionHasInput() {
		return "", nil, fmt.Errorf("as_input_from_terraform_id fn has no input")
	} else if fn.Parent == nil {
		return "", nil, fmt.Errorf("fn.Parent is nil")
	}

	manager := imports.NewManager()

	fm := template.FuncMap{
		"ShortName": func() string { return fn.Parent.ShortName },
	}

	t := template.Must(
		template.New(
			"as-input-from-terraform-id",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $fn := . }}
    input := {{ ShortName }}.{{ $fn.Name }}Input{}
{{- range $pp := $fn.QueryParams }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, fn)

	return b.String(), manager, err
}

func AsInputFrom(item normalized.Item, locMap map[string]int) (string, *imports.Manager, error) {
	if item == nil {
		return "", nil, fmt.Errorf("item is nil")
	}

	manager := imports.NewManager()

	fm := template.FuncMap{
		"IsString": func() (bool, error) {
			switch item.(type) {
			case *normalized.String:
				return true, nil
			case *normalized.Bool:
				return false, nil
			case *normalized.Int:
				return false, nil
			case *normalized.Float:
				return false, nil
			}

			return false, fmt.Errorf("Unsupported item type: %T", item)
		},
		"NeedStrconv": func() error {
			manager.AddStandardImport("strconv", "")
			return nil
		},
		"StrconvFunction": func() (string, error) {
			switch item.(type) {
			case *normalized.Bool:
				return "ParseBool", nil
			case *normalized.Int:
				return "ParseInt", nil
			case *normalized.Float:
				return "ParseFloat", nil
			}

			return "", fmt.Errorf("Unsupported item type: %T", item)
		},
		"StrconvParams": func() (string, error) {
			switch item.(type) {
			case *normalized.Bool:
				return "", nil
			case *normalized.Int:
				return ", 10, 64", nil
			case *normalized.Float:
				return ", 64", nil
			}

			return "", fmt.Errorf("Unsupported item type: %T", item)
		},
		"IsPointer": func() string {
			if item.IsRequired() {
				return ""
			}
			return "&"
		},
		"FindToken": func() (string, error) {
			num, ok := locMap[item.GetInternalName()]
			if !ok {
				return "", fmt.Errorf("item %q not in locMap: %#v", item.GetInternalName(), locMap)
			}
			return fmt.Sprintf("tokens[%d]", num), nil
		},
	}

	t := template.Must(
		template.New(
			"as-input-from",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $item := . }}
{{- if IsString }}
    input.{{ $item.GetCamelCaseName }} = {{ if not $item.IsRequired }}&{{ end }}{{ FindToken }}
{{- else }}
{{- NeedStrconv }}
{{- if $item.IsRequired }}
    input.{{ $item.GetCamelCaseName }}, err = strconv.{{ StrconvFunction }}({{ FindToken }}{{ StrconvParams }})
    if err != nil {
        resp.Diagnostics.AddError("Error parsing ID param '{{ $item.GetUnderscoreName }}'", err.Error())
        return
    }
{{- else }}
    if {{ FindToken }} != "" {
        tokval, err := strconv.{{ StrconvFunction }}({{ FindToken }}{{ StrconvParams }})
        if err != nil {
            resp.Diagnostics.AddError("Error parsing ID param '{{ $item.GetUnderscoreName }}'", err.Error())
            return
        }
        input.{{ $item.GetCamelCaseName }} = &tokval
    }
{{- end }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, item)

	return b.String(), manager, err
}

func AsTerraformInput(fn *normalized.Function, tfType byte, locMap map[string]int, namer *naming.Namer, schemas map[string]normalized.Item, spec *properties.Normalization) (string, *imports.Manager, error) {
	if fn == nil {
		return "", nil, fmt.Errorf("asInput function is nil")
	} else if fn.Parent == nil {
		return "", nil, fmt.Errorf("fn.Parent is nil")
	} else if !fn.FunctionHasInput() {
		return "", nil, fmt.Errorf("asInput function does not have input")
	} else if tfType != 'x' && tfType != 'c' && tfType != 'u' {
		// tfType options:
		// x: no special handling
		// c: resource create
		// u: resource update
		return "", nil, fmt.Errorf("tfType must be 'x', 'c', or 'u'")
	}

	// If this is true then we need to save encrypted values to the "ev" map.
	saveEncryptedValues := false
	if tfType != 'x' && fn.Request != nil {
		hev, err := fn.Request.HasEncryptedItems(schemas)
		if err != nil {
			return "", nil, err
		}
		saveEncryptedValues = hev
	}

	var sourceVariable string
	switch tfType {
	case 'u':
		sourceVariable = "plan"
	default:
		sourceVariable = "state"
	}

	shortName := fn.Parent.ShortName
	manager := imports.NewManager()
	manager.AddSdkImport(fn.Parent.Path(spec.Repository, spec.Name), fn.Parent.ShortName)

	fm := template.FuncMap{
		"ShortName":      func() string { return shortName },
		"HasRequestBody": func() bool { return fn.Request != nil },
		"Request":        func() normalized.Item { return fn.Request },
		"RequestIsObject": func() bool {
			_, ok := fn.Request.(*normalized.Object)
			return ok
		},
		"RequestObjectInit": func() (string, error) {
			if fn.Request == nil {
				return "", fmt.Errorf("fn.Request is nil")
			}
			robj, ok := fn.Request.(*normalized.Object)
			if !ok {
				return "", fmt.Errorf("fn.Request is not an object")
			}
			if robj.Reference == "" {
				return fmt.Sprintf("%s.%s", shortName, robj.ClassName), nil
			}
			lo, err := normalized.ItemLookup(robj, schemas)
			if err != nil {
				return "", fmt.Errorf("Err lookuping up object: %s", err)
			}
			lobj, ok := lo.(*normalized.Object)
			if !ok {
				return "", fmt.Errorf("obj lookup returned non-object: %T", lo)
			}
			AddItemToManager(lobj, spec.Repository, spec.Name, manager)
			return fmt.Sprintf("%s.%s", lobj.ShortName, lobj.ClassName), nil
		},
		"RequestAsObject": func() (*normalized.Object, error) {
			lo, err := normalized.ItemLookup(fn.Request, schemas)
			if err != nil {
				return nil, err
			}
			obj, ok := lo.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("request is not an object: %T", fn.Request)
			}
			return obj, nil
		},
		"ShouldIncludeParam": func(obj *normalized.Object, pname string) (bool, error) {
			if obj == nil {
				return false, fmt.Errorf("obj is nil")
			}
			param, ok := obj.Params[pname]
			if !ok {
				return false, fmt.Errorf("Param %q not in obj.Params", pname)
			}
			return !param.IsReadOnly(), nil
		},
		"RequestParamAsInput": func(obj *normalized.Object, pname string) (string, error) {
			param, ok := obj.Params[pname]
			if !ok {
				return "", fmt.Errorf("param %q not in request obj", pname)
			}
			code, libs, err := ItemAsInput(param, saveEncryptedValues, shortName, sourceVariable, "input.Request", nil, namer, schemas, spec)
			manager.Merge(libs)
			return code, err
		},
		"ItemAsInput": func(theitem normalized.Item, isReq bool) (string, error) {
			dst := "input"
			locMapParam := locMap
			if isReq {
				dst += ".Request"
				locMapParam = nil
			}
			code, libs, err := ItemAsInput(theitem, saveEncryptedValues, shortName, sourceVariable, dst, locMapParam, namer, schemas, spec)
			manager.Merge(libs)
			return code, err
		},
	}

	t := template.Must(
		template.New(
			"as-input",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $fn := . }}
{{- range $pp := $fn.PathParams }}
{{ ItemAsInput $pp false }}
{{- end }}
{{- range $pp := $fn.QueryParams }}
{{ ItemAsInput $pp false }}
{{- end }}
{{- if HasRequestBody }}
{{- if RequestIsObject }}
    input.Request = &{{ RequestObjectInit }}{}
{{- $obj := RequestAsObject }}
{{- range $pname := $obj.OrderedParams 1 }}
{{- if ShouldIncludeParam $obj $pname }}
{{ RequestParamAsInput $obj $pname }}
{{- end }}
{{- end }}
{{- else }}
{{ ItemAsInput Request true }}
{{- end }}
{{- end }}
{{- /* End */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, fn)

	return b.String(), manager, err
}

func AsTerraformSaveState(fn *normalized.Function, tfType byte, namer *naming.Namer, modelName, defaultShortName, tfid string, schemas map[string]normalized.Item, spec *properties.Normalization) (string, *imports.Manager, error) {
	if fn == nil {
		return "", nil, fmt.Errorf("fn is nil")
	} else if fn.Parent == nil {
		return "", nil, fmt.Errorf("fn.Parent is nil")
	} else if fn.Output == nil {
		return "", nil, nil
	} else if tfType != 'x' && tfType != 'c' && tfType != 'r' && tfType != 'u' {
		// 'x': data source
		// 'c': resource create
		// 'r': resource read
		// 'u': resource update
		return "", nil, fmt.Errorf("tfType must be 'x', 'c', 'r', or 'u'")
	}

	manager := imports.NewManager()

	fm := template.FuncMap{
		"Tfid":      func() string { return naming.CamelCase("", tfid, "", true) },
		"TfType":    func() byte { return tfType },
		"HasOutput": func() bool { return fn.Output != nil },
		"OutputIsObject": func() bool {
			_, ok := fn.Output.(*normalized.Object)
			return ok
		},
		"OutputAsObject": func() (*normalized.Object, error) {
			lo, err := normalized.ItemLookup(fn.Output, schemas)
			if err != nil {
				return nil, err
			}
			obj, ok := lo.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("Output is %T, not normalized.Object", lo)
			}
			return obj, nil
		},
		"OutputParamAsSaveState": func(obj *normalized.Object, pname string) (string, error) {
			if obj == nil {
				return "", fmt.Errorf("obj is nil")
			}
			param, ok := obj.Params[pname]
			if !ok {
				return "", fmt.Errorf("Param %q not in obj.Params", pname)
			}
			return ItemAsSaveState(param, tfType, modelName, defaultShortName, "ans", "state", namer, schemas, spec)
		},
		"OutputAsSaveState": func() (string, error) {
			lo, err := normalized.ItemLookup(fn.Output, schemas)
			if err != nil {
				return "", err
			}
			return ItemAsSaveState(lo, tfType, modelName, defaultShortName, "ans", "state", namer, schemas, spec)
		},
	}

	t := template.Must(
		template.New(
			"as-terraform-save-state",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- if or (eq TfType 'x') (eq TfType 'c') }}
    state.{{ Tfid }} = types.StringValue(idBuilder.String())
{{- end }}
{{- if HasOutput }}
{{- if OutputIsObject }}
{{- $obj := OutputAsObject }}
{{- range $pname := $obj.OrderedParams 1 }}
{{ OutputParamAsSaveState $obj $pname }}
{{- end }}
{{- else }}
{{ OutputAsSaveState }}
{{- end }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, nil)

	return b.String(), manager, err
}

func ItemAsSaveState(i normalized.Item, tfType byte, modelName, defaultShortName, src, dst string, namer *naming.Namer, schemas map[string]normalized.Item, spec *properties.Normalization) (string, error) {
	v, err := normalized.ItemLookup(i, schemas)
	if err != nil {
		return "", err
	} else if tfType != 'x' && tfType != 'c' && tfType != 'r' && tfType != 'u' {
		return "", fmt.Errorf("invalid tfType: %s", tfType)
	}

	fm := template.FuncMap{
		"Name": func() string { return v.GetCamelCaseName() },
		"ShouldProcessAsEncryptedCreateOrUpdate": func() bool {
			s, ok := v.(*normalized.String)
			if !ok {
				return false
			}
			if tfType != 'c' && tfType != 'u' {
				return false
			}
			return s.IsEncrypted()
		},
		"ShouldProcessAsEncryptedRead": func() bool {
			s, ok := v.(*normalized.String)
			if !ok {
				return false
			}
			if tfType != 'r' {
				return false
			}
			return s.IsEncrypted()
		},
		"SaveEncryptionKeyTo": func(varName string, srcIsTfType bool, keyType byte) (string, error) {
			return v.GetEncryptionKey(src, varName, srcIsTfType, keyType)
		},
		"ModelName":   func() string { return modelName },
		"Source":      func() string { return fmt.Sprintf("%s.%s", src, i.GetCamelCaseName()) },
		"Destination": func() string { return fmt.Sprintf("%s.%s", dst, i.GetCamelCaseName()) },
		"NewVarName":  func() string { return namer.NextVarName() },
		"IsRegularBool": func() bool {
			x, ok := v.(*normalized.Bool)
			return ok && (x.IsObjectBool == nil || !*x.IsObjectBool)
		},
		"IsObjectBool": func() bool {
			x, ok := v.(*normalized.Bool)
			return ok && (x.IsObjectBool != nil && *x.IsObjectBool)
		},
		"IsInt": func() bool {
			_, ok := v.(*normalized.Int)
			return ok
		},
		"IsFloat": func() bool {
			_, ok := v.(*normalized.Float)
			return ok
		},
		"IsString": func() bool {
			_, ok := v.(*normalized.String)
			return ok
		},
		"IsArray": func() bool {
			_, ok := v.(*normalized.Array)
			return ok
		},
		"IsObject": func() bool {
			_, ok := v.(*normalized.Object)
			return ok
		},
		"IsUnsupported": func() error {
			return fmt.Errorf("Unsupported input type %T", v)
		},
		"AsObject": func() (*normalized.Object, error) {
			lans, err := normalized.ItemLookup(v, schemas)
			if err != nil {
				return nil, err
			}
			obj, ok := lans.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("lookup ans is not *Object: %T", lans)
			}
			return obj, nil
		},
		"ObjectParamAsSaveState": func(obj *normalized.Object, pname string) (string, error) {
			if obj == nil {
				return "", fmt.Errorf("obj is nil")
			}
			pi, ok := obj.Params[pname]
			if !ok {
				return "", fmt.Errorf("obj.Parans[%s] not found", pname)
			}
			suffix := i.GetCamelCaseName()
			osrc := fmt.Sprintf("%s.%s", src, suffix)
			odst := fmt.Sprintf("%s.%s", dst, suffix)
			return ItemAsSaveState(pi, tfType, modelName, defaultShortName, osrc, odst, namer, schemas, spec)
		},
		"ArraySpecAsObject": func(spec normalized.Item) (*normalized.Object, error) {
			lo, err := normalized.ItemLookup(spec, schemas)
			if err != nil {
				return nil, err
			}
			obj, ok := lo.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("Array spec is not an object: %T", lo)
			}
			return obj, nil
		},
		"ArraySpecAsSaveState": func(item normalized.Item, ssrc, sdst string) (string, error) {
			if item == nil {
				return "", fmt.Errorf("array spec is nil")
			}
			return ItemAsSaveState(item, tfType, modelName, defaultShortName, ssrc, sdst, namer, schemas, spec)
		},
		"BasicSpecFunc": func(spec normalized.Item) (string, error) {
			if spec == nil {
				return "", fmt.Errorf("spec is nil")
			}
			switch x := spec.(type) {
			case *normalized.Bool:
				if x.IsObjectBool != nil && *x.IsObjectBool {
					return "", fmt.Errorf("TODO: array of anys: %s : %#v", x.UnderscoreName, x.Path())
				}
				return "BoolValue", nil
			case *normalized.Int:
				return "Int64Value", nil
			case *normalized.Float:
				return "Float64Value", nil
			case *normalized.String:
				return "StringValue", nil
			}
			return "", fmt.Errorf("not a basic type: %T", spec)
		},
		"AsArray": func() (*normalized.Array, error) {
			lans, err := normalized.ItemLookup(v, schemas)
			if err != nil {
				return nil, err
			}
			arr, ok := lans.(*normalized.Array)
			if !ok {
				return nil, fmt.Errorf("lookup ans is not *Array: %T", lans)
			}
			return arr, nil
		},
		"AsArraySpec": func(arr *normalized.Array) (normalized.Item, error) {
			if arr == nil {
				return nil, fmt.Errorf("array is nil")
			}
			spec, err := normalized.ItemLookup(arr.Spec, schemas)
			if err != nil {
				return nil, err
			}
			return spec, nil
		},
		"ArraySpecIsBasicType": func(spec normalized.Item) (bool, error) {
			if spec == nil {
				return false, fmt.Errorf("spec is nil")
			}
			switch spec.(type) {
			case *normalized.Bool:
				return true, nil
			case *normalized.Int:
				return true, nil
			case *normalized.Float:
				return true, nil
			case *normalized.String:
				return true, nil
			case *normalized.Array:
				return false, fmt.Errorf("TODO: array of arrays")
			case *normalized.Object:
				return false, nil
			}
			return false, fmt.Errorf("unsupported spec type: %T", spec)
		},
		"ArraySpecBasicType": func(spec normalized.Item) (string, error) {
			if spec == nil {
				return "", fmt.Errorf("spec is nil")
			}
			switch spec.(type) {
			case *normalized.Bool:
				return "types.BoolType", nil
			case *normalized.Int:
				return "types.Int64Type", nil
			case *normalized.String:
				return "types.StringType", nil
			case *normalized.Float:
				// NOTE: for some reason, the ListValueFrom() seems to have problems
				// with a list of float64s, so for right now just error out so people
				// don't think it's our problem.
				return "", fmt.Errorf("TODO: []float64 hasn't worked in my testing")
			}
			return "", fmt.Errorf("type %T is not a basic type", spec)
		},
		"ArraySpecIsObjectType": func(spec normalized.Item) (bool, error) {
			if spec == nil {
				return false, fmt.Errorf("spec is nil")
			}
			_, ok := spec.(*normalized.Object)
			return ok, nil
		},
		"ArraySpecType": func(spec normalized.Item) (string, error) {
			if spec == nil {
				return "", fmt.Errorf("array spec prefix item is nil")
			}

			return spec.TerraformModelType(modelName, defaultShortName, schemas)
		},
		"UnsupportedArraySpec": func(spec normalized.Item) error {
			return fmt.Errorf("unsupported array spec type: %T", spec)
		},
		"ToTerraformModelName": func(obj *normalized.Object) (string, error) {
			return obj.TerraformModelType(modelName, defaultShortName, schemas)
		},
	}

	t := template.Must(
		template.New(
			"item-as-save-state",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $pp := . }}
{{- if IsRegularBool }}
    {{ Destination }} = types.Bool{{ if not $pp.IsRequired }}Pointer{{ end }}Value({{ Source }})
{{- else if IsObjectBool }}
    if {{ Source }} != nil {
        {{ Destination }} = types.BoolValue(true)
    } else {
        {{ Destination }} = types.BoolPointerValue(nil)
    }
    //{{ Destination }} = types.BoolValue({{ Source }} != nil)
{{- else if IsInt }}
    {{ Destination }} = types.Int64{{ if not $pp.IsRequired }}Pointer{{ end }}Value({{ Source }})
{{- else if IsFloat }}
    {{ Destination }} = types.Float64{{ if not $pp.IsRequired }}Pointer{{ end }}Value({{ Source }})
{{- else if IsString }}
{{- if ShouldProcessAsEncryptedCreateOrUpdate }}
{{- $encKey := NewVarName }}
{{- $ptKey := NewVarName }}
{{ SaveEncryptionKeyTo $encKey false 'e' }}
    ev[{{ $encKey }}] = types.String{{ if not $pp.IsRequired }}Pointer{{ end }}Value({{ Source }})
{{ SaveEncryptionKeyTo $ptKey false 'p' }}
    {{ Destination }} = ev[{{ $ptKey }}]
{{- else if ShouldProcessAsEncryptedRead }}
{{- $encKey := NewVarName }}
{{- $ptKey := NewVarName }}
{{ SaveEncryptionKeyTo $encKey false 'e' }}
    if ev[{{ $encKey }}].Equal(types.String{{ if not $pp.IsRequired }}Pointer{{ end }}Value({{ Source }})) {
{{ SaveEncryptionKeyTo $ptKey false 'p' }}
        {{ Destination }} = ev[{{ $ptKey }}]
    } else {
        {{ Destination }} = types.StringNull()
    }
{{- else }}
    {{ Destination }} = types.String{{ if not $pp.IsRequired }}Pointer{{ end }}Value({{ Source }})
{{- end }}
{{- else if IsObject }}
{{- $obj := AsObject }}
{{- if $pp.IsRequired }}
    {{ Destination }} = {{ ToTerraformModelName $obj }}{}
{{- range $pname := $obj.OrderedParams 1 }}
{{ ObjectParamAsSaveState $obj $pname }}
{{- end }}
{{- else }}
    if {{ Source }} == nil {
        {{ Destination }} = nil
    } else {
        {{ Destination }} = &{{ ToTerraformModelName $obj }}{}
{{- range $pname := $obj.OrderedParams 1 }}
{{ ObjectParamAsSaveState $obj $pname }}
{{- end }}
    }
{{- end }}
{{- else if IsArray }}
{{- $arr := AsArray }}
{{- $vname := NewVarName }}
{{- $vsub := NewVarName }}
{{- $spec := AsArraySpec $arr }}
{{- if ArraySpecIsBasicType $spec }}
    {{ $vname }}, {{ $vsub }} := types.ListValueFrom(ctx, {{ ArraySpecBasicType $spec }}, {{ Source }})
    {{ Destination }} = {{ $vname }}
    resp.Diagnostics.Append({{ $vsub }}.Errors()...)
{{- else if ArraySpecIsObjectType $spec }}
{{- $ospec := ArraySpecAsObject $spec }}
    if len({{ Source }}) == 0 {
        {{ Destination }} = nil
    } else {
        {{ Destination }} = make([]{{ ArraySpecType $spec }}, 0, len({{ Source }}))
        for _, {{ $vname }} := range {{ Source }} {
            {{ $vsub }} := {{ ArraySpecType $spec }}{}
{{- range $pname := $ospec.OrderedParams 1 }}
{{- $pp := index $ospec.Params $pname }}
{{ ArraySpecAsSaveState $pp $vname $vsub }}
{{- end }}
            {{ Destination }} = append({{ Destination }}, {{ $vsub }})
        }
    }
{{- else }}
{{ UnsupportedArraySpec $spec }}
{{- end }}
{{- else }}
{{ IsUnsupported }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err = t.Execute(&b, v)

	return b.String(), err
}

func ItemAsInput(i normalized.Item, saveEncryptedValues bool, defaultShortName, src, dst string, locMap map[string]int, namer *naming.Namer, schemas map[string]normalized.Item, spec *properties.Normalization) (string, *imports.Manager, error) {
	if i == nil {
		return "", nil, fmt.Errorf("itemAsInput item is nil")
	} else if spec == nil {
		return "", nil, fmt.Errorf("props is nil")
	} else if spec.Repository == "" {
		return "", nil, fmt.Errorf("spec.Repository is not defined")
	}

	v, err := normalized.ItemLookup(i, schemas)
	if err != nil {
		return "", nil, err
	}

	manager := imports.NewManager()
	AddItemToManager(v, spec.Repository, spec.Name, manager)

	fm := template.FuncMap{
		"Name":                     func() string { return v.GetCamelCaseName() },
		"ShouldSaveEncryptedValue": func() bool { return saveEncryptedValues && v.IsEncrypted() },
		"SaveEncryptionKeyTo":      func(varName string) (string, error) { return v.GetEncryptionKey(src, varName, true, 'p') },
		"IsToken": func() bool {
			_, ok := locMap[i.GetInternalName()]
			return ok
		},
		"TokenNumber": func() (int, error) {
			value, ok := locMap[i.GetInternalName()]
			if !ok {
				return 0, fmt.Errorf("name %q not in locMap", i.GetInternalName())
			}
			return value, nil
		},
		"TokensStrconv": func() (string, error) {
			num, ok := locMap[v.GetInternalName()]
			if !ok {
				return "", fmt.Errorf("name %q not in locMap", v.GetInternalName())
			}
			switch v.(type) {
			case *normalized.Bool:
				return fmt.Sprintf("ParseBool(tokens[%d])", num), nil
			case *normalized.Int:
				return fmt.Sprintf("ParseInt(tokens[%d], 10, 64)", num), nil
			case *normalized.Float:
				return fmt.Sprintf("ParseFloat(tokens[%d], 64)", num), nil
			}
			return "", fmt.Errorf("Unsupported token type: %T", v)
		},
		"TokensTypeFunction": func() (string, error) {
			switch v.(type) {
			case *normalized.Bool:
				return "Bool", nil
			case *normalized.Int:
				return "Int64", nil
			case *normalized.Float:
				return "Float64", nil
			case *normalized.String:
				return "String", nil
			}
			return "", fmt.Errorf("Unsupported token type: %T", v)
		},
		//"Source": func() string { return src },
		//"Destination": func() string { return dst },
		"Source":      func() string { return fmt.Sprintf("%s.%s", src, i.GetCamelCaseName()) },
		"Destination": func() string { return fmt.Sprintf("%s.%s", dst, i.GetCamelCaseName()) },
		"NewVarName":  func() string { return namer.NextVarName() },
		"IsRegularBool": func() bool {
			x, ok := v.(*normalized.Bool)
			return ok && (x.IsObjectBool == nil || !*x.IsObjectBool)
		},
		"IsObjectBool": func() bool {
			x, ok := v.(*normalized.Bool)
			return ok && (x.IsObjectBool != nil && *x.IsObjectBool)
		},
		"IsInt": func() bool {
			_, ok := v.(*normalized.Int)
			return ok
		},
		"IsFloat": func() bool {
			_, ok := v.(*normalized.Float)
			return ok
		},
		"IsString": func() bool {
			_, ok := v.(*normalized.String)
			return ok
		},
		"IsArray": func() bool {
			_, ok := v.(*normalized.Array)
			return ok
		},
		"IsObject": func() bool {
			_, ok := v.(*normalized.Object)
			return ok
		},
		"IsUnsupported": func() error {
			return fmt.Errorf("Unsupported input type %T", v)
		},
		"AsObject": func() (*normalized.Object, error) {
			lans, err := normalized.ItemLookup(v, schemas)
			if err != nil {
				return nil, err
			}
			obj, ok := lans.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("lookup ans is not *Object: %T", lans)
			}
			return obj, nil
		},
		"ShortName": func(obj *normalized.Object) (string, error) {
			if obj == nil {
				return "", fmt.Errorf("obj is nil")
			}
			sn := obj.GetShortName()
			if sn != "" {
				return sn, nil
			}
			return defaultShortName, nil
		},
		"ShouldIncludeParam": func(obj *normalized.Object, pname string) (bool, error) {
			if obj == nil {
				return false, fmt.Errorf("obj is nil")
			}
			param, ok := obj.Params[pname]
			if !ok {
				return false, fmt.Errorf("Param %q not in obj.Params", pname)
			}
			return !param.IsReadOnly(), nil
		},
		"ObjectParamAsInput": func(obj *normalized.Object, pname string) (string, error) {
			if obj == nil {
				return "", fmt.Errorf("obj is nil")
			}
			pi, ok := obj.Params[pname]
			if !ok {
				return "", fmt.Errorf("obj.Parans[%s] not found", pname)
			}
			suffix := i.GetCamelCaseName()
			osrc := fmt.Sprintf("%s.%s", src, suffix)
			odst := fmt.Sprintf("%s.%s", dst, suffix)
			code, libs, err := ItemAsInput(pi, saveEncryptedValues, defaultShortName, osrc, odst, nil, namer, schemas, spec)
			manager.Merge(libs)
			return code, err
		},
		"ArraySpecAsInput": func(item normalized.Item, ssrc, sdst string) (string, error) {
			if item == nil {
				return "", fmt.Errorf("array spec is nil")
			}
			_, ok := item.(*normalized.Object)
			if !ok {
				return "", fmt.Errorf("spec is not an object, but %T", item)
			}
			code, libs, err := ItemAsInput(item, saveEncryptedValues, defaultShortName, ssrc, sdst, nil, namer, schemas, spec)
			manager.Merge(libs)
			return code, err
		},
		"ArrayParamAsInput": func(ai normalized.Item, asrc, adst string) (string, error) {
			//psrc := fmt.Sprintf("%s.%s", asrc, ai.GetCamelCaseName())
			//pdst := fmt.Sprintf("%s.%s", adst, ai.GetCamelCaseName())
			code, libs, err := ItemAsInput(ai, saveEncryptedValues, defaultShortName, asrc, adst, nil, namer, schemas, spec)
			manager.Merge(libs)
			return code, err
		},
		"BasicSpecFunc": func(spec normalized.Item) (string, error) {
			if spec == nil {
				return "", fmt.Errorf("spec is nil")
			}
			switch x := spec.(type) {
			case *normalized.Bool:
				if x.IsObjectBool != nil && *x.IsObjectBool {
					return "", fmt.Errorf("TODO: array of any's")
				}
				return "ValueBool", nil
			case *normalized.Int:
				return "ValueInt64", nil
			case *normalized.Float:
				return "ValueFloat64", nil
			case *normalized.String:
				return "ValueString", nil
			}
			return "", fmt.Errorf("not a basic type: %T", spec)
		},
		"AsArray": func() (*normalized.Array, error) {
			lans, err := normalized.ItemLookup(v, schemas)
			if err != nil {
				return nil, err
			}
			arr, ok := lans.(*normalized.Array)
			if !ok {
				return nil, fmt.Errorf("lookup ans is not *Array: %T", lans)
			}
			return arr, nil
		},
		"AsArraySpec": func(arr *normalized.Array) (normalized.Item, error) {
			if arr == nil {
				return nil, fmt.Errorf("array is nil")
			}
			spec, err := normalized.ItemLookup(arr.Spec, schemas)
			if err != nil {
				return nil, err
			}
			return spec, nil
		},
		"ArraySpecIsBasicType": func(spec normalized.Item) (bool, error) {
			if spec == nil {
				return false, fmt.Errorf("spec is nil")
			}
			switch spec.(type) {
			case *normalized.Bool:
				return true, nil
			case *normalized.Int:
				return true, nil
			case *normalized.Float:
				return true, nil
			case *normalized.String:
				return true, nil
			case *normalized.Array:
				return false, fmt.Errorf("TODO: array of arrays")
			case *normalized.Object:
				return false, nil
			}
			return false, fmt.Errorf("unsupported spec type: %T", spec)
		},
		"ArraySpecIsObjectType": func(spec normalized.Item) (bool, error) {
			if spec == nil {
				return false, fmt.Errorf("spec is nil")
			}
			_, ok := spec.(*normalized.Object)
			return ok, nil
		},
		"ArraySpecAsObject": func(spec normalized.Item) (*normalized.Object, error) {
			if spec == nil {
				return nil, fmt.Errorf("spec is nil")
			}
			sl, err := normalized.ItemLookup(spec, schemas)
			if err != nil {
				return nil, err
			}
			sobj, ok := sl.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("array spec is not object: %T", sl)
			}
			return sobj, nil
		},
		"ArraySpecType": func(spec normalized.Item) (string, error) {
			if spec == nil {
				return "", fmt.Errorf("array spec prefix item is nil")
			}

			switch x := spec.(type) {
			case *normalized.Bool:
				if x.IsObjectBool != nil && *x.IsObjectBool {
					return "any", nil
				}
				return "bool", nil
			case *normalized.Int:
				return "int64", nil
			case *normalized.Float:
				return "float64", nil
			case *normalized.String:
				return "string", nil
			case *normalized.Object:
				return fmt.Sprintf("%s.%s", x.GetShortName(), x.ClassName), nil
			case *normalized.Array:
				return "", fmt.Errorf("TODO: array of arrays")
			}
			return "", fmt.Errorf("unsupported type passed in to array spec type: %T", spec)
		},
		"UnsupportedArraySpec": func(spec normalized.Item) error {
			return fmt.Errorf("unsupported array spec type: %T", spec)
		},
		"UnsupportedTokenType": func() error {
			return fmt.Errorf("unsupported token type: %T", v)
		},
	}

	t := template.Must(
		template.New(
			"item-as-input",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $pp := . }}
{{- if IsToken }}
{{- $tnum := TokenNumber }}
    if tokens[{{ $tnum }}] != "" {
{{- if IsString }}
        {{ Destination }} = {{ if not $pp.IsRequired }}&{{ end }}tokens[{{ $tnum }}]
{{- else }}
        x, err := strconv.{{ TokensStrconv }}
        if err != nil {
            resp.Diagnostics.AddError("Error parsing token", fmt.Sprintf("token:%d err:%s", {{ $tnum }}, err))
            return
        }
        {{ Destination }} = {{ if not $pp.IsRequired }}&{{ end }}x
{{- end }}
    }
{{- else if IsRegularBool }}
    {{ Destination }} = {{ Source }}.ValueBool{{ if not $pp.IsRequired }}Pointer{{ end }}()
{{- else if IsObjectBool }}
    if !{{ Source }}.IsNull() && {{ Source }}.ValueBool() {
        {{ Destination }} = map[string] interface{}{}
    }
{{- else if IsInt }}
    {{ Destination }} = {{ Source }}.ValueInt64{{ if not $pp.IsRequired }}Pointer{{ end }}()
{{- else if IsFloat }}
    {{ Destination }} = {{ Source }}.ValueFloat64{{ if not $pp.IsRequired }}Pointer{{ end }}()
{{- else if IsString }}
{{- if ShouldSaveEncryptedValue }}
{{- $ptKey := NewVarName }}
{{ SaveEncryptionKeyTo $ptKey }}
    ev[{{ $ptKey }}] = {{ Source }}
{{- end }}
    {{ Destination }} = {{ Source }}.ValueString{{ if not $pp.IsRequired }}Pointer{{ end }}()
{{- else if IsObject }}
{{- $obj := AsObject }}
{{- if $pp.IsRequired }}
{{- range $pname := $obj.OrderedParams 1 }}
{{- if ShouldIncludeParam $obj $pname }}
{{ ObjectParamAsInput $obj $pname }}
{{- end }}
{{- end }}
{{- else }}
    if {{ Source }} != nil {
        {{ Destination }} = &{{ ShortName $obj }}.{{ $obj.ClassName }}{}
{{- range $pname := $obj.OrderedParams 1 }}
{{- if ShouldIncludeParam $obj $pname }}
{{ ObjectParamAsInput $obj $pname }}
{{- end }}
{{- end }}
    }
{{- end }}
{{- else if IsArray }}
{{- $arr := AsArray }}
{{- $spec := AsArraySpec $arr }}
{{- $vname := NewVarName }}
{{- if ArraySpecIsBasicType $spec }}
    resp.Diagnostics.Append({{ Source }}.ElementsAs(ctx, &{{ Destination }}, false)...)
    //if len({{ Source }}) != 0 {
    //    {{ Destination }} = make([]{{ ArraySpecType $spec }}, 0, len({{ Source }}))
    //    for _, {{ $vname }} := range {{ Source }} {
    //        {{ Destination }} = append({{ Destination }}, {{ $vname }}.{{ BasicSpecFunc $spec }}())
    //    }
    //}
{{- else if ArraySpecIsObjectType $spec }}
{{- $specObj := ArraySpecAsObject $spec }}
{{- $vsub := NewVarName }}
    if len({{ Source }}) != 0 {
        {{ Destination }} = make([]{{ ArraySpecType $spec }}, 0, len({{ Source }}))
        for _, {{ $vname }} := range {{ Source }} {
            var {{ $vsub }} {{ ArraySpecType $spec }}
{{- range $pp := $specObj.Params }}
{{ ArrayParamAsInput $pp $vname $vsub }}
{{- end }}
            {{ Destination }} = append({{ Destination }}, {{ $vsub }})
        }
    }
{{- else }}
{{ UnsupportedArraySpec $spec }}
{{- end }}
{{- else }}
{{ IsUnsupported }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err = t.Execute(&b, v)

	return b.String(), manager, err
}

func AddParamToTerraformId(param normalized.Item, src string, mustBeDefined bool) (string, *imports.Manager, error) {
	manager := imports.NewManager()

	if _, ok := param.(*normalized.String); !ok {
		manager.AddStandardImport("strconv", "")
	}

	fm := template.FuncMap{
		"MustBeDefined": func() bool { return mustBeDefined },
		"Name":          func() string { return fmt.Sprintf("%s.%s", src, param.GetCamelCaseName()) },
		"IsInt": func() bool {
			_, ok := param.(*normalized.Int)
			return ok
		},
		"IsFloat": func() bool {
			_, ok := param.(*normalized.Float)
			return ok
		},
		"IsBool": func() bool {
			_, ok := param.(*normalized.Bool)
			return ok
		},
		"IsString": func() bool {
			_, ok := param.(*normalized.String)
			return ok
		},
		"IsUnsupported": func() error {
			return fmt.Errorf("Unsupported type cannot be in the ID: %T", param)
		},
	}

	t := template.Must(
		template.New(
			"add-param-to-terraform-id",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $pp := . }}
{{- if and (not $pp.IsRequired) MustBeDefined }}
    if {{ Name }} == nil {
        resp.Diagnostics.AddError("Undefined param required for ID", "{{ $pp.GetCamelCaseName }}")
        return
    }
{{- end }}
{{- if IsString }}
{{- if $pp.IsRequired }}
    idBuilder.WriteString({{ Name }})
{{- else }}
    if {{ Name }} != nil {
        idBuilder.WriteString(*{{ Name }})
    }
{{- end }}
{{- else }}
{{- if IsBool }}
{{- if $pp.IsRequired }}
    idBuilder.WriteString(strconv.FormatBool({{ Name }}))
{{- else }}
    if {{ Name }} != nil {
        idBuilder.WriteString(strconv.FormatBool(*{{ Name }}))
    }
{{- end }}
{{- else if IsInt }}
{{- if $pp.IsRequired }}
    idBuilder.WriteString(strconv.FormatInt({{ Name }}, 10))
{{- else }}
    if {{ Name }} != nil {
        idBuilder.WriteString(strconv.FormatInt(*{{ Name }}, 10))
    }
{{- end }}
{{- else if IsFloat }}
{{- if $pp.IsRequired }}
    idBuilder.WriteString(strconv.FormatFloat({{ Name }}, 'g', -1, 64))
{{- else }}
    if {{ Name }} != nil {
        idBuilder.WriteString(strconv.FormatFloat(*{{ Name }}, 'g', -1, 64))
    }
{{- end }}
{{- else }}
{{ IsUnsupported }}
{{- end }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, param)

	return b.String(), manager, err
}

func AsTerraformId(fin, fout *normalized.Function, schemas map[string]normalized.Item) (string, map[string]int, *imports.Manager, error) {
	if fin == nil {
		return "", nil, nil, fmt.Errorf("input func is nil")
	} else if fout == nil {
		return "", nil, nil, fmt.Errorf("output func is nil")
	}

	locMap := make(map[string]int)

	manager := imports.NewManager()
	manager.AddStandardImport("strings", "")

	fm := template.FuncMap{
		"Separator": func() string {
			if len(locMap) != 0 {
				return "\n\n    idBuilder.WriteString(IdSeparator)"
			}
			return ""
		},
		"AddParamToTerraformId": func(i normalized.Item, src string, mustBeDefined bool) (string, error) {
			locMap[i.GetInternalName()] = len(locMap)
			code, libs, err := AddParamToTerraformId(i, src, mustBeDefined)
			manager.Merge(libs)
			return code, err
		},
		"OutputHasMissingIdField": func() (bool, error) {
			if fout == nil {
				return false, fmt.Errorf("output func is nil")
			} else if fout.Output == nil {
				return false, fmt.Errorf("output func has nil output")
			}

			v, err := normalized.ItemLookup(fout.Output, schemas)
			if err != nil {
				return false, err
			}

			vobj, ok := v.(*normalized.Object)
			if !ok {
				return false, nil
			}

			checks := []string{"id", "uuid"}
			for _, chk := range checks {
				// The Read function will already include the id param, so if it's already
				// been added to the locMap we don't need to do it again.
				if _, present := locMap[chk]; present {
					continue
				}
				p, ok := vobj.Params[chk]
				if !ok {
					continue
				}
				if p.IsReadOnly() {
					return true, nil
				}
			}

			return false, nil
		},
		"GetIdParam": func() (normalized.Item, error) {
			if fout == nil {
				return nil, fmt.Errorf("output func is nil")
			} else if fout.Output == nil {
				return nil, fmt.Errorf("output func has nil output")
			}

			v, err := normalized.ItemLookup(fout.Output, schemas)
			if err != nil {
				return nil, err
			}

			vobj, ok := v.(*normalized.Object)
			if !ok {
				return nil, fmt.Errorf("not an object")
			}

			checks := []string{"id", "uuid"}
			for _, chk := range checks {
				if _, present := locMap[chk]; present {
					continue
				}
				p, ok := vobj.Params[chk]
				if !ok {
					continue
				}
				if p.IsReadOnly() {
					return p, nil
				}
			}

			return nil, fmt.Errorf("no suitable id param found")
		},
		"IncludeReadOnlyOutput": func() (string, error) {
			if fout == nil {
				return "", fmt.Errorf("output func is nil")
			} else if fout.Output == nil {
				return "", fmt.Errorf("output func output is nil")
			}

			v, err := normalized.ItemLookup(fout.Output, schemas)
			if err != nil {
				return "", err
			}

			vobj, ok := v.(*normalized.Object)
			if !ok {
				return "", fmt.Errorf("fn.Output is not an object")
			}

			for _, p := range vobj.Params {
				if p.IsReadOnly() && (p.GetInternalName() == "id" || p.GetInternalName() == "uuid") {
					switch x := p.(type) {
					case *normalized.Int:
						manager.AddStandardImport("strconv", "")
						return fmt.Sprintf("idBuilder.WriteString(strconv.FormatInt(ans.%s, 10))", x.CamelCaseName), nil
					case *normalized.Float:
						manager.AddStandardImport("strconv", "")
						return fmt.Sprintf("idBuilder.WriteString(strconv.FormatFloat(ans.%s, 'g', -1, 64))", x.CamelCaseName), nil
					case *normalized.String:
						return fmt.Sprintf("idBuilder.WriteString(ans.%s)", x.CamelCaseName), nil
					}
					return "", fmt.Errorf("Unsupported read-only field type %T", p)
				}
			}

			return "", fmt.Errorf("no read-only ID-like field found in output")
		},
		"NoLocMap": func() bool {
			return len(locMap) == 0
		},
	}

	t := template.Must(
		template.New(
			"as-terraform-input",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $fn := . }}
    // Create the Terraform ID.
    var idBuilder strings.Builder
{{- range $pp := $fn.PathParams }}
{{- Separator }}
{{- AddParamToTerraformId $pp "input" false }}
{{- end }}
{{- range $qp := $fn.QueryParams }}
{{- Separator }}
{{- AddParamToTerraformId $qp "input" false }}
{{- end }}
{{- if OutputHasMissingIdField }}
{{- $idp := GetIdParam }}
{{- Separator }}
{{- AddParamToTerraformId $idp "ans" true }}
{{- end }}
{{- if NoLocMap }}
    idBuilder.WriteString("x")
{{- end }}
{{- /* End */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, fin)

	return b.String(), locMap, manager, err
}

func AsModelLine(i normalized.Item, prefix, defaultShortName string, schemas map[string]normalized.Item) (string, error) {
	if i == nil {
		return "", fmt.Errorf("item is nil")
	}

	if err := HasReservedName(i); err != nil {
		return "", err
	}

	theType, err := i.TerraformModelType(prefix, defaultShortName, schemas)
	if err != nil {
		return "", err
	}

	ptr := ""
	if iAsItem, ok := i.(*normalized.Object); ok && !iAsItem.IsRequired() {
		ptr = "*"
	}

	return fmt.Sprintf("    %s %s%s `tfsdk:%q`", i.GetCamelCaseName(), ptr, theType, i.GetUnderscoreName()), nil
}

func HasReservedName(i normalized.Item) error {
	if i == nil {
		return fmt.Errorf("normalized item is nil")
	}

	name := i.GetUnderscoreName()
	reservedKeywords := []string{
		"depends_on",
		"count",
		"for_each",
		"provider",
		"lifecycle",
		"dynamic",
	}

	for _, reserved := range reservedKeywords {
		if name == reserved {
			return fmt.Errorf("Path:%#v name is %s, a reserved keyword", i.Path(), reserved)
		}
	}

	return nil
}

func AddItemToManager(i normalized.Item, base, name string, manager *imports.Manager) {
	path := i.GetSdkPath()
	if len(path) == 0 {
		return
	}

	parts := []string{base, name, "schemas"}
	parts = append(parts, path...)
	importPath := strings.Join(parts, "/")

	manager.AddSdkImport(importPath, i.GetShortName())
}

func SaveParamUsingLocMap(i normalized.Item, dst string, locMap map[string]int) (string, *imports.Manager, error) {
	if i == nil {
		return "", nil, fmt.Errorf("item is nil")
	}

	name := i.GetInternalName()
	num, ok := locMap[name]
	if !ok {
		return "", nil, fmt.Errorf("Name %q not in locMap: %#v", name, locMap)
	}

	manager := imports.NewManager()

	fm := template.FuncMap{
		"TokenNumber": func() int { return num },
		"Destination": func() string { return dst },
		"Name":        func() string { return i.GetCamelCaseName() },
		"TypeFunction": func() (string, error) {
			switch i.(type) {
			case *normalized.String:
				return "String", nil
			case *normalized.Bool:
				return "Bool", nil
			case *normalized.Int:
				return "Int64", nil
			case *normalized.Float:
				return "Float64", nil
			}
			return "", fmt.Errorf("Unknown token type: %T", i)
		},
		"TokensStrconv": func() (string, error) {
			manager.AddStandardImport("strconv", "")
			switch i.(type) {
			case *normalized.Bool:
				return fmt.Sprintf("ParseBool(tokens[%d])", num), nil
			case *normalized.Int:
				return fmt.Sprintf("ParseInt(tokens[%d], 10, 64)", num), nil
			case *normalized.Float:
				return fmt.Sprintf("ParseFloat(tokens[%d], 64)", num), nil
			}
			return "", fmt.Errorf("Unsupported token type: %T", i)
		},
		"IsString": func() bool {
			_, ok := i.(*normalized.String)
			return ok
		},
	}

	t := template.Must(
		template.New(
			"save-param-using-loc-map",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $tnum := TokenNumber }}
    if tokens[{{ $tnum }}] == "" {
        {{ Destination }}.{{ Name }} = types.{{ TypeFunction }}Null()
    } else {
{{- if IsString }}
        {{ Destination }}.{{ Name }} = types.{{ TypeFunction }}Value(tokens[{{ $tnum }}])
{{- else }}
        x, err := strconv.{{ TokensStrconv }}
        if err != nil {
            resp.Diagnostics.AddError("Error parsing token {{ $tnum }}: {{ Name }}", err.Error())
            return
        }
        {{ Destination }}.{{ Name }} = types.{{ TypeFunction }}Value(x)
{{- end }}
    }
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, i)

	return b.String(), manager, err
}
