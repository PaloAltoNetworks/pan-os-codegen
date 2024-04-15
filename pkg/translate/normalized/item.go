package normalized

import (
	"fmt"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
)

type Item interface {
	Path() []string
	Copy() Item
	ApplyUserConfig(Item)
	String() string
	NameAs(int) string
	GolangType(bool, map[string]Item) (string, error)
	ValidatorString(bool) string
	GetInternalName() string
	GetUnderscoreName() string
	GetCamelCaseName() string
	SchemaInit(string, string) error
	GetShortName() string
	Items() []Item
	GetItems(bool, bool, map[string]Item) ([]Item, error)
	ToGolangSdkString(string, string, map[string]Item) (string, error)
	SchemaReferences() []string
	ApplyParameterConfig(string, bool) error
	GetLocation() string
	GetReference() string
	GetSdkImports(bool, map[string]Item) (map[string]bool, error)
	PackageName() string
	GetSdkPath() []string
	ToGolangSdkQueryParam() (string, bool, error)
	ToGolangSdkPathParam() (string, bool, error)
	Rename(string)
	TerraformModelType(string, string, map[string]Item) (string, error)
	IsRequired() bool
	IsReadOnly() bool
	HasDefault() bool
	ClearDefault()
	GetObjects(map[string]Item) ([]*Object, error)
	EncryptedParams() []*String
	RootParent() Item
	EncHasName() (bool, error)
	GetParent() Item
	SetParent(Item)
	IsEncrypted() bool
	HasEncryptedItems(map[string]Item) (bool, error)
	GetEncryptionKey(string, string, bool, byte) (string, error)
	RenderTerraformValidation() ([]string, *imports.Manager, error)
}

func ItemLookup(i Item, schemas map[string]Item) (Item, error) {
	if i == nil {
		return nil, fmt.Errorf("no item passed in to item lookup")
	}

	ref := i.GetReference()
	if ref == "" {
		return i, nil
	}

	switch x := i.(type) {
	case *Array:
		other := schemas[ref]
		if other == nil {
			return nil, fmt.Errorf("item is %T but ref doesn't exist", x)
		}
		if _, ok := other.(*Array); !ok {
			return nil, fmt.Errorf("item is %T, ref is %T", x, other)
		}
		return other, nil
	case *Object:
		other := schemas[ref]
		if other == nil {
			return nil, fmt.Errorf("item is %T but ref doesn't exist", x)
		}
		if _, ok := other.(*Object); !ok {
			return nil, fmt.Errorf("item is %T, ref is %T", x, other)
		}
		return other, nil
	}

	// NOTE: I am intentionally not handling a schema that is a "basic" type
	// in the above.  If this becomes a problem then it can be addressed later.
	return i, nil
}
