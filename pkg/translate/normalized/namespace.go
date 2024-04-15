package normalized

import (
	"fmt"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

type Namespace struct {
	Create *Function
	Read   *Function
	Update *Function
	Delete *Function
	List   *Function
	Misc   map[string]*Function

	Name      string   `json:"name" yaml:"name"`
	Schema    string   `json:"schema" yaml:"schema"`
	SdkPath   []string `json:"sdk_path" yaml:"sdk_path"`
	ShortName string   `json:"-" yaml:"-"`

	namer *naming.Namer
}

func NewNamespace(fn *Function, namer *naming.Namer) *Namespace {
	if fn == nil || namer == nil {
		return nil
	}

	schema := fn.AssociatedSchema()
	shortName := namer.NewSlug(schema + " sdk")
	theName := schema
	if strings.HasPrefix(theName, SchemaPrefix) {
		theName = theName[len(SchemaPrefix):]
	}
	theName = naming.Underscore("", theName, "")

	ans := &Namespace{
		Name:      theName,
		Schema:    schema,
		SdkPath:   naming.PathNameToSdkPath(fn.Uri),
		ShortName: shortName,
		Misc:      make(map[string]*Function),

		namer: namer,
	}

	return ans
}

func (o *Namespace) ModuleSuffix() string {
	if o.Name != "" {
		return o.Name
	}

	if o.Schema == "" {
		return "emptyschema"
	} else if !strings.HasPrefix(o.Schema, SchemaPrefix) {
		return "noschemaprefix"
	}

	x := o.Schema[len(SchemaPrefix):]
	return naming.Underscore("", x, "")
}

func (o *Namespace) String() string {
	if o == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("[%s]\n", o.Schema))
	b.WriteString(fmt.Sprintf("    ShortName: %q", o.ShortName))

	if o.Create != nil {
		b.WriteString(fmt.Sprintf("\n    Create: %s", o.Create))
	}
	if o.Read != nil {
		b.WriteString(fmt.Sprintf("\n    Read: %s", o.Read))
	}
	if o.Update != nil {
		b.WriteString(fmt.Sprintf("\n    Update: %s", o.Update))
	}
	if o.Delete != nil {
		b.WriteString(fmt.Sprintf("\n    Delete: %s", o.Delete))
	}
	if o.List != nil {
		b.WriteString(fmt.Sprintf("\n    List: %s", o.List))
	}

	if len(o.Misc) != 0 {
		b.WriteString("\n    Misc:")
		for _, x := range o.Misc {
			b.WriteString(fmt.Sprintf("\n      * %s", x))
		}
	}

	return b.String()
}

func (o *Namespace) ApplyUserConfig(v *Namespace) {
	if o == nil || v == nil {
		return
	}

	if len(v.SdkPath) != 0 {
		o.SdkPath = naming.MakePathFrom(v.SdkPath)
	}

	if v.Name != "" {
		o.Name = v.Name
	}
}

func (o *Namespace) Path(base, name string) string {
	parts := []string{base, name, "services"}
	parts = append(parts, o.SdkPath...)

	return strings.Join(parts, "/")
}
