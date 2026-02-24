package imports_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/imports"
)

var _ = Describe("Import", func() {
	Context("when unmarshalling from YAML", func() {
		Context("with default_value present", func() {
			It("should parse all fields correctly including default_value", func() {
				yamlData := `target: interface
variants:
  - layer2
  - layer3
default_value: vsys1`

				var result imports.Import
				err := yaml.Unmarshal([]byte(yamlData), &result)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Target).Should(Equal("interface"))
				Expect(result.Variants).Should(Equal([]string{"layer2", "layer3"}))
				Expect(result.DefaultValue).ShouldNot(BeNil())
				Expect(*result.DefaultValue).Should(Equal("vsys1"))
			})
		})

		Context("without default_value", func() {
			It("should set DefaultValue to nil", func() {
				yamlData := `target: interface
variants:
  - layer2
  - layer3`

				var result imports.Import
				err := yaml.Unmarshal([]byte(yamlData), &result)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Target).Should(Equal("interface"))
				Expect(result.Variants).Should(Equal([]string{"layer2", "layer3"}))
				Expect(result.DefaultValue).Should(BeNil())
			})
		})

		Context("with empty default_value", func() {
			It("should parse empty string as the default_value", func() {
				yamlData := `target: interface
variants:
  - "*"
default_value: ""`

				var result imports.Import
				err := yaml.Unmarshal([]byte(yamlData), &result)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Target).Should(Equal("interface"))
				Expect(result.Variants).Should(Equal([]string{"*"}))
				Expect(result.DefaultValue).ShouldNot(BeNil())
				Expect(*result.DefaultValue).Should(Equal(""))
			})
		})
	})

	Context("when marshalling to YAML", func() {
		Context("with DefaultValue set", func() {
			It("should include default_value in output", func() {
				input := imports.Import{
					Target:       "interface",
					Variants:     []string{"layer2", "layer3"},
					DefaultValue: stringPtr("vsys1"),
				}

				data, err := yaml.Marshal(input)

				Expect(err).ShouldNot(HaveOccurred())
				output := string(data)
				Expect(output).Should(ContainSubstring("default_value: vsys1"))
			})
		})

		Context("with DefaultValue nil", func() {
			It("should omit default_value from output due to omitempty tag", func() {
				input := imports.Import{
					Target:       "interface",
					Variants:     []string{"layer2", "layer3"},
					DefaultValue: nil,
				}

				data, err := yaml.Marshal(input)

				Expect(err).ShouldNot(HaveOccurred())
				output := string(data)
				Expect(output).ShouldNot(ContainSubstring("default_value"))
			})
		})
	})
})

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}
