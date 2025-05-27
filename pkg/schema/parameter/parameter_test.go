package parameter_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/profile"
)

var _ = Describe("Parameter", func() {
	Describe("When unmarshalling YAML object with hashing", func() {
		Context("with hashing type set to solo", func() {
			It("should unmarshal into a parameter with hashing solo spec", func() {
				bytes := []byte(`{name: test-param, type: string, profiles: [], spec: {}, hashing: { type: solo }}`)
				var param parameter.Parameter
				err := yaml.Unmarshal(bytes, &param)
				Expect(err).ToNot(HaveOccurred())
				Expect(param.Hashing.Type).To(Equal(parameter.HashingSoloType))
			})
		})

		Context("with hashing type set to custom", func() {
			It("should unmarshal into a parameter with hashing custom spec", func() {
				bytes := []byte(`{name: test-param, type: string, profiles: [], spec: {}, hashing: { type: client,  spec: { hashing_func: { name: CustomHashingMethod}}}}}`)
				var param parameter.Parameter
				err := yaml.Unmarshal(bytes, &param)
				Expect(err).ToNot(HaveOccurred())

				Expect(param.Hashing.Type).To(Equal(parameter.HashingClientType))
				spec, ok := param.Hashing.Spec.(*parameter.HashingClientSpec)
				Expect(ok).To(BeTrue())

				Expect(spec.HashingFunc.Name).To(Equal("CustomHashingMethod"))
			})
		})
	})
	Describe("When unmarshalling YAML object into parameter", func() {
		Context("with parameter type set to nil", func() {
			It("should unmarshal into Parameter with ParameterNilSpec Spec", func() {
				bytes := []byte(`{name: test-param, type: nil, profiles: [], spec: {}}`)
				var param parameter.Parameter
				err := yaml.Unmarshal(bytes, &param)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
	Describe("When a list parameter is created", func() {
		Context("with required set to true, items' type set to a simple type and with singular profile xpath", func() {
			var param parameter.Parameter
			BeforeEach(func() {
				param = parameter.Parameter{
					Name:     "test-params",
					Required: true,
					Type:     "list",
					Spec: &parameter.ListSpec{
						Items: parameter.ListSpecElement{
							Type: "string",
						},
					},
					Profiles: []profile.Profile{
						{
							Type:  "entry",
							Xpath: []string{"test-param"},
						},
					},
				}
			})
			It("the required state should be correct", func() {
				Expect(param.Required).To(BeTrue())
			})
			It("the SpecItemsType shortcut should return correct type of the items", func() {
				Expect(param.SpecItemsType()).To(Equal("string"))
			})
			It("the singular name should be properly generated", func() {
				Expect(param.SingularName()).To(Equal("test-param"))
			})
		})
		Context("with required not set explicitly", func() {
			var param parameter.Parameter
			BeforeEach(func() {
				param = parameter.Parameter{
					Name: "test-params",
					Type: "list",
					Spec: &parameter.ListSpec{
						Items: parameter.ListSpecElement{
							Type: "object",
						},
					},
					Profiles: []profile.Profile{
						{
							Type:  "entry",
							Xpath: []string{"test-param", "entry"},
						},
					},
				}
			})
			It("the required should return false", func() {
				Expect(param.Required).To(BeFalse())
			})
			It("the singular name should be properly generated", func() {
				Expect(param.SingularName()).To(Equal("test-param"))
			})
		})
	})
})
