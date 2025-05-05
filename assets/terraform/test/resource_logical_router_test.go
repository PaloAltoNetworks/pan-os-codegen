package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	//"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccPanosLogicalRouter_1(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: panosLogicalRouterTmpl1,
				ConfigVariables: map[string]config.Variable{
					"prefix": config.StringVariable(prefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_logical_router.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("%s-router", prefix)),
					),
				},
			},
		},
	})
}

const panosLogicalRouterTmpl1 = `
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = format("%s-tmpl", var.prefix)
}

resource "panos_logical_router" "example" {
  location = { template = { name = panos_template.example.name } }

  name = format("%s-router", var.prefix)
}
`
