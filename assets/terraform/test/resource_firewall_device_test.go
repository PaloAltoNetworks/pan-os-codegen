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

func TestAccFirewallDevice_Basic(t *testing.T) {
	t.Parallel()

	// Generate a serial number matching PAN-OS format: 00 + 13 digits
	suffix := acctest.RandStringFromCharSet(13, "0123456789")
	serialNumber := fmt.Sprintf("00%s", suffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"panorama": config.ObjectVariable(map[string]config.Variable{}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: firewallDevice_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"serial_number": config.StringVariable(serialNumber),
					"location":      location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(serialNumber),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("auto_push"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("disable_config_backup"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("fw.example.com"),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.0.2.1"),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("to_sw_version"),
						knownvalue.StringExact("11.0.0"),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("vsys"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

const firewallDevice_Basic_Tmpl = `
variable "serial_number" { type = string }
variable "location" { type = any }

resource "panos_firewall_device" "example" {
  location = var.location
  name = var.serial_number

  auto_push = true
  disable_config_backup = false
  hostname = "fw.example.com"
  ip = "192.0.2.1"
  to_sw_version = "11.0.0"
}
`

func TestAccFirewallDevice_Vsys(t *testing.T) {
	t.Parallel()

	// Generate a serial number matching PAN-OS format: 00 + 13 digits
	suffix := acctest.RandStringFromCharSet(13, "0123456789")
	serialNumber := fmt.Sprintf("00%s", suffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"panorama": config.ObjectVariable(map[string]config.Variable{}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: firewallDevice_Vsys_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"serial_number": config.StringVariable(serialNumber),
					"location":      location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(serialNumber),
					),
					statecheck.ExpectKnownValue(
						"panos_firewall_device.example",
						tfjsonpath.New("vsys"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("vsys1"),
								"tags": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("tag1"),
									knownvalue.StringExact("tag2"),
								}),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("vsys2"),
								"tags": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("tag3"),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

const firewallDevice_Vsys_Tmpl = `
variable "serial_number" { type = string }
variable "location" { type = any }

resource "panos_firewall_device" "example" {
  location = var.location
  name = var.serial_number

  vsys = [
    {
      name = "vsys1"
      tags = ["tag1", "tag2"]
    },
    {
      name = "vsys2"
      tags = ["tag3"]
    }
  ]
}
`
