package imports

import (
    "fmt"
    "sort"
    "strings"
    "text/template"
)


func NewManager() *Manager {
    return &Manager{
        Standard: make(map[string] string),
        Sdk: make(map[string] string),
        Hashicorp: make(map[string] string),
        Other: make(map[string] string),
    }
}

type Manager struct {
    Standard map[string] string
    Sdk map[string] string
    Hashicorp map[string] string
    Other map[string] string
}

func (o *Manager) AddStandardImport(path, shortName string) { o.Standard[path] = shortName }
func (o *Manager) AddSdkImport(path, shortName string) { o.Sdk[path] = shortName }
func (o *Manager) AddHashicorpImport(path, shortName string) { o.Hashicorp[path] = shortName }
func (o *Manager) AddOtherImport(path, shortName string) { o.Other[path] = shortName }

func (o *Manager) Merge(v *Manager) {
    if v == nil {
        return
    }

    for key, value := range v.Standard {
        o.Standard[key] = value
    }

    for key, value := range v.Sdk {
        o.Sdk[key] = value
    }

    for key, value := range v.Hashicorp {
        o.Hashicorp[key] = value
    }

    for key, value := range v.Other {
        o.Other[key] = value
    }
}

func (o *Manager) RenderImports() (string, error) {
    needSpacing := false

    fm := template.FuncMap{
        "SortLibs": func(v map[string] string) []string { return sortImports(v) },
        "Render": func(v map[string] string) (string, error) {
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
{{- Render $obj.Standard }}
{{- Render $obj.Sdk }}
{{- Render $obj.Hashicorp }}
{{- Render $obj.Other }}
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

func sortImports(v map[string] string) []string {
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
