package terraform_provider

import (
	"fmt"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// locationStructFieldCtx describes a field in a location struct.
type locationStructFieldCtx struct {
	Name          *properties.NameVariant
	TerraformType string
	Type          string
	Tags          []string
}

// locationStructCtx describes a location struct.
type locationStructCtx struct {
	StructName string
	Fields     []locationStructFieldCtx
}

// getLocationStructsContext creates location struct specifications.
func getLocationStructsContext(names *NameProvider, spec *properties.Normalization) []locationStructCtx {
	var locations []locationStructCtx

	if len(spec.Locations) == 0 {
		return nil
	}

	// Create the top location structure that references other locations
	topLocation := locationStructCtx{
		StructName: fmt.Sprintf("%sLocation", names.StructName),
	}

	for _, data := range spec.OrderedLocations() {
		structName := fmt.Sprintf("%s%sLocation", names.StructName, data.Name.CamelCase)
		tfTag := fmt.Sprintf("`tfsdk:\"%s\"`", data.Name.Underscore)
		structType := "types.Object"

		topLocation.Fields = append(topLocation.Fields, locationStructFieldCtx{
			Name:          data.Name,
			TerraformType: structName,
			Type:          structType,
			Tags:          []string{tfTag},
		})

		var fields []locationStructFieldCtx

		for _, i := range spec.Imports {
			if i.Type.CamelCase != data.Name.CamelCase {
				continue
			}

			for _, elt := range i.OrderedLocations() {
				if elt.Required {
					fields = append(fields, locationStructFieldCtx{
						Name: elt.Name,
						Type: "types.String",
						Tags: []string{fmt.Sprintf("`tfsdk:\"%s\"`", elt.Name.Underscore)},
					})
				}
			}
		}

		for _, param := range data.OrderedVars() {
			paramTag := fmt.Sprintf("`tfsdk:\"%s\"`", param.Name.Underscore)
			name := param.Name
			if name.CamelCase == data.Name.CamelCase {
				name = properties.NewNameVariant("name")
				paramTag = "`tfsdk:\"name\"`"
			}
			fields = append(fields, locationStructFieldCtx{
				Name: name,
				Type: "types.String",
				Tags: []string{paramTag},
			})
		}

		location := locationStructCtx{
			StructName: structName,
			Fields:     fields,
		}
		locations = append(locations, location)
	}

	locations = append(locations, topLocation)

	return locations
}

// RenderLocationStructs generates location struct definitions.
func RenderLocationStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Locations []locationStructCtx
	}

	locations := getLocationStructsContext(names, spec)

	data := context{
		Locations: locations,
	}
	return processTemplate("location/render.tmpl", "render-location-structs", data, commonFuncMap)
}

