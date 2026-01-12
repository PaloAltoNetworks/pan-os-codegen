package terraform_provider

type Field struct {
	Name    string
	Type    string
	TagName string
}

type StructData struct {
	StructName string
	Fields     []Field
}

// ParamToModelBasic converts the given parameter name and properties to a model representation.
func ParamToModelBasic(paramName string, paramProp interface{}) (string, error) {
	data := map[string]interface{}{
		"paramName": paramName,
	}
	paramPropMap := structToMap(paramProp)
	for k, v := range paramPropMap {
		data[k] = v
	}
	return processTemplate("provider/model_field.tmpl", "param-to-model", data, nil)
}

// ParamToSchemaProvider converts the given parameter name and properties to a schema representation.
func ParamToSchemaProvider(paramName string, paramProp interface{}) (string, error) {
	data := map[string]interface{}{
		"paramName": paramName,
	}
	paramPropMap := structToMap(paramProp)
	for k, v := range paramPropMap {
		data[k] = v
	}
	return processTemplate("provider/schema_attribute.tmpl", "param-to-schema", data, nil)
}

func CreateResourceSchemaLocationAttribute() (string, error) {
	return processTemplate("schema/location_attribute.tmpl", "resource-schema-location", nil, nil)
}
