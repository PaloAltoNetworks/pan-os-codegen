package imports

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
)

type ImportType int

const (
	Standard ImportType = iota
	Sdk
	Hashicorp
	Other
)

type ImportManager struct {
	Imports map[ImportType]map[string]string
}

func NewImportManager() *ImportManager {
	return &ImportManager{
		Imports: map[ImportType]map[string]string{
			Standard:  make(map[string]string),
			Sdk:       make(map[string]string),
			Hashicorp: make(map[string]string),
			Other:     make(map[string]string),
		},
	}
}

func (o *ImportManager) AddImport(importType ImportType, path, shortName string) {
	o.Imports[importType][path] = shortName
}

func (o *ImportManager) Merge(v *ImportManager) {
	if v == nil {
		return
	}
	for importType, imports := range v.Imports {
		for key, value := range imports {
			o.Imports[importType][key] = value
		}
	}
}
func (o *ImportManager) RenderImports() (string, error) {
	needSpacing := false
	fm := template.FuncMap{
		"SortLibs": func(v map[string]string) []string { return sortImports(v) },
		"Render": func(v map[string]string) (string, error) {
			list := sortImports(v)
			ans, err := renderBlock(list)
			if ans != "" {
				if needSpacing {
					ans = fmt.Sprintf("\n%s", ans)
				} else {
					needSpacing = true
				}
			}
			return ans, err
		},
	}
	t := template.Must(
		template.New(
			"import-render",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $obj := . }}
import (
{{- range $importType, $imports := $obj.Imports }}
{{- Render $imports }}
{{- end }}
)
{{- /* Done */ -}}`,
		),
	)
	var b strings.Builder
	err := t.Execute(&b, o)
	return b.String(), err
}

func renderBlock(libs []string) (string, error) {
	if len(libs) == 0 {
		return "", nil
	}

	t := template.Must(
		template.New(
			"render-import-block",
		).Parse(`
{{- /* Begin */ -}}
{{- range $lib := . }}
    {{ $lib }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, libs)

	return b.String(), err
}

func sortImports(v map[string]string) []string {
	libs := make([]string, 0, len(v))
	for key := range v {
		libs = append(libs, key)
	}
	sort.Strings(libs)

	ans := make([]string, 0, len(libs))
	for _, path := range libs {
		if v[path] == "" {
			ans = append(ans, fmt.Sprintf("%q", path))
		} else {
			ans = append(ans, fmt.Sprintf("%s %q", v[path], path))
		}
	}

	return ans
}
