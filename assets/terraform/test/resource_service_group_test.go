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

func TestAccPanosServiceGroup(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccPanosServiceGroupTmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix": config.StringVariable(prefix),
					"groups": config.MapVariable(map[string]config.Variable{
						"group1": config.ObjectVariable(map[string]config.Variable{
							"tags":    config.ListVariable(config.StringVariable(fmt.Sprintf("%s-tag", prefix))),
							"members": config.ListVariable(config.StringVariable(fmt.Sprintf("%s-svc", prefix))),
						}),
						"group2": config.ObjectVariable(map[string]config.Variable{
							"members": config.ListVariable(),
						}),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_service_group.group1",
						tfjsonpath.New("tags"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact(fmt.Sprintf("%s-tag", prefix)),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_service_group.group1",
						tfjsonpath.New("members"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact(fmt.Sprintf("%s-svc", prefix)),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_service_group.group2",
						tfjsonpath.New("members"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

const testAccPanosServiceGroupTmpl = `
variable "prefix" { type = string }
variable "groups" {
  type = map(object({
    tags = optional(list(string)),
    members = optional(list(string)),
  }))
}

resource "panos_service" "svc" {
  location = { shared = {} }

  name = format("%s-svc", var.prefix)
  protocol = { tcp = { source_port = 80, destination_port = 443 }}
}

resource "panos_administrative_tag" "tag" {
  location = { shared = {} }

  name = format("%s-tag", var.prefix)
}

resource "panos_service_group" "group1" {
  depends_on = [
    resource.panos_service.svc,
    resource.panos_administrative_tag.tag
  ]
  location = { shared = {} }

  name = format("%s-group1", var.prefix)
  members = var.groups["group1"].members
  tags = var.groups["group1"].tags
}

resource "panos_service_group" "group2" {
  location = { shared = {} }

  name = format("%s-group2", var.prefix)
  members = var.groups["group2"].members
}
`
