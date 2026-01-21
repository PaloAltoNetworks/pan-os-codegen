package terraform_provider

import (
	"fmt"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
)

// RenderImportStateMarshallers generates import state marshaller structs.
func RenderImportStateMarshallers(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	// Only singular entries can be imported at the time
	if resourceTyp == properties.ResourceCustom || resourceTyp == properties.ResourceConfig {
		return "", nil
	}

	var context struct {
		Specs []marshallerSpec
	}
	context.Specs = createImportStateMarshallerSpecs(resourceTyp, names, spec)

	return processTemplate("location/marshallers.tmpl", "render-import-state-marshallers", context, commonFuncMap)
}

// RenderImportLocationAssignment generates code to assign import location data.
func RenderImportLocationAssignment(names *NameProvider, spec *properties.Normalization, source string, dest string) (string, error) {
	if len(spec.Imports) == 0 {
		return "", nil
	}

	type importVariantSpec struct {
		PangoStructNames *map[string]string
	}

	type importLocationSpec struct {
		TerraformStructName string
		Name                *properties.NameVariant
		Fields              []string
	}

	type importSpec struct {
		TerraformStructName string
		Name                *properties.NameVariant
		Locations           []importLocationSpec
	}

	var importSpecs []importSpec
	variantsByName := make(map[string]importVariantSpec)
	for _, elt := range spec.Imports {
		existing, found := variantsByName[elt.Type.CamelCase]
		if !found {
			pangoStructNames := make(map[string]string)
			existing = importVariantSpec{
				PangoStructNames: &pangoStructNames,
			}
		}

		var locations []importLocationSpec
		for _, loc := range elt.Locations {
			if !loc.Required {
				continue
			}

			var fields []string
			for _, elt := range loc.XpathVariables {
				fields = append(fields, elt.Name.CamelCase)
			}

			tfStructName := fmt.Sprintf("%s%sLocation", names.StructName, elt.Type.CamelCase)
			pangoStructName := fmt.Sprintf("%s%s%sImportLocation", elt.Variant.CamelCase, elt.Type.CamelCase, loc.Name.CamelCase)
			(*existing.PangoStructNames)[loc.Name.CamelCase] = pangoStructName
			locations = append(locations, importLocationSpec{
				TerraformStructName: tfStructName,
				Name:                loc.Name,
				Fields:              fields,
			})
		}
		variantsByName[elt.Type.CamelCase] = existing

		importSpecs = append(importSpecs, importSpec{
			Name:      elt.Type,
			Locations: locations,
		})
	}

	type context struct {
		TerraformStructName string
		PackageName         string
		Source              string
		Dest                string
		Variants            map[string]importVariantSpec
		Specs               []importSpec
	}

	data := context{
		TerraformStructName: fmt.Sprintf("%sLocation", names.StructName),
		PackageName:         names.PackageName,
		Source:              source,
		Dest:                dest,
		Variants:            variantsByName,
		Specs:               importSpecs,
	}

	funcMap := template.FuncMap{
		"GetPangoStructForLocation": func(variants map[string]importVariantSpec, typ *properties.NameVariant, location *properties.NameVariant) (string, error) {
			variantSpec, found := variants[typ.CamelCase]
			if !found {
				return "", fmt.Errorf("failed to find variant for type '%s'", typ.CamelCase)
			}

			structName, found := (*variantSpec.PangoStructNames)[location.CamelCase]
			if !found {
				return "", fmt.Errorf("failed to find variant for type '%s', location '%s'", typ.CamelCase, location.CamelCase)
			}

			return structName, nil
		},
	}

	return processTemplate("location/assignment.tmpl", "render-locations-pango-to-state", data, funcMap)
}

// importStateStructFieldSpec describes a field in an import state struct.
type importStateStructFieldSpec struct {
	Name          string
	TerraformType string
	Type          string
	Tags          string
}

// importStateStructSpec describes an import state struct.
type importStateStructSpec struct {
	StructName string
	Fields     []importStateStructFieldSpec
}

// RenderImportStateStructs generates import state struct definitions.
func RenderImportStateStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	// Only singular entries can be imported at the time
	if resourceTyp == properties.ResourceCustom || resourceTyp == properties.ResourceConfig {
		return "", nil
	}

	type context struct {
		Specs []importStateStructSpec
	}

	data := context{
		Specs: createImportStateStructSpecs(resourceTyp, names, spec),
	}

	return processTemplate("import/import_structs.tmpl", "render-import-state-structs", data, nil)
}

// ResourceImportStateFunction generates the import state function for resources.
func ResourceImportStateFunction(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceConfig || resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	type context struct {
		StructName      string
		PluralType      object.TerraformPluralType
		ResourceIsList  bool
		HasPosition     bool
		HasEntryName    bool
		ListAttribute   *properties.NameVariant
		ListStructName  string
		PangoStructName string
		HasParent       bool
		ParentAttribute *properties.NameVariant
	}

	data := context{
		StructName: names.StructName,
	}

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.ParentAttribute = parentParam.Name
			data.HasParent = true
		}
		data.HasEntryName = spec.HasEntryName()
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.PluralType = spec.TerraformProviderConfig.PluralType
			data.ParentAttribute = parentParam.Name
			data.HasParent = true

		} else {
			listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
			data.PluralType = spec.TerraformProviderConfig.PluralType
			data.ListAttribute = listAttribute
			data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
			data.PangoStructName = fmt.Sprintf("%s.Entry", names.PackageName)
		}
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		data.ResourceIsList = true
		data.PluralType = spec.TerraformProviderConfig.PluralType
		listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		data.ListAttribute = properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
		data.PangoStructName = fmt.Sprintf("%s.Entry", names.PackageName)
		if resourceTyp == properties.ResourceUuidPlural {
			data.HasPosition = true
		}
	case properties.ResourceCustom, properties.ResourceConfig:
		panic("unreachable")
	}

	funcMap := template.FuncMap{
		"ConfigToEntry": ConfigEntry,
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, spec, "import")
		},
	}

	return processTemplate("import/import_state.tmpl", "resource-import-state-function", data, funcMap)
}

