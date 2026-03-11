package terraform_provider_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/terraform_provider"
)

var _ = Describe("RenderImportLocationAssignment", func() {
	var (
		names *terraform_provider.NameProvider
		spec  *properties.Normalization
	)

	BeforeEach(func() {
		spec = &properties.Normalization{
			Name: "TestResource",
		}
		names = &terraform_provider.NameProvider{
			StructName: "TestResource",
		}
	})

	Context("with empty variants", func() {
		It("should return empty string", func() {
			spec.Imports = imports.Import{
				Variants: []string{},
			}

			result, err := terraform_provider.RenderImportLocationAssignment(names, spec, "state", "importVsys")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(BeEmpty())
		})
	})

	Context("with wildcard-only variants", func() {
		It("should generate unconditional locationRequiresImport assignment", func() {
			spec.Imports = imports.Import{
				Variants: []string{"*"},
				Target:   "interface",
			}

			result, err := terraform_provider.RenderImportLocationAssignment(names, spec, "state", "importVsys")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(ContainSubstring("locationRequiresImport := true"))
			Expect(result).ShouldNot(ContainSubstring("var locationRequiresImport bool"))
		})
	})

	Context("with specific variants", func() {
		It("should generate conditional checks for each variant", func() {
			spec.Imports = imports.Import{
				Variants: []string{"layer2", "layer3"},
				Target:   "interface",
			}

			result, err := terraform_provider.RenderImportLocationAssignment(names, spec, "state", "importVsys")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(ContainSubstring("var locationRequiresImport bool"))
			Expect(result).Should(ContainSubstring("state.Layer2"))
			Expect(result).Should(ContainSubstring("state.Layer3"))
			Expect(result).Should(ContainSubstring("locationRequiresImport = true"))
		})
	})

	Context("with mixed variants including wildcard", func() {
		It("should filter out wildcard and generate checks for remaining variants", func() {
			spec.Imports = imports.Import{
				Variants: []string{"layer2", "*"},
				Target:   "interface",
			}

			result, err := terraform_provider.RenderImportLocationAssignment(names, spec, "state", "importVsys")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(ContainSubstring("var locationRequiresImport bool"))
			Expect(result).Should(ContainSubstring("state.Layer2"))
			Expect(result).ShouldNot(ContainSubstring("state.*"))
		})
	})

	Context("source and destination substitution", func() {
		It("should use the provided source and destination values", func() {
			spec.Imports = imports.Import{
				Variants: []string{"*"},
				Target:   "interface",
			}

			result, err := terraform_provider.RenderImportLocationAssignment(names, spec, "state", "importVsys")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(ContainSubstring("state.Location"))
			Expect(result).Should(ContainSubstring("importVsys = terraformInnerLocation.Vsys.ValueString()"))
		})
	})
})
