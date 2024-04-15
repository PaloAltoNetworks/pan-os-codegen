package normalized

import (
	"fmt"
	"strings"
	"text/template"
)

func SchemaToSdkPath(i Item, repo, subdir string) (string, error) {
	if i == nil {
		return "", fmt.Errorf("item is nil")
	}

	fm := template.FuncMap{
		"Repo":   func() string { return repo },
		"Subdir": func() string { return subdir },
		"Suffix": func(v []string) string { return strings.Join(v, "/") },
	}

	t := template.Must(
		template.New(
			"schema-to-sdk-path",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{ .ShortName }} "{{ Repo }}/{{ Subdir }}/schemas/{{ Suffix .SdkPath }}"
{{- /* End */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, i)

	return b.String(), err
}

func NamespaceToSdkPath(ns *Namespace, repo, subdir string) (string, error) {
	if ns == nil {
		return "", fmt.Errorf("ns is nil")
	}

	fm := template.FuncMap{
		"Repo":   func() string { return repo },
		"Subdir": func() string { return subdir },
		"Suffix": func(v []string) string { return strings.Join(v, "/") },
	}

	t := template.Must(
		template.New(
			"namespace-to-sdk-path",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{ .ShortName }} "{{ Repo }}/{{ Subdir }}/services/{{ Suffix .SdkPath }}"
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, ns)

	return b.String(), err
}

func GetPathStdImports(v []Item) (map[string]bool, error) {
	if len(v) == 0 {
		return nil, nil
	}

	return map[string]bool{
		"strings": true,
	}, nil
}

func GetQueryStdImports(v []Item) (map[string]bool, error) {
	if len(v) == 0 {
		return nil, nil
	}

	ans := make(map[string]bool)
	for _, i := range v {
		switch x := i.(type) {
		case *Bool:
			ans["strconv"] = true
		case *Int:
			ans["strconv"] = true
		case *Float:
			ans["strconv"] = true
		case *String:
		case *Array:
			if x.Spec == nil {
				return nil, fmt.Errorf("array query param has nil spec")
			}
			a2, err := GetQueryStdImports([]Item{x.Spec})
			if err != nil {
				return nil, fmt.Errorf("array query spec failed to get std imports: %s", err)
			}
			for key, value := range a2 {
				ans[key] = value
			}
		case *Object:
			return nil, fmt.Errorf("not sure how to put an object in query params")
		}
	}

	return ans, nil
}

func TerraformDocstringTemplate(isResource bool) string {
	if isResource {
		return descBasic + space + descValidators
	}

	return descBasic
}

func SdkDocstringTemplate() string {
	return descName + space + descParenthesis + ":" + space + descBasic + space + descValidators
}

func GolangDocstring(i Item, includePreamble, includeValidators bool, suffix string, schemas map[string]Item) (string, error) {
	if i == nil {
		return "", nil
	}

	fm := templateFuncMap(1, false, schemas)
	fm["IncludePreamble"] = func() bool { return includePreamble }
	fm["IncludeValidators"] = func() bool { return includeValidators }
	fm["TrailingPeriod"] = func(s string) (string, error) {
		if strings.Contains(s, "\"") {
			return "", fmt.Errorf("TODO: escaping quoting in the description")
		} else if s == "" {
			return "", fmt.Errorf("String is empty?")
		}

		if strings.HasSuffix(s, ".") {
			return s, nil
		}
		return s + ".", nil
	}
	fm["Suffix"] = func() string { return suffix }

	var b strings.Builder
	t := template.Must(
		template.New(
			"golang-docstring",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- if IncludePreamble -}}
Param {{ .CamelCaseName }} (
{{- GetGolangType . }}
{{- if .ReadOnly }}, read-only
{{- end }}
{{- if .Required }}, required
{{- end -}}
):{{ " " }}
{{- end }}
{{- if eq .Description "" }}The {{ .CamelCaseName }} param.
{{- else }}{{ TrailingPeriod .Description }}
{{- end }}
{{- if ne Suffix "" }} {{ Suffix }}
{{- end }}
{{- if IncludeValidators }}
{{- .ValidatorString true }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, i)

	return b.String(), err
}

func GolangClassCode(item Item, includePreamble, includeValidators, includeShortName bool, suffixes map[string]string, schemas map[string]Item) (string, error) {
	if item == nil {
		return "", nil
	}
	i2, ok := item.(*Object)
	if !ok {
		return "", fmt.Errorf("GolangClassCode needs *Object as input, not %T", item)
	}

	fm := templateFuncMap(1, includeShortName, schemas)
	fm["ParamDoc"] = func(x Item) (string, error) {
		return GolangDocstring(x, includePreamble, includeValidators, suffixes[x.GetInternalName()], schemas)
	}

	var b strings.Builder
	t := template.Must(
		template.New(
			"golang-classcode",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin SDK class template string */ -}}
{{- $cls := . }}
/*
{{ $cls.ClassName }} object.

Parent chains:
{{- range $pname := $cls.Path }}
* {{ $pname }}
{{- end }}

ShortName: {{ .ShortName }}

Args:
{{- range $pname := $cls.OrderedParams 1 }}
{{- $theItem := index $cls.Params $pname }}
{{ ParamDoc $theItem }}
{{- end }}
{{- if ne (len $cls.OneOf) 0 }}

NOTE:  One of the following params should be specified:
{{- range $val := $cls.OneOf }}
    - {{ Name (index $cls.Params $val) }}
{{- end }}
{{- end }}
*/
type {{ $cls.ClassName }} struct {
{{- range $pname := $cls.OrderedParams 1 }}
{{- $theItem := index $cls.Params $pname }}
    {{ Name $theItem }}{{ " " }}
{{- if IsArrayType $theItem }}
{{- else if and (IsBoolType $theItem) ($theItem.IsObjectBool) }}
{{- else if $theItem.Required }}*
{{- end }}
{{- GetGolangType $theItem }}
{{- " " }}` + "`" + `json:"{{ $theItem.Name }}
{{- if not $theItem.Required }},omitempty
{{- end }}"` + "`" + `
{{- end }}
}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, i2)

	return b.String(), err
}

const (
	space = `{{ " " }}`

	descName = `
{{- /* Begin */ -}}
Param {{ Name $theItem }}
{{- /* Done */ -}}
`

	descParenthesis = `
{{- /* Begin */ -}}
(
{{- GetGolangType $theItem }}
{{- if BoolIsTrue $theItem.ReadOnly }}, read-only
{{- end }}
{{- if BoolIsTrue $theItem.Required }}, required
{{- end -}}
)
{{- /* Done */ -}}
`

	descBasic = `
{{- /* Begin */ -}}
{{- if eq $theItem.Description "" }}the {{ Name $theItem }} param.
{{- else }}{{ $theItem.Description }}
{{- end }}
{{- /* Done */ -}}
`

	descValidators = `
{{- /* Begin */ -}}
{{ $theItem.ValidatorString true }}
{{- /* Done */ -}}
`

	sdkGolangClass = `
{{- /* Begin SDK class template string */ -}}
{{- $cls := . }}
/*
{{ $cls.ClassName }}{{ " " }}
{{- if ne $cls.Description "" }}{{ $cls.Description }}
{{- else }}object.
{{- end }}

ShortName: {{ .ShortName }}
Parent chains:
{{- range $pname := $cls.Path }}
* {{ $pname }}
{{- end }}

Args:
{{- range $pname := $cls.OrderedParams 1 }}
{{- $theItem := index $cls.Params $pname }}{{ "\n\n" }}
` + descName + space + descParenthesis + ":" + space + descBasic + descValidators + `
{{- end }}
{{- if ne (len $cls.OneOf) 0 }}

NOTE:  One of the following params should be specified:
{{- range $val := $cls.OneOf }}
    - {{ Name (index $cls.Params $val) }}
{{- end }}
{{- end }}
*/
type {{ $cls.ClassName }} struct {{ "{" }}
{{- range $pname := $cls.OrderedParams 1 }}
{{- $theItem := index $cls.Params $pname }}
{{ "    " }}{{ Name $theItem }}{{ " " }}
{{- /* Determine if the * should be there or not */ -}}
{{- if IsArrayType $theItem }}
{{- else if and (IsBoolType $theItem) (BoolIsTrue $theItem.IsObjectBool) }}
{{- else if not (BoolIsTrue $theItem.Required) }}*
{{- end }}

{{- /* Output the param type */ -}}
{{- GetGolangType $theItem }}

{{- /* Output the JSON string literal */ -}}
{{- " " }}` + "`" + `json:"{{ $theItem.Name }}
{{- if not (BoolIsTrue $theItem.Required) }},omitempty
{{- end }}"` + "`" + `

{{- end }}
{{ "}" }}
{{- /* Done */ -}}
`

	sdkParamPrefix = `
` + space + space + space + space

	sdkString = `
{{- /* Begin SDK template string */ -}}
{{- "    " }}{{ TheName . }}{{ " " }}

{{- /* Determine if the * should be there or not */ -}}
{{- if IsArrayType }}[]
{{- else if and IsBoolType (BoolIsTrue .IsObjectBool) }}
{{- else if not (BoolIsTrue .Required) }}*
{{- end }}

{{- /* Output the param type */ -}}
{{- GetGolangType . }}

{{- /* Output the JSON string literal */ -}}
{{- " " }}` + "`" + `json:"{{ .Name }}
{{- if not (BoolIsTrue .Required) }},omitempty
{{- end }}"` + "`" + `
{{- /* Done */ -}}
`

	sdkNamespaceFunction = `
{{- /* Begin namespace docstring */ -}}
{{- if IsNotNil .Input }}{{ .Input.ToGolangSdkString }}
{{- end }}
{{- if IsNotNil .Output }}
{{- if IsNotNil .Input }}


{{- end }}{{ .Output.ToGolangSdkString }}
{{- end }}
{{- if or (IsNotNil .Input) (IsNotNil .Output) }}


{{ end }}{{ .ToGolangSdkString }}
{{- /* Done */ -}}
`
)

func templateFuncMap(style int, includeShortName bool, schemas map[string]Item) template.FuncMap {
	ans := template.FuncMap{
		"TheName": func(i Item) (string, error) { return "", fmt.Errorf("Name style has not been defined") },
		"Name": func(i Item) (string, error) {
			switch style {
			case 0:
				return i.GetUnderscoreName(), nil
			case 1:
				return i.GetCamelCaseName(), nil
			}

			return "", fmt.Errorf("Unknown style: %d", style)
		},
		"IsArrayType": func(i Item) bool {
			_, ok := i.(*Array)
			return ok
		},
		"IsBoolType": func(i Item) bool {
			_, ok := i.(*Bool)
			return ok
		},
		"IsFloatType": func(i Item) bool {
			_, ok := i.(*Float)
			return ok
		},
		"IsIntType": func(i Item) bool {
			_, ok := i.(*Int)
			return ok
		},
		"IsObjectType": func(i Item) bool {
			_, ok := i.(*Object)
			return ok
		},
		"IsStringType": func(i Item) bool {
			_, ok := i.(*String)
			return ok
		},
		"IsNotNil": func(v any) bool {
			switch x := v.(type) {
			case nil:
				return false
			case *bool:
				return x != nil
			case *float64:
				return x != nil
			case *int64:
				return x != nil
			case *string:
				return x != nil
			case map[string]any:
				return x != nil
			case []string:
				return x != nil
			//case Item:
			//    return x == nil
			case *Function:
				return x != nil
			case *Array:
				return x != nil
			case *Object:
				return x != nil
			}
			panic(fmt.Sprintf("Unknown IsNotNil type: %T", v))
		},
		"BoolIsTrue":    func(v *bool) bool { return v != nil && *v },
		"GetFloat":      func(v *float64) float64 { return *v },
		"GetInt":        func(v *int64) int64 { return *v },
		"GetString":     func(v *string) string { return *v },
		"GetGolangType": func(v Item) (string, error) { return v.GolangType(includeShortName, schemas) },
		"Validators":    func(v Item) string { return v.ValidatorString(true) },
	}

	switch style {
	case 0:
		ans["TheName"] = func(i Item) (string, error) { return i.GetUnderscoreName(), nil }
	case 1:
		ans["TheName"] = func(i Item) (string, error) { return i.GetCamelCaseName(), nil }
	}

	return ans
}