// RenderLocationSchemaGetter generates a location schema getter function.
func RenderLocationSchemaGetter(names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	var attributes []attributeCtx

	if len(spec.Locations) == 0 {
		return "", nil
	}

	var locations []string
	for _, loc := range spec.OrderedLocations() {
		locations = append(locations, loc.Name.Underscore)
	}

	var idx int
	for _, data := range spec.OrderedLocations() {
		var variableAttrs []attributeCtx

		for _, i := range spec.Imports {
			if i.Type.CamelCase != data.Name.CamelCase {
				continue
			}

			for _, elt := range i.OrderedLocations() {
				if elt.Required {
					var defaultValue *defaultCtx
					for varName, variable := range elt.XpathVariables {
						if varName == elt.Name.Original && variable.Default != "" {
							defaultValue = &defaultCtx{
								Type:  "stringdefault.StaticString",
								Value: fmt.Sprintf(`"%s"`, variable.Default),
							}
						}
					}
					variableAttrs = append(variableAttrs, attributeCtx{
						Name:         elt.Name,
						SchemaType:   "rsschema.StringAttribute",
						Required:     defaultValue == nil,
						Optional:     defaultValue != nil,
						Computed:     defaultValue != nil,
						ModifierType: "String",
						Default:      defaultValue,
					})
				}
			}
		}

		for _, variable := range data.OrderedVars() {
			name := variable.Name
			if name.CamelCase == data.Name.CamelCase {
				name = properties.NewNameVariant("name")
			}
			attribute := attributeCtx{
				Name:        name,
				Description: variable.Description,
				SchemaType:  "rsschema.StringAttribute",
				Optional:    true,
				Required:    false,
				Computed:    true,
				Default: &defaultCtx{
					Type:  "stringdefault.StaticString",
					Value: fmt.Sprintf(`"%s"`, variable.Default),
				},
				ModifierType: "String",
			}
			variableAttrs = append(variableAttrs, attribute)
		}

		modifierType := "Object"

		var validators *validatorCtx
		if len(locations) > 1 && idx == 0 {
			var expressions []string
			for _, location := range locations {
				expressions = append(expressions, fmt.Sprintf(`path.MatchRelative().AtParent().AtName("%s")`, location))
			}

			functions := []validatorFunctionCtx{{
				Function:    "ExactlyOneOf",
				Expressions: expressions,
			}}

			typ := data.ValidatorType()
			validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
			manager.AddHashicorpImport(validatorImport, "")

			validators = &validatorCtx{
				ListType:  pascalCase(typ),
				Package:   fmt.Sprintf("%svalidator", typ),
				Functions: functions,
			}
		}

		attribute := attributeCtx{
			Name:         data.Name,
			SchemaType:   "rsschema.SingleNestedAttribute",
			Description:  data.Description,
			Optional:     true,
			Required:     false,
			Attributes:   variableAttrs,
			ModifierType: modifierType,
			Validators:   validators,
		}
		attributes = append(attributes, attribute)

		idx += 1
	}

	locationName := properties.NewNameVariant("location")

	topAttribute := attributeCtx{
		Name:         locationName,
		SchemaType:   "rsschema.SingleNestedAttribute",
		Description:  "The location of this object.",
		Required:     true,
		Attributes:   attributes,
		ModifierType: "Object",
	}

	type context struct {
		StructName string
		Schema     attributeCtx
	}

	data := context{
		StructName: names.StructName,
		Schema:     topAttribute,
	}

	return processTemplate("schema/location_schema_getter.tmpl", "render-location-schema-getter", data, commonFuncMap)
}

// marshallerFieldSpec describes a field in a marshaller struct.
type marshallerFieldSpec struct {
	Name       *properties.NameVariant
	Type       string
	StructName string
	Tags       string
}

// marshallerSpec describes a marshaller struct.
type marshallerSpec struct {
	StructName string
	Fields     []marshallerFieldSpec
}

// createLocationMarshallerSpecs creates marshaller specifications for locations.
func createLocationMarshallerSpecs(names *NameProvider, spec *properties.Normalization) []marshallerSpec {
	var specs []marshallerSpec

	var topFields []marshallerFieldSpec
	for _, loc := range spec.OrderedLocations() {
		topFields = append(topFields, marshallerFieldSpec{
			Name:       loc.Name,
			Type:       "types.Object",
			StructName: fmt.Sprintf("%s%sLocation", names.StructName, loc.Name.CamelCase),
			Tags:       fmt.Sprintf("`json:\"%s,omitempty\"`", loc.Name.Underscore),
		})

		var fields []marshallerFieldSpec
		for _, field := range loc.OrderedVars() {
			name := field.Name
			tag := field.Name.Underscore
			if name.CamelCase == loc.Name.CamelCase {
				name = properties.NewNameVariant("name")
				tag = "name"
			}

			fields = append(fields, marshallerFieldSpec{
				Name: name,
				Type: "string",
				Tags: fmt.Sprintf("`json:\"%s,omitempty\"`", tag),
			})
		}

		// Add import location (e.g. vsys) name to location
		for _, i := range spec.Imports {
			if i.Type.CamelCase != loc.Name.CamelCase {
				continue
			}

			for _, elt := range i.OrderedLocations() {
				if elt.Required {
					fields = append(fields, marshallerFieldSpec{
						Name: elt.Name,
						Type: "string",
						Tags: fmt.Sprintf("`tfsdk:\"%s\"`", elt.Name.Underscore),
					})
				}
			}
		}

		specs = append(specs, marshallerSpec{
			StructName: fmt.Sprintf("%s%sLocation", names.StructName, loc.Name.CamelCase),
			Fields:     fields,
		})
	}

	specs = append(specs, marshallerSpec{
		StructName: fmt.Sprintf("%sLocation", names.StructName),
		Fields:     topFields,
	})

	return specs
}

