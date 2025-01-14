package provider_test

import (
	"context"
	"fmt"
	"testing"

	sdkErrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/network/subinterface/ethernet/layer3"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccEthernetLayer3Subinterface_1(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)
	entry := "ethernet1/1.1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             testAccCheckEthernetLayer3SubinterfaceDestroy(entry, prefix),
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

func testAccCheckEthernetLayer3SubinterfaceDestroy(entry string, prefix string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		api := layer3.NewService(sdkClient)
		ctx := context.TODO()

		location := layer3.NewTemplateLocation()
		location.Template.Template = fmt.Sprintf("%s-tmpl", prefix)

		reply, err := api.Read(ctx, *location, entry, "show")
		if err != nil && !sdkErrors.IsObjectNotFound(err) {
			return fmt.Errorf("reading ethernet entry via sdk: %v", err)
		}

		if reply != nil {
			if reply.EntryName() == entry {
				return fmt.Errorf("ethernet object still exists: %s", entry)
			}
		}

		return nil
	}
}
