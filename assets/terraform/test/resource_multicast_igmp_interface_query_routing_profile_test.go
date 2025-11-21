package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMulticastIgmpInterfaceQueryRoutingProfile_Basic(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"template": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: multicastIgmpInterfaceQueryRoutingProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_igmp_interface_query_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_igmp_interface_query_routing_profile.example",
						tfjsonpath.New("immediate_leave"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_igmp_interface_query_routing_profile.example",
						tfjsonpath.New("last_member_query_interval"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_igmp_interface_query_routing_profile.example",
						tfjsonpath.New("max_query_response_time"),
						knownvalue.Int64Exact(15),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_igmp_interface_query_routing_profile.example",
						tfjsonpath.New("query_interval"),
						knownvalue.Int64Exact(200),
					),
				},
			},
		},
	})
}

const multicastIgmpInterfaceQueryRoutingProfile_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_igmp_interface_query_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  immediate_leave = true
  last_member_query_interval = 5
  max_query_response_time = 15
  query_interval = 200
}
`
