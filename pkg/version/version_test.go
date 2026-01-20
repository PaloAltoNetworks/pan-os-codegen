package version_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Version", func() {
	Context("when version is unmarshalled from yaml", func() {
		Context("with invalid yaml data", func() {
			It("should return a wrapped error", func() {
				data := "-"
				var v version.Version
				err := yaml.Unmarshal([]byte(data), &v)
				Expect(err).Should(MatchError(MatchRegexp("failed to unmarshal YAML structure:.+")))
			})
		})
		Context("with an invalid string version", func() {
			var v version.Version
			Context("where not all components are set", func() {

				It("should return an error", func() {
					data := "10"
					err := yaml.Unmarshal([]byte(data), &v)
					Expect(err).Should(MatchError(fmt.Errorf("invalid version string: not enough components")))
				})
			})
			Context("where major component is not a number", func() {
				It("should return an error", func() {
					data := "a.b.c"
					err := yaml.Unmarshal([]byte(data), &v)
					Expect(err).Should(MatchError(MatchRegexp(".*major component must be a number$")))
				})
			})
			Context("where minor component is not a number", func() {
				It("should return an error", func() {
					data := "10.b.c"
					err := yaml.Unmarshal([]byte(data), &v)
					Expect(err).Should(MatchError(MatchRegexp(".*minor component must be a number$")))
				})
			})
			Context("where patch component is not a number", func() {
				It("should return an error", func() {
					data := "10.0.c"
					err := yaml.Unmarshal([]byte(data), &v)
					Expect(err).Should(MatchError(MatchRegexp(".*patch component must be a number$")))
				})
			})
			Context("where hotfix component is empty", func() {
				It("should return an error", func() {
					data := "10.0.0-"
					err := yaml.Unmarshal([]byte(data), &v)
					Expect(err).Should(MatchError(MatchRegexp(".*hotfix part must be set")))
				})
			})
			Context("where patch component is not a number, and there is hotfix component", func() {
				It("should return an error", func() {
					data := "10.0.a-hotfix1"
					err := yaml.Unmarshal([]byte(data), &v)
					Expect(err).Should(MatchError(MatchRegexp(".*patch component must be a number")))
				})
			})
		})
		Context("with a hotfix", func() {
			It("should have a proper struct representation", func() {
				data := "10.0.0-hotfix1"
				var v version.Version
				err := yaml.Unmarshal([]byte(data), &v)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(v).To(Equal(version.Version{Major: 10, Minor: 0, Patch: 0, Hotfix: "hotfix1"}))
			})
		})
		Context("without a hotfix", func() {
			It("should have a proper struct representation", func() {
				data := "10.0.0"
				var v version.Version
				err := yaml.Unmarshal([]byte(data), &v)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(v).To(Equal(version.Version{Major: 10, Minor: 0, Patch: 0}))
			})
		})
	})
	Context("when version is rendered to a string", func() {
		Context("with a hotfix", func() {
			It("should render to a valid string representation", func() {
				expected := "10.0.0-hotfix1"
				v := version.Version{Major: 10, Minor: 0, Patch: 0, Hotfix: "hotfix1"}
				Expect(v.String()).To(Equal(expected))
			})
		})
		Context("wthout a hotfix", func() {
			It("should render to a valid string representation", func() {
				expected := "10.0.0"
				v := version.Version{Major: 10, Minor: 0, Patch: 0}
				Expect(v.String()).To(Equal(expected))
			})
		})
	})
	Context("When comparing version against another", func() {
		this := version.Version{Major: 10, Minor: 0, Patch: 0}
		Context("where both versions are missing hotfix", func() {
			other := version.Version{Major: 10, Minor: 0, Patch: 0}
			It("version should be EqualTo other version", func() {
				Expect(this.EqualTo(other)).To(BeTrue())
			})
			It("version should be GreaterThanOrEqualTo other version", func() {
				Expect(this.GreaterThanOrEqualTo(other)).To(BeTrue())
			})
			It("version should be LesserThanOrEqualTo other version", func() {
				Expect(this.LesserThanOrEqualTo(other)).To(BeTrue())
			})
			It("version should not be GreaterThan other version", func() {
				Expect(this.GreaterThan(other)).To(BeFalse())
			})
			It("version should not be LesserThan other version", func() {
				Expect(this.LesserThan(other)).To(BeFalse())
			})
		})
		Context("where one of versions has a hotfix", func() {
			other := version.Version{Major: 10, Minor: 0, Patch: 0, Hotfix: "hotfix1"}
			It("version should be EqualTo other version", func() {
				Expect(this.EqualTo(other)).To(BeTrue())
			})
			It("version should be GreaterThanOrEqualTo other version", func() {
				Expect(this.GreaterThanOrEqualTo(other)).To(BeTrue())
			})
			It("version should be LesserThanOrEqualTo other version", func() {
				Expect(this.LesserThanOrEqualTo(other)).To(BeTrue())
			})
			It("version should not be GreaterThan other version", func() {
				Expect(this.GreaterThan(other)).To(BeFalse())
			})
			It("version should not be LesserThan other version", func() {
				Expect(this.LesserThan(other)).To(BeFalse())
			})
		})
	})
})
