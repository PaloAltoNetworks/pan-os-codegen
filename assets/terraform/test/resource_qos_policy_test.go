package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/PaloAltoNetworks/pango/policies/rules/qos"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const qosPolicy_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)
    description = "QoS policy test rule"

    source_zones = ["any"]
    source_addresses = ["any"]
    source_users = ["any"]
    negate_source = false

    destination_zones = ["any"]
    destination_addresses = ["any"]
    negate_destination = false

    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    disabled = false
  }]
}
`

func TestAccQosPolicy_Basic(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":                  knownvalue.StringExact(fmt.Sprintf("%s-rule", prefix)),
							"description":           knownvalue.StringExact("QoS policy test rule"),
							"source_zones":          knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"source_addresses":      knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"source_users":          knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"negate_source":         knownvalue.Bool(false),
							"destination_zones":     knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"destination_addresses": knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"negate_destination":    knownvalue.Bool(false),
							"applications":          knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"services":              knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("any")}),
							"action":                knownvalue.ObjectExact(map[string]knownvalue.Check{"class": knownvalue.StringExact("3")}),
							"disabled":              knownvalue.Bool(false),
							"category":              knownvalue.Null(),
							"destination_hip":       knownvalue.Null(),
							"dscp_tos":              knownvalue.Null(),
							"group_tag":             knownvalue.Null(),
							"schedule":              knownvalue.Null(),
							"source_hip":            knownvalue.Null(),
							"tag":                   knownvalue.Null(),
							"target":                knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Any_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      any = {}
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Any(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Any_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"any":        knownvalue.ObjectExact(map[string]knownvalue.Check{}),
							"codepoints": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Codepoints_Ef_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      codepoints = [
        {
          name = "ef-codepoint"
          ef = {
            codepoint = "ef"
          }
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Codepoints_Ef(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Codepoints_Ef_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("ef-codepoint"),
							"ef":     knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("ef")}),
							"af":     knownvalue.Null(),
							"cs":     knownvalue.Null(),
							"tos":    knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Codepoints_Af_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      codepoints = [
        {
          name = "af-codepoint"
          af = {
            codepoint = "af11"
          }
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Codepoints_Af(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Codepoints_Af_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("af-codepoint"),
							"af":     knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("af11")}),
							"ef":     knownvalue.Null(),
							"cs":     knownvalue.Null(),
							"tos":    knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Codepoints_Multiple_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      codepoints = [
        {
          name = "ef-codepoint"
          ef = {
            codepoint = "ef"
          }
        },
        {
          name = "af-codepoint"
          af = {
            codepoint = "af11"
          }
        },
        {
          name = "cs-codepoint"
          cs = {
            codepoint = "cs1"
          }
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Codepoints_Multiple(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Codepoints_Multiple_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("ef-codepoint"),
							"ef":     knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("ef")}),
							"af":     knownvalue.Null(),
							"cs":     knownvalue.Null(),
							"tos":    knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(1),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("af-codepoint"),
							"af":     knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("af11")}),
							"ef":     knownvalue.Null(),
							"cs":     knownvalue.Null(),
							"tos":    knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(2),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("cs-codepoint"),
							"cs":     knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("cs1")}),
							"ef":     knownvalue.Null(),
							"af":     knownvalue.Null(),
							"tos":    knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Codepoints_Cs_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      codepoints = [
        {
          name = "cs-codepoint"
          cs = {
            codepoint = "cs2"
          }
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Codepoints_Cs(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Codepoints_Cs_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("cs-codepoint"),
							"cs":     knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("cs2")}),
							"ef":     knownvalue.Null(),
							"af":     knownvalue.Null(),
							"tos":    knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Codepoints_Tos_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      codepoints = [
        {
          name = "tos-codepoint"
          tos = {
            codepoint = "cs3"
          }
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Codepoints_Tos(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Codepoints_Tos_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":   knownvalue.StringExact("tos-codepoint"),
							"tos":    knownvalue.ObjectExact(map[string]knownvalue.Check{"codepoint": knownvalue.StringExact("cs3")}),
							"ef":     knownvalue.Null(),
							"af":     knownvalue.Null(),
							"cs":     knownvalue.Null(),
							"custom": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_DscpTos_Codepoints_Custom_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    dscp_tos = {
      codepoints = [
        {
          name = "custom-codepoint"
          custom = {
            codepoint = {
              name = "my-custom-cp"
              value = "101010"
            }
          }
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_DscpTos_Codepoints_Custom(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_DscpTos_Codepoints_Custom_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("dscp_tos").AtMapKey("codepoints").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name": knownvalue.StringExact("custom-codepoint"),
							"custom": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"codepoint": knownvalue.ObjectExact(map[string]knownvalue.Check{
									"name":  knownvalue.StringExact("my-custom-cp"),
									"value": knownvalue.StringExact("101010"),
								}),
							}),
							"ef":  knownvalue.Null(),
							"af":  knownvalue.Null(),
							"cs":  knownvalue.Null(),
							"tos": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_Target_Devices_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "serial_number" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_firewall_device" "example" {
  location = { panorama = {} }
  name = var.serial_number
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example, panos_firewall_device.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    target = {
      devices = [
        {
          name = panos_firewall_device.example.name
          vsys = [
            {
              name = "vsys1"
            }
          ]
        }
      ]
    }
  }]
}
`