// RenderLocationMarshallers generates location marshaller structs.
func RenderLocationMarshallers(names *NameProvider, spec *properties.Normalization) (string, error) {
	var context struct {
		Specs []marshallerSpec
	}
	context.Specs = createLocationMarshallerSpecs(names, spec)

	return processTemplate("location/marshallers.tmpl", "render-location-marshallers", context, commonFuncMap)
}

// locationFieldCtx describes a location field for conversion.
type locationFieldCtx struct {
	PangoName     string
	TerraformName string
	Type          string
}

// locationCtx describes a location for conversion.
type locationCtx struct {
	Name                string
	PangoStructName     string
	TerraformStructName string
	SdkStructName       string
	Fields              []locationFieldCtx
}

// renderLocationsGetContext creates location conversion specifications.
func renderLocationsGetContext(names *NameProvider, spec *properties.Normalization) []locationCtx {
	var locations []locationCtx

	for _, location := range spec.OrderedLocations() {
		var fields []locationFieldCtx
		for _, variable := range location.OrderedVars() {
			name := variable.Name.CamelCase
			if variable.Name.CamelCase == location.Name.CamelCase {
				name = "Name"
			}

			fields = append(fields, locationFieldCtx{
				PangoName:     variable.Name.CamelCase,
				TerraformName: name,
				Type:          "String",
			})
		}
		locations = append(locations, locationCtx{
			Name:                location.Name.CamelCase,
			PangoStructName:     fmt.Sprintf("%s.%sLocation", names.PackageName, location.Name.CamelCase),
			TerraformStructName: fmt.Sprintf("%s%sLocation", names.StructName, location.Name.CamelCase),
			SdkStructName:       fmt.Sprintf("%s.%sLocation", names.PackageName, location.Name.CamelCase),
			Fields:              fields,
		})
	}

	return locations
}

// RenderLocationsPangoToState generates code to convert Pango locations to Terraform state.
func RenderLocationsPangoToState(names *NameProvider, spec *properties.Normalization, source string, dest string) (string, error) {
	type context struct {
		Source    string
		Dest      string
		Locations []locationCtx
	}
	data := context{Source: source, Dest: dest, Locations: renderLocationsGetContext(names, spec)}
	return processTemplate("location/pango_to_state.tmpl", "render-locations-pango-to-state", data, commonFuncMap)
}

// RenderLocationsStateToPango generates code to convert Terraform state to Pango locations.
func RenderLocationsStateToPango(names *NameProvider, spec *properties.Normalization, source string, dest string) (string, error) {
	type context struct {
		TerraformStructName string
		Source              string
		Dest                string
		Locations           []locationCtx
	}
	data := context{
		TerraformStructName: fmt.Sprintf("%sLocation", names.StructName),
		Locations:           renderLocationsGetContext(names, spec),
		Source:              source,
		Dest:                dest,
	}
	return processTemplate("location/state_to_pango.tmpl", "render-locations-state-to-pango", data, commonFuncMap)
}

// RendeCreateUpdateMovementRequired generates code to check if movement is required.
func RendeCreateUpdateMovementRequired(state string, entries string) (string, error) {
	type context struct {
		State   string
		Entries string
	}
	data := context{State: state, Entries: entries}
	return processTemplate("resource/movement_required.tmpl", "render-create-update-movement-required", data, nil)
}

// RenderLocationAttributeTypes generates location attribute type definitions.
func RenderLocationAttributeTypes(names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Specs []locationStructCtx
	}

	locations := getLocationStructsContext(names, spec)

	data := context{
		Specs: locations,
	}
	return processTemplate("location/attribute_types.tmpl", "render-location-structs", data, commonFuncMap)
}
