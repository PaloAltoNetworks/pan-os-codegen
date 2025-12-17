package translate

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
	"github.com/stretchr/testify/assert"
)

const addressSpecPath = "../../specs/objects/address.yaml"

func TestLocationType(t *testing.T) {
	sampleSpec, err := os.ReadFile(addressSpecPath)
	assert.Nil(t, err, "failed to read address spec")
	// given
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))

	locationKeys := []string{"device-group", "shared"}
	locations := yamlParsedData.Locations
	var locationTypes []string

	// when
	for _, locationKey := range locationKeys {
		locationTypes = append(locationTypes, LocationType(locations[locationKey], true))
	}

	// then
	assert.NotEmpty(t, locationTypes)
	assert.Contains(t, locationTypes, "*DeviceGroupLocation")
	assert.Contains(t, locationTypes, "*SharedLocation")
}

func TestSpecParamType(t *testing.T) {
	// given
	paramTypeRequiredString := properties.SpecParam{
		Type:     "string",
		Required: true,
	}
	itemsForParam := properties.SpecParamItems{
		Type: "string",
	}
	paramTypeListString := properties.SpecParam{
		Type:  "list",
		Items: &itemsForParam,
	}
	paramTypeOptionalString := properties.SpecParam{
		Type: "string",
	}

	// when
	calculatedTypeRequiredString := SpecParamType("", &paramTypeRequiredString)
	calculatedTypeListString := SpecParamType("", &paramTypeListString)
	calculatedTypeOptionalString := SpecParamType("", &paramTypeOptionalString)

	// then
	assert.Equal(t, "string", calculatedTypeRequiredString)
	assert.Equal(t, "[]string", calculatedTypeListString)
	assert.Equal(t, "*string", calculatedTypeOptionalString)
}

func TestOmitEmpty(t *testing.T) {
	sampleSpec, err := os.ReadFile(addressSpecPath)
	assert.Nil(t, err, "failed to read address spec")

	// given
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))
	locationKeys := []string{"device-group", "shared"}
	locations := yamlParsedData.Locations
	var omitEmptyLocations []string

	// when
	for _, locationKey := range locationKeys {
		omitEmptyLocations = append(omitEmptyLocations, OmitEmpty(locations[locationKey]))
	}

	// then
	assert.NotEmpty(t, omitEmptyLocations)
	assert.Contains(t, omitEmptyLocations, ",omitempty")
	assert.Contains(t, omitEmptyLocations, "")
}

var _ = Describe("ParamType", func() {
	Context("When generating XML types for basic types", func() {
		structTyp := structXmlType
		parent := properties.NewNameVariant("")
		suffix := ""
		Context("when parameter is not a list", func() {
			It("should return a normal go type for required params", func() {
				param := &properties.SpecParam{
					Name:     properties.NewNameVariant("test-param"),
					Required: true,
					Type:     "string",
				}
				Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("string"))
			})
			It("should return a pointer to a go type for optional params", func() {
				param := &properties.SpecParam{
					Name:     properties.NewNameVariant("test-param"),
					Required: false,
					Type:     "string",
				}
				Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("string"))
			})
			It("should return string as type for properties with bool type", func() {
				param := &properties.SpecParam{
					Name:     properties.NewNameVariant("test-param"),
					Required: false,
					Type:     "bool",
				}
				Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("string"))
			})
		})
		Context("when parameter is a list", func() {
			It("should return a custom pango type for member lists", func() {
				param := &properties.SpecParam{
					Name:     properties.NewNameVariant("test-param"),
					Required: true,
					Type:     "list",
					Profiles: []*properties.SpecParamProfile{{Type: "member"}},
					Items: &properties.SpecParamItems{
						Type: "string",
					},
				}
				Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("util.Member"))
			})
			It("should return a custom pango type for optional member lists", func() {
				param := &properties.SpecParam{
					Name:     properties.NewNameVariant("test-param"),
					Required: false,
					Type:     "list",
					Profiles: []*properties.SpecParamProfile{{Type: "member"}},
					Items: &properties.SpecParamItems{
						Type: "string",
					},
				}
				Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("util.Member"))
			})
		})
	})
	Context("When generating XML types for custom types", func() {
		structTyp := structXmlType
		Context("when parent is an empty name", func() {
			parent := properties.NewNameVariant("")
			Context("and suffix is empty", func() {
				suffix := ""
				It("should return a correct XML custom type", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "",
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("testParamXml"))
				})
			})
			Context("and suffix is non-empty", func() {
				suffix := "_10_2"
				It("should return a correct XML custom type", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "",
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("testParamXml_10_2"))
				})
			})
		})
		Context("when parent is non-empty name", func() {
			parent := properties.NewNameVariant("parent-param")
			Context("and suffix is empty", func() {
				suffix := ""
				It("should return a correct XML custom type prefixed with parent name", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "",
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("parentParamTestParamXml"))
				})
			})
			Context("and suffix is non-empty", func() {
				suffix := "_10_2"
				It("should return a correct XML custom type", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "",
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("parentParamTestParamXml_10_2"))
				})
			})
		})
	})
	Context("When generating XML types for custom type lists", func() {
		structTyp := structXmlType
		Context("when parent is an empty name", func() {
			parent := properties.NewNameVariant("")
			Context("and suffix is empty", func() {
				suffix := ""
				It("should return a correct XML custom type", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "list",
						Profiles: []*properties.SpecParamProfile{{Type: "entry"}},
						Items: &properties.SpecParamItems{
							Type: "object",
						},
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("testParamXml"))
				})
			})
			Context("and suffix is non-empty", func() {
				suffix := "_10_2"
				It("should return a correct XML custom type", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "list",
						Profiles: []*properties.SpecParamProfile{{Type: "entry"}},
						Items: &properties.SpecParamItems{
							Type: "object",
						},
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("testParamXml_10_2"))
				})
			})
		})
		Context("when parent is non-empty name", func() {
			parent := properties.NewNameVariant("parent-param")
			Context("and suffix is empty", func() {
				suffix := ""
				It("should return a correct XML custom type prefixed with parent name", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "list",
						Profiles: []*properties.SpecParamProfile{{Type: "entry"}},
						Items: &properties.SpecParamItems{
							Type: "object",
						},
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("parentParamTestParamXml"))
				})
			})
			Context("and suffix is non-empty", func() {
				suffix := "_10_2"
				It("should return a correct XML custom type", func() {
					param := &properties.SpecParam{
						Name:     properties.NewNameVariant("test-param"),
						Required: true,
						Type:     "list",
						Profiles: []*properties.SpecParamProfile{{Type: "entry"}},
						Items: &properties.SpecParamItems{
							Type: "object",
						},
					}
					Expect(ParamType(structTyp, parent, param, suffix)).To(Equal("parentParamTestParamXml_10_2"))
				})
			})
		})
	})
})

