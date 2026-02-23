package terraform_provider

import (
	"log"
	"runtime/debug"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// encryptedValuesContext holds the context for rendering encrypted value management code.
type encryptedValuesContext struct {
	SchemaType properties.SchemaType
	Method     string
}

// RenderEncryptedValuesInitialization generates code to initialize encrypted values management.
func RenderEncryptedValuesInitialization(schemaTyp properties.SchemaType, spec *properties.Normalization, method string) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	data := encryptedValuesContext{
		SchemaType: schemaTyp,
		Method:     method,
	}

	return processTemplate("encrypted/initialization.tmpl", "encrypted-values-manager-initialization", data, nil)
}

// RenderEncryptedValuesFinalizer generates code to finalize encrypted values management.
func RenderEncryptedValuesFinalizer(schemaTyp properties.SchemaType, spec *properties.Normalization) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	data := encryptedValuesContext{
		SchemaType: schemaTyp,
	}

	return processTemplate("encrypted/finalizer.tmpl", "encrypted-values-manager-finalizer", data, nil)
}
