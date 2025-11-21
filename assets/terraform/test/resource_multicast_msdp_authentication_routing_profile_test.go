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

func TestAccMulticastMsdpAuthenticationRoutingProfile_Basic(t *testing.T) {
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
				Config: multicastMsdpAuthenticationRoutingProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_authentication_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_authentication_routing_profile.example",
						tfjsonpath.New("secret"),
						knownvalue.StringExact("mySecret123!"),
					),
				},
			},
		},
	})
}

const multicastMsdpAuthenticationRoutingProfile_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_authentication_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  secret = "mySecret123!"
}
`

func TestAccMulticastMsdpAuthenticationRoutingProfile_NoSecret(t *testing.T) {
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
				Config: multicastMsdpAuthenticationRoutingProfile_NoSecret_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_authentication_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_authentication_routing_profile.example",
						tfjsonpath.New("secret"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

const multicastMsdpAuthenticationRoutingProfile_NoSecret_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_authentication_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
}
`

func TestAccMulticastMsdpAuthenticationRoutingProfile_MaxLengthSecret(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"template": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	// Generate a 63-character secret (max allowed length)
	maxLengthSecret := acctest.RandStringFromCharSet(63, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#%^")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: multicastMsdpAuthenticationRoutingProfile_MaxLengthSecret_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":            config.StringVariable(prefix),
					"location":          location,
					"max_length_secret": config.StringVariable(maxLengthSecret),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_authentication_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_authentication_routing_profile.example",
						tfjsonpath.New("secret"),
						knownvalue.StringExact(maxLengthSecret),
					),
				},
			},
		},
	})
}

const multicastMsdpAuthenticationRoutingProfile_MaxLengthSecret_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "max_length_secret" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_authentication_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  secret = var.max_length_secret
}
`