var _ = Describe("createEntryXmlStructSpecsForParameter", func() {
	Context("when creating struct context for a parameter", func() {
		It("should return a proper struct context", func() {
			parent := properties.NewNameVariant("parent-param")
			param := &properties.SpecParam{
				Name:     properties.NewNameVariant("child-param"),
				Type:     "",
				Profiles: []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"child-param"}}},
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"grandchild-param": {
							Name:     properties.NewNameVariant("grandchild-param"),
							Type:     "",
							Profiles: []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"grandchild-param"}}},
							Spec:     &properties.Spec{},
						},
					},
				},
			}

			result := createEntryXmlStructSpecsForParameter(structXmlType, parent, param, nil)
			Expect(result).To(HaveLen(2))

			Expect(result[0].StructName()).To(Equal("ParentParamChildParam"))
			Expect(result[0].XmlStructName()).To(Equal("parentParamChildParamXml"))
			Expect(result[0].Fields[0]).To(Equal(entryStructFieldContext{
				Name:      properties.NewNameVariant("grandchild-param"),
				FieldType: "object",
				Type:      "ParentParamChildParamGrandchildParam",
				XmlType:   "parentParamChildParamGrandchildParamXml",
				Tags:      "`xml:\"grandchild-param,omitempty\"`",
			}))
			Expect(result[0].Fields[0].FinalType()).To(Equal("*ParentParamChildParamGrandchildParam"))
			Expect(result[0].Fields[0].FinalXmlType()).To(Equal("*parentParamChildParamGrandchildParamXml"))
			Expect(result[1].StructName()).To(Equal("ParentParamChildParamGrandchildParam"))
			Expect(result[1].XmlStructName()).To(Equal("parentParamChildParamGrandchildParamXml"))
		})
	})
	Context("when creating struct context for a simple type list parameter", func() {
		It("should return a proper struct context", func() {
			parent := properties.NewNameVariant("")
			param := &properties.SpecParam{
				Name:     properties.NewNameVariant("parent-param"),
				Type:     "",
				Profiles: []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"child-param"}}},
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"child-param-string": {
							Name:      properties.NewNameVariant("child-param-string"),
							SpecOrder: 0,
							Type:      "list",
							Profiles:  []*properties.SpecParamProfile{{Type: "member", Xpath: []string{"child-param-string"}}},
							Items: &properties.SpecParamItems{
								Type: "string",
							},
						},
						"child-param-int64": {
							Name:      properties.NewNameVariant("child-param-int64"),
							SpecOrder: 1,
							Type:      "list",
							Profiles:  []*properties.SpecParamProfile{{Type: "member", Xpath: []string{"child-param-int64"}}},
							Items: &properties.SpecParamItems{
								Type: "int64",
							},
						},
					},
				},
			}

			result := createEntryXmlStructSpecsForParameter(structXmlType, parent, param, nil)
			Expect(result).To(HaveLen(1))

			Expect(result[0].StructName()).To(Equal("ParentParam"))
			Expect(result[0].XmlStructName()).To(Equal("parentParamXml"))
			Expect(result[0].Fields[0]).To(Equal(entryStructFieldContext{
				Name:         properties.NewNameVariant("child-param-string"),
				FieldType:    "list-member",
				Type:         "string",
				ItemsType:    "[]string",
				XmlType:      "util.Member",
				ItemsXmlType: "util.MemberType",
				Tags:         "`xml:\"child-param-string,omitempty\"`",
			}))
			Expect(result[0].Fields[1]).To(Equal(entryStructFieldContext{
				Name:         properties.NewNameVariant("child-param-int64"),
				FieldType:    "list-member",
				Type:         "int64",
				ItemsType:    "[]int64",
				XmlType:      "util.Member",
				ItemsXmlType: "util.MemberType",
				Tags:         "`xml:\"child-param-int64,omitempty\"`",
			}))
		})
	})
	Context("when creating struct context for a complex type list parameter", func() {
		It("should return a proper struct context", func() {
			parent := properties.NewNameVariant("")
			param := &properties.SpecParam{
				Name:     properties.NewNameVariant("parent-param"),
				Type:     "",
				Profiles: []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"child-param"}}},
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"child-param": {
							Name:      properties.NewNameVariant("child-param"),
							SpecOrder: 0,
							Type:      "list",
							Profiles:  []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"child-param"}}},
							Items: &properties.SpecParamItems{
								Type: "entry",
							},
							Spec: &properties.Spec{
								Params: map[string]*properties.SpecParam{
									"grandchild-param": {
										Name:     properties.NewNameVariant("grandchild-param"),
										Type:     "int64",
										Profiles: []*properties.SpecParamProfile{{Xpath: []string{"grandchild-param"}}},
									},
								},
							},
						},
					},
				},
			}

			result := createEntryXmlStructSpecsForParameter(structXmlType, parent, param, nil)
			Expect(result).To(HaveLen(3))

			Expect(result[0].StructName()).To(Equal("ParentParam"))
			Expect(result[0].XmlStructName()).To(Equal("parentParamXml"))
			Expect(result[0].Fields[0]).To(Equal(entryStructFieldContext{
				Name:             properties.NewNameVariant("child-param"),
				FieldType:        "list-entry",
				Type:             "ParentParamChildParam",
				ItemsType:        "[]ParentParamChildParam",
				XmlType:          "parentParamChildParamXml",
				XmlContainerType: "parentParamChildParamContainerXml",
				ItemsXmlType:     "[]parentParamChildParamXml",
				Tags:             "`xml:\"child-param,omitempty\"`",
			}))
			Expect(result[0].Fields[0].FinalXmlType()).To(Equal("*parentParamChildParamContainerXml"))

			Expect(result[1].IsXmlContainer).To(BeTrue())
			Expect(result[1].XmlStructName()).To(Equal("parentParamChildParamContainerXml"))
			Expect(result[1].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:         properties.NewNameVariant("entries"),
					FieldType:    "list-entry",
					Type:         "ParentParamChildParam",
					ItemsType:    "[]ParentParamChildParam",
					XmlType:      "parentParamChildParamXml",
					ItemsXmlType: "[]parentParamChildParamXml",
					Tags:         "`xml:\"entry\"`",
				},
			}))

			Expect(result[2].StructName()).To(Equal("ParentParamChildParam"))
			Expect(result[2].XmlStructName()).To(Equal("parentParamChildParamXml"))
			Expect(result[2].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:         xmlNameVariant,
					IsInternal:   true,
					Required:     false,
					FieldType:    "internal",
					Type:         "",
					ItemsType:    "",
					XmlType:      "xml.Name",
					ItemsXmlType: "",
					Tags:         "`xml:\"entry\"`",
				},
				{
					Name:         properties.NewNameVariant("name"),
					Required:     true,
					FieldType:    "simple",
					Type:         "string",
					ItemsType:    "",
					XmlType:      "string",
					ItemsXmlType: "",
					Tags:         "`xml:\"name,attr\"`",
				},
				{
					Name:         properties.NewNameVariant("grandchild-param"),
					FieldType:    "simple",
					Type:         "int64",
					ItemsType:    "",
					XmlType:      "int64",
					ItemsXmlType: "",
					Tags:         "`xml:\"grandchild-param,omitempty\"`",
				},
				{
					Name:      properties.NewNameVariant("misc"),
					FieldType: "internal",
					Type:      "[]generic.Xml",
					XmlType:   "[]generic.Xml",
					Tags:      "`xml:\",any\"`",
				},
				{
					Name:      properties.NewNameVariant("misc-attributes"),
					FieldType: "internal",
					Type:      "[]xml.Attr",
					XmlType:   "[]xml.Attr",
					Tags:      "`xml:\",any,attr\"`",
				},
			}))
		})
	})
	Context("when creating struct context for parameter with versioned children", func() {
		parent := properties.NewNameVariant("")
		It("should return all specs when nil version is passed", func() {
			param := &properties.SpecParam{
				Name:     properties.NewNameVariant("parent-param"),
				Type:     "",
				Profiles: []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"child-param"}}},
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"child-param1": {
							Name:      properties.NewNameVariant("child-param1"),
							SpecOrder: 0,
							Type:      "",
							Profiles: []*properties.SpecParamProfile{
								{
									Type:  "entry",
									Xpath: []string{"child-param1"},
								},
							},
							Spec: &properties.Spec{},
						},
						"child-param2": {
							Name:      properties.NewNameVariant("child-param2"),
							SpecOrder: 1,
							Type:      "",
							Profiles: []*properties.SpecParamProfile{
								{
									Type:  "entry",
									Xpath: []string{"child-param2"},
								},
							},
							Spec: &properties.Spec{},
						},
					},
				},
			}

			result := createEntryXmlStructSpecsForParameter(structXmlType, parent, param, nil)
			Expect(result).To(HaveLen(3))

			Expect(result[0].StructName()).To(Equal("ParentParam"))
			Expect(result[0].XmlStructName()).To(Equal("parentParamXml"))
			Expect(result[0].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:      properties.NewNameVariant("child-param1"),
					FieldType: "object",
					Type:      "ParentParamChildParam1",
					XmlType:   "parentParamChildParam1Xml",
					Tags:      "`xml:\"child-param1,omitempty\"`",
				},
				{
					Name:      properties.NewNameVariant("child-param2"),
					FieldType: "object",
					Type:      "ParentParamChildParam2",
					XmlType:   "parentParamChildParam2Xml",
					Tags:      "`xml:\"child-param2,omitempty\"`",
				},
				{
					Name:      properties.NewNameVariant("misc"),
					FieldType: "internal",
					Type:      "[]generic.Xml",
					XmlType:   "[]generic.Xml",
					Tags:      "`xml:\",any\"`",
				},
				{
					Name:      properties.NewNameVariant("misc-attributes"),
					FieldType: "internal",
					Type:      "[]xml.Attr",
					XmlType:   "[]xml.Attr",
					Tags:      "`xml:\",any,attr\"`",
				},
			}))

			Expect(result[0].Fields[0].FinalType()).To(Equal("*ParentParamChildParam1"))
			Expect(result[0].Fields[0].FinalXmlType()).To(Equal("*parentParamChildParam1Xml"))

			Expect(result[0].Fields[1].FinalType()).To(Equal("*ParentParamChildParam2"))
			Expect(result[0].Fields[1].FinalXmlType()).To(Equal("*parentParamChildParam2Xml"))

			Expect(result[1].StructName()).To(Equal("ParentParamChildParam1"))
			Expect(result[1].XmlStructName()).To(Equal("parentParamChildParam1Xml"))
			Expect(result[2].StructName()).To(Equal("ParentParamChildParam2"))
			Expect(result[2].XmlStructName()).To(Equal("parentParamChildParam2Xml"))
		})
		It("should only return relevant specs when version is non-nill", func() {
			param := &properties.SpecParam{
				Name:     properties.NewNameVariant("parent-param"),
				Type:     "",
				Profiles: []*properties.SpecParamProfile{{Type: "entry", Xpath: []string{"child-param"}}},
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"child-param1": {
							Name:      properties.NewNameVariant("child-param1"),
							SpecOrder: 0,
							Type:      "",
							Profiles: []*properties.SpecParamProfile{
								{
									Type:  "entry",
									Xpath: []string{"child-param1"},
								},
							},
							Spec: &properties.Spec{},
						},
						"child-param2": {
							Name:      properties.NewNameVariant("child-param2"),
							SpecOrder: 1,
							Type:      "",
							Profiles: []*properties.SpecParamProfile{
								{
									Type:       "entry",
									Xpath:      []string{"child-param2"},
									MinVersion: version.MustNewVersionFromString("11.0.0"),
									MaxVersion: version.MustNewVersionFromString("11.0.5"),
								},
							},
							Spec: &properties.Spec{},
						},
					},
				},
			}
			paramVersion := version.MustNewVersionFromString("10.0.0")
			result := createEntryXmlStructSpecsForParameter(structXmlType, parent, param, paramVersion)
			Expect(result).To(HaveLen(2))

			Expect(result[0].StructName()).To(Equal("ParentParam"))
			Expect(result[0].XmlStructName()).To(Equal("parentParamXml_10_0_0"))
			Expect(result[0].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:      properties.NewNameVariant("child-param1"),
					FieldType: "object",
					Type:      "ParentParamChildParam1",
					XmlType:   "parentParamChildParam1Xml_10_0_0",
					Tags:      "`xml:\"child-param1,omitempty\"`",
					version:   paramVersion,
				},
				{
					Name:      properties.NewNameVariant("misc"),
					FieldType: "internal",
					Type:      "[]generic.Xml",
					XmlType:   "[]generic.Xml",
					Tags:      "`xml:\",any\"`",
				},
				{
					Name:      properties.NewNameVariant("misc-attributes"),
					FieldType: "internal",
					Type:      "[]xml.Attr",
					XmlType:   "[]xml.Attr",
					Tags:      "`xml:\",any,attr\"`",
				},
			}))

			Expect(result[1].StructName()).To(Equal("ParentParamChildParam1"))
			Expect(result[1].XmlStructName()).To(Equal("parentParamChildParam1Xml_10_0_0"))
		})
	})
})

