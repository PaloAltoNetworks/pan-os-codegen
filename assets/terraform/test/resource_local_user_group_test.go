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

func TestAccLocalUserGroup_Basic(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"template_vsys": config.ObjectVariable(map[string]config.Variable{
			"template": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			// Step 1: Create with initial users
			{
				Config: localUserGroup_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
					"users":    config.ListVariable(config.StringVariable("user1"), config.StringVariable("user2")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_local_user_group.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_local_user_group.example",
						tfjsonpath.New("users"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("user1"),
							knownvalue.StringExact("user2"),
						}),
					),
				},
			},
			// Step 2: Update users list - add user3
			{
				Config: localUserGroup_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
					"users":    config.ListVariable(config.StringVariable("user1"), config.StringVariable("user2"), config.StringVariable("user3")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_local_user_group.example",
						tfjsonpath.New("users"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("user1"),
							knownvalue.StringExact("user2"),
							knownvalue.StringExact("user3"),
						}),
					),
				},
			},
			// Step 3: Update users list - remove user1
			{
				Config: localUserGroup_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
					"users":    config.ListVariable(config.StringVariable("user2"), config.StringVariable("user3")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_local_user_group.example",
						tfjsonpath.New("users"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("user2"),
							knownvalue.StringExact("user3"),
						}),
					),
				},
			},
		},
	})
}

const localUserGroup_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "users" { type = list(string) }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_local_user" "user1" {
  depends_on = [panos_template.example]
  location = var.location
  name     = "user1"
  password = "Password123!"
}

resource "panos_local_user" "user2" {
  depends_on = [panos_template.example]
  location = var.location
  name     = "user2"
  password = "Password123!"
}

resource "panos_local_user" "user3" {
  depends_on = [panos_template.example]
  location = var.location
  name     = "user3"
  password = "Password123!"
}

resource "panos_local_user_group" "example" {
  depends_on = [panos_local_user.user1, panos_local_user.user2, panos_local_user.user3]
  location = var.location
  name     = var.prefix
  users    = var.users
}
`
