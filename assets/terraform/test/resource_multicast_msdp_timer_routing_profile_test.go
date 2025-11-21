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

func TestAccMulticastMsdpTimerRoutingProfile_Basic(t *testing.T) {
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
				Config: multicastMsdpTimerRoutingProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("connection_retry_interval"),
						knownvalue.Int64Exact(15),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("keep_alive_interval"),
						knownvalue.Int64Exact(30),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("message_timeout"),
						knownvalue.Int64Exact(45),
					),
				},
			},
		},
	})
}

const multicastMsdpTimerRoutingProfile_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_timer_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  connection_retry_interval = 15
  keep_alive_interval = 30
  message_timeout = 45
}
`

func TestAccMulticastMsdpTimerRoutingProfile_MinValues(t *testing.T) {
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
				Config: multicastMsdpTimerRoutingProfile_MinValues_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("connection_retry_interval"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("keep_alive_interval"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("message_timeout"),
						knownvalue.Int64Exact(1),
					),
				},
			},
		},
	})
}

const multicastMsdpTimerRoutingProfile_MinValues_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_timer_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  connection_retry_interval = 1
  keep_alive_interval = 1
  message_timeout = 1
}
`

func TestAccMulticastMsdpTimerRoutingProfile_MaxValues(t *testing.T) {
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
				Config: multicastMsdpTimerRoutingProfile_MaxValues_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("connection_retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("keep_alive_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("message_timeout"),
						knownvalue.Int64Exact(75),
					),
				},
			},
		},
	})
}

const multicastMsdpTimerRoutingProfile_MaxValues_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_timer_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  connection_retry_interval = 60
  keep_alive_interval = 60
  message_timeout = 75
}
`

func TestAccMulticastMsdpTimerRoutingProfile_Defaults(t *testing.T) {
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
				Config: multicastMsdpTimerRoutingProfile_Defaults_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("connection_retry_interval"),
						knownvalue.Int64Exact(30),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("keep_alive_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"panos_multicast_msdp_timer_routing_profile.example",
						tfjsonpath.New("message_timeout"),
						knownvalue.Int64Exact(75),
					),
				},
			},
		},
	})
}

const multicastMsdpTimerRoutingProfile_Defaults_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_multicast_msdp_timer_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
}
`