var _ = Describe("createStructSpecs", func() {
	Context("when generating xml struct specs for a spec without parameters", func() {
		It("should return a list of struct specs with a single element", func() {
			spec := &properties.Normalization{
				TerraformProviderConfig: properties.TerraformProviderConfig{
					ResourceType: properties.TerraformResourceEntry,
				},
				Name: "test-spec",
				Spec: &properties.Spec{},
			}

			result := createStructSpecs(structXmlType, spec, nil)
			Expect(result).To(HaveLen(1))

			Expect(result[0].StructName()).To(Equal("Entry"))
			Expect(result[0].XmlStructName()).To(Equal("entryXml"))
			Expect(result[0].Fields[0].Name.CamelCase).To(Equal("XMLName"))
			Expect(result[0].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:       xmlNameVariant,
					FieldType:  "internal",
					IsInternal: true,
					XmlType:    "xml.Name",
					Tags:       "`xml:\"entry\"`",
				},
				{
					Name:      properties.NewNameVariant("name"),
					Required:  true,
					FieldType: "simple",
					Type:      "string",
					XmlType:   "string",
					Tags:      "`xml:\"name,attr\"`",
				},
				{
					Name:      properties.NewNameVariant("misc"),
					FieldType: "internal",
					Type:      "[]generic.Xml",
					XmlType:   "[]generic.Xml",
					Tags:      "`xml:\",any\"`",
				},
				{
					Name:      properties.NewNameVariant("misc-attributes"),
					FieldType: "internal",
					Type:      "[]xml.Attr",
					XmlType:   "[]xml.Attr",
					Tags:      "`xml:\",any,attr\"`",
				},
			}))
		})
	})
	Context("when generating struct specs for a spec with entry lists", func() {
		spec := &properties.Normalization{
			TerraformProviderConfig: properties.TerraformProviderConfig{
				ResourceType: properties.TerraformResourceEntry,
			},
			Name: "test-spec",
			Spec: &properties.Spec{
				Params: map[string]*properties.SpecParam{
					"param-one": {
						Name: properties.NewNameVariant("param-one"),
						Profiles: []*properties.SpecParamProfile{
							{
								Type:  "entry",
								Xpath: []string{"param-one"},
							},
						},
						Type: "list",
						Items: &properties.SpecParamItems{
							Type: "entry",
						},
						Spec: &properties.Spec{
							Params: map[string]*properties.SpecParam{
								"param-two": {
									Name: properties.NewNameVariant("param-two"),
									Profiles: []*properties.SpecParamProfile{
										{
											Xpath: []string{"param-two"},
										},
									},
									Type: "int64",
								},
							},
						},
					},
				},
			},
		}
		Context("and the struct type is structApiType", func() {
			It("should create a parent spec and a single child spec", func() {
				result := createStructSpecs(structApiType, spec, nil)

				Expect(result).To(HaveLen(2))
				Expect(result[0].name).To(Equal(properties.NewNameVariant("entry")))
				Expect(result[0].Fields).To(Equal([]entryStructFieldContext{
					{
						Name:      properties.NewNameVariant("name"),
						Required:  true,
						FieldType: "simple",
						Type:      "string",
						XmlType:   "string",
						Tags:      "`xml:\"name,attr\"`",
					},
					{
						Name:         properties.NewNameVariant("param-one"),
						FieldType:    "list-entry",
						Type:         "ParamOne",
						ItemsType:    "[]ParamOne",
						XmlType:      "paramOneXml",
						ItemsXmlType: "[]paramOneXml",
						Tags:         "`xml:\"param-one,omitempty\"`",
					},
					{
						Name:      properties.NewNameVariant("misc"),
						FieldType: "internal",
						Type:      "[]generic.Xml",
						XmlType:   "[]generic.Xml",
						Tags:      "`xml:\",any\"`",
					},
					{
						Name:      properties.NewNameVariant("misc-attributes"),
						FieldType: "internal",
						Type:      "[]xml.Attr",
						XmlType:   "[]xml.Attr",
						Tags:      "`xml:\",any,attr\"`",
					},
				}))
				Expect(result[1].name).To(Equal(properties.NewNameVariant("param-one")))
				Expect(result[1].Fields).To(Equal([]entryStructFieldContext{
					{
						Name:      properties.NewNameVariant("name"),
						Required:  true,
						FieldType: "simple",
						Type:      "string",
						XmlType:   "string",
						Tags:      "`xml:\"name,attr\"`",
					},
					{
						Name:         properties.NewNameVariant("param-two"),
						FieldType:    "simple",
						Type:         "int64",
						ItemsType:    "",
						XmlType:      "int64",
						ItemsXmlType: "",
						Tags:         "`xml:\"param-two,omitempty\"`",
					},
					{
						Name:      properties.NewNameVariant("misc"),
						FieldType: "internal",
						Type:      "[]generic.Xml",
						XmlType:   "[]generic.Xml",
						Tags:      "`xml:\",any\"`",
					},
					{
						Name:      properties.NewNameVariant("misc-attributes"),
						FieldType: "internal",
						Type:      "[]xml.Attr",
						XmlType:   "[]xml.Attr",
						Tags:      "`xml:\",any,attr\"`",
					},
				}))
			})
		})
		Context("and the struct type is structXmlType", func() {
			It("should create a parent spec and two child specs", func() {
				result := createStructSpecs(structXmlType, spec, nil)

				Expect(result).To(HaveLen(3))
				Expect(result[0].name).To(Equal(properties.NewNameVariant("entry")))
				Expect(result[0].Fields).To(Equal([]entryStructFieldContext{
					{
						Name:       xmlNameVariant,
						FieldType:  "internal",
						IsInternal: true,
						XmlType:    "xml.Name",
						Tags:       "`xml:\"entry\"`",
					},
					{
						Name:      properties.NewNameVariant("name"),
						Required:  true,
						FieldType: "simple",
						Type:      "string",
						XmlType:   "string",
						Tags:      "`xml:\"name,attr\"`",
					},
					{
						Name:             properties.NewNameVariant("param-one"),
						FieldType:        "list-entry",
						Type:             "ParamOne",
						ItemsType:        "[]ParamOne",
						XmlType:          "paramOneXml",
						XmlContainerType: "paramOneContainerXml",
						ItemsXmlType:     "[]paramOneXml",
						Tags:             "`xml:\"param-one,omitempty\"`",
					},
					{
						Name:      properties.NewNameVariant("misc"),
						FieldType: "internal",
						Type:      "[]generic.Xml",
						XmlType:   "[]generic.Xml",
						Tags:      "`xml:\",any\"`",
					},
					{
						Name:      properties.NewNameVariant("misc-attributes"),
						FieldType: "internal",
						Type:      "[]xml.Attr",
						XmlType:   "[]xml.Attr",
						Tags:      "`xml:\",any,attr\"`",
					},
				}))

				Expect(result[1].name).To(Equal(properties.NewNameVariant("param-one-container")))
				Expect(result[1].Fields).To(Equal([]entryStructFieldContext{
					{
						Name:         properties.NewNameVariant("entries"),
						FieldType:    "list-entry",
						Type:         "ParamOne",
						ItemsType:    "[]ParamOne",
						XmlType:      "paramOneXml",
						ItemsXmlType: "[]paramOneXml",
						Tags:         "`xml:\"entry\"`",
					},
				}))

				Expect(result[2].name).To(Equal(properties.NewNameVariant("param-one")))
				Expect(result[2].Fields).To(Equal([]entryStructFieldContext{
					{
						Name:       xmlNameVariant,
						FieldType:  "internal",
						IsInternal: true,
						XmlType:    "xml.Name",
						Tags:       "`xml:\"entry\"`",
					},
					{
						Name:      properties.NewNameVariant("name"),
						Required:  true,
						FieldType: "simple",
						Type:      "string",
						XmlType:   "string",
						Tags:      "`xml:\"name,attr\"`",
					},
					{
						Name:         properties.NewNameVariant("param-two"),
						FieldType:    "simple",
						Type:         "int64",
						ItemsType:    "",
						XmlType:      "int64",
						ItemsXmlType: "",
						Tags:         "`xml:\"param-two,omitempty\"`",
					},
					{
						Name:      properties.NewNameVariant("misc"),
						FieldType: "internal",
						Type:      "[]generic.Xml",
						XmlType:   "[]generic.Xml",
						Tags:      "`xml:\",any\"`",
					},
					{
						Name:      properties.NewNameVariant("misc-attributes"),
						FieldType: "internal",
						Type:      "[]xml.Attr",
						XmlType:   "[]xml.Attr",
						Tags:      "`xml:\",any,attr\"`",
					},
				}))
			})
		})
	})
	Context("when generating xml struct specs for a spec with one object parameter", func() {
		It("should return a list of two specs", func() {
			version11_0_0 := version.MustNewVersionFromString("11.0.0")
			version11_0_3 := version.MustNewVersionFromString("11.0.3")
			spec := &properties.Normalization{
				TerraformProviderConfig: properties.TerraformProviderConfig{
					ResourceType: properties.TerraformResourceEntry,
				},
				Name: "test-spec",
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"param-one": {
							Name: properties.NewNameVariant("param-one"),
							Profiles: []*properties.SpecParamProfile{
								{
									Type:       "entry",
									Xpath:      []string{"param-one"},
									MinVersion: version11_0_3,
									MaxVersion: version.MustNewVersionFromString("11.0.5"),
								},
							},
							Type: "bool",
						},
					},
					OneOf: map[string]*properties.SpecParam{
						"param-two": {
							Name: properties.NewNameVariant("param-two"),
							Type: "",
							Profiles: []*properties.SpecParamProfile{
								{
									Type:  "entry",
									Xpath: []string{"param-two"},
								},
							},
							Items: &properties.SpecParamItems{
								Type: "object",
							},
							Spec: &properties.Spec{
								Params: map[string]*properties.SpecParam{
									"param-three": {
										Name: properties.NewNameVariant("param-three"),
										Type: "",
										Profiles: []*properties.SpecParamProfile{
											{
												Type:       "entry",
												Xpath:      []string{"param-three"},
												MinVersion: version11_0_0,
												MaxVersion: version.MustNewVersionFromString("11.0.5"),
											},
										},
										Items: &properties.SpecParamItems{
											Type: "object",
										},
										Spec: &properties.Spec{
											Params: map[string]*properties.SpecParam{
												"param-four": {
													Name: properties.NewNameVariant("param-four"),
													Profiles: []*properties.SpecParamProfile{
														{
															Xpath: []string{"param-four"},
														},
													},
													Type: "int64",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := createStructSpecs(structXmlType, spec, nil)
			Expect(result).To(HaveLen(3))

			Expect(result[0].XmlStructName()).To(Equal("entryXml"))
			Expect(result[1].XmlStructName()).To(Equal("paramTwoXml"))
			Expect(result[2].XmlStructName()).To(Equal("paramTwoParamThreeXml"))

			result = createStructSpecs(structXmlType, spec, version11_0_3)
			Expect(result).To(HaveLen(3))
			Expect(result[0].XmlStructName()).To(Equal("entryXml_11_0_3"))
			Expect(result[1].XmlStructName()).To(Equal("paramTwoXml_11_0_3"))
			Expect(result[2].XmlStructName()).To(Equal("paramTwoParamThreeXml_11_0_3"))

			Expect(result[0].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:       xmlNameVariant,
					FieldType:  "internal",
					IsInternal: true,
					XmlType:    "xml.Name",
					Tags:       "`xml:\"entry\"`",
				},
				{
					Name:      properties.NewNameVariant("name"),
					Required:  true,
					FieldType: "simple",
					Type:      "string",
					XmlType:   "string",
					Tags:      "`xml:\"name,attr\"`",
				},
				{
					Name:      properties.NewNameVariant("param-one"),
					FieldType: "simple",
					Type:      "bool",
					XmlType:   "string",
					Tags:      "`xml:\"param-one,omitempty\"`",
					version:   version11_0_3,
				},
				{
					Name:      properties.NewNameVariant("param-two"),
					FieldType: "object",
					Type:      "ParamTwo",
					XmlType:   "paramTwoXml_11_0_3",
					Tags:      "`xml:\"param-two,omitempty\"`",
					version:   version11_0_3,
				},
				{
					Name:      properties.NewNameVariant("misc"),
					FieldType: "internal",
					Type:      "[]generic.Xml",
					XmlType:   "[]generic.Xml",
					Tags:      "`xml:\",any\"`",
				},
				{
					Name:      properties.NewNameVariant("misc-attributes"),
					FieldType: "internal",
					Type:      "[]xml.Attr",
					XmlType:   "[]xml.Attr",
					Tags:      "`xml:\",any,attr\"`",
				},
			}))

			result = createStructSpecs(structXmlType, spec, version11_0_0)
			Expect(result).To(HaveLen(3))
			Expect(result[0].XmlStructName()).To(Equal("entryXml_11_0_0"))
			Expect(result[1].XmlStructName()).To(Equal("paramTwoXml_11_0_0"))
			Expect(result[2].XmlStructName()).To(Equal("paramTwoParamThreeXml_11_0_0"))

			Expect(result[0].Fields).To(HaveExactElements([]entryStructFieldContext{
				{
					Name:       xmlNameVariant,
					FieldType:  "internal",
					IsInternal: true,
					XmlType:    "xml.Name",
					Tags:       "`xml:\"entry\"`",
				},
				{
					Name:      properties.NewNameVariant("name"),
					Required:  true,
					FieldType: "simple",
					Type:      "string",
					XmlType:   "string",
					Tags:      "`xml:\"name,attr\"`",
				},
				{
					Name:      properties.NewNameVariant("param-two"),
					FieldType: "object",
					Type:      "ParamTwo",
					XmlType:   "paramTwoXml_11_0_0",
					Tags:      "`xml:\"param-two,omitempty\"`",
					version:   version11_0_0,
				},
				{
					Name:      properties.NewNameVariant("misc"),
					FieldType: "internal",
					Type:      "[]generic.Xml",
					XmlType:   "[]generic.Xml",
					Tags:      "`xml:\",any\"`",
				},
				{
					Name:      properties.NewNameVariant("misc-attributes"),
					FieldType: "internal",
					Type:      "[]xml.Attr",
					XmlType:   "[]xml.Attr",
					Tags:      "`xml:\",any,attr\"`",
				},
			}))
		})
	})
})

func TestXmlTag(t *testing.T) {
	// given
	paramTypeRequiredString := properties.SpecParam{
		Type:     "string",
		Required: true,
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "string",
				Xpath: []string{"description"},
			},
		},
	}
	paramTypeUuid := properties.SpecParam{
		Type:     "string",
		Required: false,
		Name:     properties.NewNameVariant("uuid"),
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "string",
				Xpath: []string{"uuid"},
			},
		},
	}
	paramTypeListString := properties.SpecParam{
		Type: "list",
		Items: &properties.SpecParamItems{
			Type: "string",
		},
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "member",
				Xpath: []string{"tag"},
			},
		},
	}

	// when
	calculatedXmlTagRequiredString := XmlTag(&paramTypeRequiredString)
	calculatedXmlTagListString := XmlTag(&paramTypeListString)
	calculatedXmlTagUuid := XmlTag(&paramTypeUuid)

	// then
	assert.Equal(t, "`xml:\"description\"`", calculatedXmlTagRequiredString)
	assert.Equal(t, "`xml:\"tag,omitempty\"`", calculatedXmlTagListString)
	assert.Equal(t, "`xml:\"uuid,attr,omitempty\"`", calculatedXmlTagUuid)
}

