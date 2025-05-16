package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccEthernetLayer3Subinterface_1(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: ethernetLayer3Subinterface_1,
				ConfigVariables: map[string]config.Variable{
					"prefix": config.StringVariable(prefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_ethernet_layer3_subinterface.subinterface",
						tfjsonpath.New("name"),
						knownvalue.StringExact("ethernet1/1.1"),
					),
					statecheck.ExpectKnownValue(
						"panos_ethernet_layer3_subinterface.subinterface",
						tfjsonpath.New("tag"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"panos_ethernet_layer3_subinterface.subinterface",
						tfjsonpath.New("parent"),
						knownvalue.StringExact("ethernet1/1"),
					),
				},
			},
			{
				Config: ethernetLayer3Subinterface_1,
				ConfigVariables: map[string]config.Variable{
					"prefix": config.StringVariable(prefix),
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

const ethernetLayer3Subinterface_1 = `
variable "prefix" { type = string }

locals {
  template_name = "${var.prefix}-tmpl"
}

resource "panos_template" "template" {
  location = { panorama = {} }
  name = local.template_name
}

resource "panos_ethernet_interface" "interface" {
  location = {
    template = {
      vsys = "vsys1"
      name = local.template_name
    }
  }

  name = "ethernet1/1"
  layer3 = {}
}

resource "panos_ethernet_layer3_subinterface" "subinterface" {
  location = {
    template = {
      vsys = "vsys1"
      name = local.template_name
    }
  }

  parent = resource.panos_ethernet_interface.interface.name
  name = "ethernet1/1.1"
  tag = 1
}
`