// RenderImportStateCreator generates the import state creator function.
func RenderImportStateCreator(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceConfig || resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	type context struct {
		FuncName         string
		ModelName        string
		StructNamePrefix string
		ListAttribute    *properties.NameVariant
		ListStructName   string
		ResourceType     properties.ResourceType
		HasParent        bool
		ParentAttribute  *properties.NameVariant
	}

	data := context{
		FuncName:         fmt.Sprintf("%sImportStateCreator", names.StructName),
		ModelName:        fmt.Sprintf("%sModel", names.ResourceStructName),
		ResourceType:     resourceTyp,
		StructNamePrefix: names.StructName,
	}

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.HasParent = true
			data.ParentAttribute = parentParam.Name
		}
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.HasParent = true
			data.ParentAttribute = parentParam.Name
		} else {
			listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
			data.ListAttribute = listAttribute
			data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
		}
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		data.ListAttribute = listAttribute
		data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
	case properties.ResourceCustom, properties.ResourceConfig:
		panic("unreachable")
	}

	return processTemplate("import/import_creator.tmpl", "render-import-state-creator", data, commonFuncMap)
}

// createImportStateStructSpecs creates import state struct specifications.
func createImportStateStructSpecs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) []importStateStructSpec {
	var specs []importStateStructSpec

	var fields []importStateStructFieldSpec
	fields = append(fields, importStateStructFieldSpec{
		Name:          "Location",
		TerraformType: fmt.Sprintf("%sLocation", names.StructName),
		Type:          "types.Object",
		Tags:          "`json:\"location\"`",
	})

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, importStateStructFieldSpec{
				Name: parentParam.Name.CamelCase,
				Type: "types.String",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		}

		fields = append(fields, importStateStructFieldSpec{
			Name: "Name",
			Type: "types.String",
			Tags: "`json:\"name\"`",
		})
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, importStateStructFieldSpec{
				Name: parentParam.Name.CamelCase,
				Type: "types.String",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		} else {
			fields = append(fields, importStateStructFieldSpec{
				Name: "Names",
				Type: "types.List",
				Tags: "`json:\"names\"`",
			})
		}
	case properties.ResourceUuid:
		fields = append(fields, importStateStructFieldSpec{
			Name: "Names",
			Type: "types.List",
			Tags: "`json:\"names\"`",
		})
	case properties.ResourceUuidPlural:
		fields = append(fields, importStateStructFieldSpec{
			Name: "Names",
			Type: "types.List",
			Tags: "`json:\"names\"`",
		})
		fields = append(fields, importStateStructFieldSpec{
			Name:          "Position",
			TerraformType: "TerraformPositionObject",
			Type:          "types.Object",
			Tags:          "`json:\"position\"`",
		})
	case properties.ResourceCustom, properties.ResourceConfig:
		panic("unreachable")
	}

	specs = append(specs, importStateStructSpec{
		StructName: fmt.Sprintf("%sImportState", names.StructName),
		Fields:     fields,
	})

	return specs
}

// createImportStateMarshallerSpecs creates import state marshaller specifications.
func createImportStateMarshallerSpecs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) []marshallerSpec {
	var specs []marshallerSpec

	var fields []marshallerFieldSpec

	fields = append(fields, marshallerFieldSpec{
		Name:       properties.NewNameVariant("location"),
		Type:       "types.Object",
		StructName: fmt.Sprintf("%sLocation", names.StructName),
		Tags:       "`json:\"location\"`",
	})

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, marshallerFieldSpec{
				Name: parentParam.Name,
				Type: "string",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		}

		fields = append(fields, marshallerFieldSpec{
			Name: properties.NewNameVariant("name"),
			Type: "string",
			Tags: "`json:\"name\"`",
		})
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, marshallerFieldSpec{
				Name: parentParam.Name,
				Type: "string",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		} else {
			fields = append(fields, marshallerFieldSpec{
				Name:       properties.NewNameVariant("names"),
				Type:       "types.List",
				StructName: "[]string",
				Tags:       "`json:\"names\"`",
			})
		}
	case properties.ResourceUuid:
		fields = append(fields, marshallerFieldSpec{
			Name:       properties.NewNameVariant("names"),
			Type:       "types.List",
			StructName: "[]string",
			Tags:       "`json:\"names\"`",
		})
	case properties.ResourceUuidPlural:
		fields = append(fields, marshallerFieldSpec{
			Name:       properties.NewNameVariant("names"),
			Type:       "types.List",
			StructName: "[]string",
			Tags:       "`json:\"names\"`",
		})
		fields = append(fields, marshallerFieldSpec{
			Name:       properties.NewNameVariant("position"),
			Type:       "types.Object",
			StructName: "TerraformPositionObject",
			Tags:       "`json:\"position\"`",
		})
	case properties.ResourceCustom, properties.ResourceConfig:
		panic(fmt.Sprintf("unreachable state: %s", resourceTyp))
	}

	specs = append(specs, marshallerSpec{
		StructName: fmt.Sprintf("%sImportState", names.StructName),
		Fields:     fields,
	})

	return specs
}
