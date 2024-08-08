package xpathschema_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/xpath"
)

var _ = Describe("Variable", func() {
	Describe("When unmarshalling yaml document", func() {
		Context("with type set to non-empty string", func() {
			It("should create a valid Variable", func() {
				data := []byte(`{name: variable, type: object}`)
				var variable xpathschema.Variable
				err := yaml.Unmarshal(data, &variable)
				Expect(err).ToNot(HaveOccurred())
				Expect(variable.Name).To(Equal("variable"))
				Expect(variable.Type).To(Equal(xpathschema.VariableObject))
			})
		})
		Context("with type left unset", func() {
			It("should create a valid Variable with type entry", func() {
				data := []byte(`{name: variable}`)
				var variable xpathschema.Variable
				err := yaml.Unmarshal(data, &variable)
				Expect(err).ToNot(HaveOccurred())
				Expect(variable.Name).To(Equal("variable"))
				Expect(variable.Type).To(Equal(xpathschema.VariableEntry))
			})
		})
	})
})
