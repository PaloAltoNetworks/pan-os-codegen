package properties_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

var _ = Describe("NameVariant", func() {
	Context("When creating name variant from empty string", func() {
		It("should return empty string for all variants", func() {
			variant := properties.NewNameVariant("")
			Expect(variant).To(Equal(&properties.NameVariant{
				Original:       "",
				LowerCamelCase: "",
				CamelCase:      "",
				Underscore:     "",
			}))
		})
	})
})