func TestAccQosPolicy_Target_Devices(t *testing.T) {
	t.Skip("Requires actual managed firewall devices in Panorama")
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)
	suffix := acctest.RandStringFromCharSet(13, "0123456789")
	serialNumber := fmt.Sprintf("00%s", suffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_Target_Devices_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":        config.StringVariable(prefix),
					"location":      location,
					"serial_number": config.StringVariable(serialNumber),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("target"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"devices": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"name": knownvalue.StringExact(serialNumber),
									"vsys": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"name": knownvalue.StringExact("vsys1"),
										}),
									}),
								}),
							}),
							"negate": knownvalue.Null(),
							"tags":   knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_Target_Tags_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    target = {
      tags = ["tag1", "tag2"]
    }
  }]
}
`

func TestAccQosPolicy_Target_Tags(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_Target_Tags_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("target"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"tags": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("tag1"),
								knownvalue.StringExact("tag2"),
							}),
							"devices": knownvalue.Null(),
							"negate":  knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_Target_Negate_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "serial_number" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_firewall_device" "example" {
  location = { panorama = {} }
  name = var.serial_number
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example, panos_firewall_device.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)

    source_zones = ["any"]
    source_addresses = ["any"]
    destination_zones = ["any"]
    destination_addresses = ["any"]
    applications = ["any"]
    services = ["any"]

    action = {
      class = "3"
    }

    target = {
      devices = [
        {
          name = panos_firewall_device.example.name
          vsys = [
            {
              name = "vsys1"
            }
          ]
        }
      ]
      negate = true
    }
  }]
}
`

func TestAccQosPolicy_Target_Negate(t *testing.T) {
	t.Skip("Requires actual managed firewall devices in Panorama")
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)
	suffix := acctest.RandStringFromCharSet(13, "0123456789")
	serialNumber := fmt.Sprintf("00%s", suffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: qosPolicy_Target_Negate_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":        config.StringVariable(prefix),
					"location":      location,
					"serial_number": config.StringVariable(serialNumber),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("target"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"devices": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"name": knownvalue.StringExact(serialNumber),
									"vsys": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"name": knownvalue.StringExact("vsys1"),
										}),
									}),
								}),
							}),
							"negate": knownvalue.Bool(true),
							"tags":   knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const qosPolicy_SetBehavior_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name = format("%s-dg", var.prefix)
  templates = [panos_template.example.name]
}

resource "panos_qos_policy" "example" {
  depends_on = [panos_device_group.example]
  location = var.location

  rules = [{
    name = format("%s-rule", var.prefix)
    description = "Testing set behavior for destination_addresses"

    source_zones = ["any"]
    source_addresses = ["any"]
    source_users = ["any"]
    negate_source = false

    destination_zones = ["any"]
    destination_addresses = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
    negate_destination = false

    applications = ["any"]
    services = ["any"]

    action = {
      class = "4"
    }

    disabled = false
  }]
}
`

// panosQosPolicyReorderDestinationAddresses modifies a QoS policy rule on the server
// by reordering its destination_addresses field. Used to test set behavior.
func panosQosPolicyReorderDestinationAddresses(prefix string, ruleName string, reorderedAddresses []string) {
	svc := qos.NewService(sdkClient)

	location := qos.NewDeviceGroupLocation()
	location.DeviceGroup.DeviceGroup = fmt.Sprintf("%s-dg", prefix)
	location.DeviceGroup.Rulebase = "pre-rulebase"

	// Construct xpath using XpathWithComponents and util.AsEntryXpath
	path, err := location.XpathWithComponents(sdkClient.Versioning(), util.AsEntryXpath(ruleName))
	if err != nil {
		panic(fmt.Sprintf("Failed to build xpath for QoS rule '%s': %v", ruleName, err))
	}
	xpath := util.AsXpath(path)

	// Read current rule using xpath
	entry, err := svc.ReadWithXpath(context.TODO(), xpath, "get")
	if err != nil {
		panic(fmt.Sprintf("Failed to read QoS rule '%s': %v", ruleName, err))
	}

	// Modify destination addresses
	entry.Destination = reorderedAddresses

	// Update on server using xpath
	err = svc.UpdateWithXpath(context.TODO(), xpath, entry, ruleName)
	if err != nil {
		panic(fmt.Sprintf("Failed to update QoS rule '%s': %v", ruleName, err))
	}
}

func TestAccQosPolicy_SetBehavior(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)
	ruleName := fmt.Sprintf("%s-rule", prefix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			// Step 1: Create with ordered addresses
			{
				Config: qosPolicy_SetBehavior_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify addresses using SetExact (order-independent)
					statecheck.ExpectKnownValue(
						"panos_qos_policy.example",
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("destination_addresses"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("10.0.1.0/24"),
							knownvalue.StringExact("10.0.2.0/24"),
							knownvalue.StringExact("10.0.3.0/24"),
						}),
					),
				},
			},
			// Step 2: Reorder via SDK, verify no drift
			{
				Config: qosPolicy_SetBehavior_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				PreConfig: func() {
					panosQosPolicyReorderDestinationAddresses(
						prefix,
						ruleName,
						[]string{"10.0.3.0/24", "10.0.1.0/24", "10.0.2.0/24"},
					)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccQosPolicy_SetBehavior_DetectsChanges(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)
	ruleName := fmt.Sprintf("%s-rule", prefix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name":     config.StringVariable(fmt.Sprintf("%s-dg", prefix)),
			"rulebase": config.StringVariable("pre-rulebase"),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			// Step 1: Create with 3 addresses
			{
				Config: qosPolicy_SetBehavior_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
			},
			// Step 2: Change content (not just order), verify drift detected
			{
				Config: qosPolicy_SetBehavior_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				PreConfig: func() {
					// Replace one address with different IP
					panosQosPolicyReorderDestinationAddresses(
						prefix,
						ruleName,
						[]string{"10.0.4.0/24", "10.0.1.0/24", "10.0.2.0/24"}, // Changed first address
					)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Plan SHOULD show changes
					},
				},
			},
		},
	})
}