func TestCreateGoSuffixFromVersion(t *testing.T) {
	// given

	// when
	version, _ := version.NewVersionFromString("10.1.1")
	suffix := CreateGoSuffixFromVersion(&version)

	// then
	assert.Equal(t, "_10_1_1", suffix)
}

func TestParamSupportedInVersion(t *testing.T) {
	// given
	deviceVersion110, _ := version.NewVersionFromString("11.0.0")
	deviceVersion101, _ := version.NewVersionFromString("10.1.1")
	deviceVersion90, _ := version.NewVersionFromString("9.0.3")

	paramName := properties.NewNameVariant("test")

	profileAlwaysPresent := properties.SpecParamProfile{
		Xpath: []string{"test"},
	}

	minVersion1010, _ := version.NewVersionFromString("10.1.0")
	maxVersion1013, _ := version.NewVersionFromString("10.1.3")
	profilePresentFrom10 := properties.SpecParamProfile{
		Xpath:      []string{"test"},
		MinVersion: &minVersion1010,
		MaxVersion: &maxVersion1013,
	}

	minVersion1100, _ := version.NewVersionFromString("11.0.0")
	maxVersion1110, _ := version.NewVersionFromString("11.1.0")
	profilePresentFrom11 := properties.SpecParamProfile{
		Xpath:      []string{"test"},
		MinVersion: &minVersion1100,
		MaxVersion: &maxVersion1110,
	}

	minVersion901, _ := version.NewVersionFromString("9.0.1")
	maxVersion910, _ := version.NewVersionFromString("9.1.0")
	profileNotPresentFrom10 := properties.SpecParamProfile{
		Xpath:      []string{"test"},
		MinVersion: &minVersion901,
		MaxVersion: &maxVersion910,
	}

	paramPresentFrom10 := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profilePresentFrom10,
		},
	}
	paramAlwaysPresent := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profileAlwaysPresent,
		},
	}
	paramNotPresentFrom10 := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profileNotPresentFrom10,
		},
	}

	paramPresentFrom10And11 := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profilePresentFrom10,
			&profilePresentFrom11,
		},
	}

	// when
	noVersionAndParamAlwaysPresent := ParamSupportedInVersion(paramAlwaysPresent, nil)
	noVersionAndParamNotPresentFrom10 := ParamSupportedInVersion(paramNotPresentFrom10, nil)
	device10AndParamPresentFrom10 := ParamSupportedInVersion(paramPresentFrom10, &deviceVersion101)
	device10AndParamAlwaysPresent := ParamSupportedInVersion(paramAlwaysPresent, &deviceVersion101)
	device10AndParamNotPresentFrom10 := ParamSupportedInVersion(paramNotPresentFrom10, &deviceVersion101)
	device9AndParamPresentFrom10 := ParamSupportedInVersion(paramPresentFrom10, &deviceVersion90)
	device9AndParamAlwaysPresent := ParamSupportedInVersion(paramAlwaysPresent, &deviceVersion90)
	device9AndParamNotPresentFrom10 := ParamSupportedInVersion(paramNotPresentFrom10, &deviceVersion90)
	assert.True(t, ParamSupportedInVersion(paramPresentFrom10And11, &deviceVersion110))

	// then
	assert.True(t, noVersionAndParamAlwaysPresent)
	assert.True(t, noVersionAndParamNotPresentFrom10)
	assert.True(t, device10AndParamPresentFrom10)
	assert.True(t, device10AndParamAlwaysPresent)
	assert.False(t, device10AndParamNotPresentFrom10)
	assert.False(t, device9AndParamPresentFrom10)
	assert.True(t, device9AndParamAlwaysPresent)
	assert.True(t, device9AndParamNotPresentFrom10)
}
