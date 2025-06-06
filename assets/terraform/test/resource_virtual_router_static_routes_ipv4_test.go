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

func TestAccVirtualRouterStaticRoutesIpv4(t *testing.T) {
	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"template": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: virtualRouterStaticRoutesIpv4Tmpl1,
				ConfigVariables: map[string]config.Variable{
					"location": location,
					"prefix":   config.StringVariable(prefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("virtual_router"),
						knownvalue.StringExact(fmt.Sprintf("%s-vr1", prefix)),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example2",
						tfjsonpath.New("virtual_router"),
						knownvalue.StringExact(fmt.Sprintf("%s-vr2", prefix)),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example2",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("admin_dist"),
						knownvalue.Int64Exact(15),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("destination"),
						knownvalue.StringExact("192.168.2.0/24"),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("interface"),
						knownvalue.StringExact("ethernet1/1"),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("metric"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("nexthop"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ip_address": knownvalue.StringExact("192.168.1.254"),
							"discard":    knownvalue.Null(),
							"fqdn":       knownvalue.Null(),
							"next_vr":    knownvalue.Null(),
							"receive":    knownvalue.Null(),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("path_monitor"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"enable":            knownvalue.Bool(true),
							"failure_condition": knownvalue.StringExact("any"),
							"hold_time":         knownvalue.Int64Exact(2),
							"monitor_destinations": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"name":        knownvalue.StringExact("dest-1"),
									"enable":      knownvalue.Bool(true),
									"source":      knownvalue.StringExact("192.168.1.1/32"),
									"destination": knownvalue.StringExact("192.168.1.254"),
									"interval":    knownvalue.Int64Exact(3),
									"count":       knownvalue.Int64Exact(5),
								}),
							}),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("route_table"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"unicast":    knownvalue.MapExact(map[string]knownvalue.Check{}),
							"both":       knownvalue.Null(),
							"multicast":  knownvalue.Null(),
							"no_install": knownvalue.Null(),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example2",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("nexthop"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ip_address": knownvalue.Null(),
							"discard":    knownvalue.Null(),
							"fqdn":       knownvalue.Null(),
							"next_vr":    knownvalue.StringExact(fmt.Sprintf("%s-vr1", prefix)),
							"receive":    knownvalue.Null(),
						}),
					),
				},
			},
			{
				Config: virtualRouterStaticRoutesIpv4Tmpl2,
				ConfigVariables: map[string]config.Variable{
					"location": location,
					"prefix":   config.StringVariable(prefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(1).
							AtMapKey("name"),
						knownvalue.StringExact("default"),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("admin_dist"),
						knownvalue.Int64Exact(20),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("destination"),
						knownvalue.StringExact("192.168.2.0/24"),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(0).
							AtMapKey("nexthop"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ip_address": knownvalue.StringExact("192.168.1.254"),
							"discard":    knownvalue.Null(),
							"fqdn":       knownvalue.Null(),
							"next_vr":    knownvalue.Null(),
							"receive":    knownvalue.Null(),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(1).
							AtMapKey("destination"),
						knownvalue.StringExact("0.0.0.0/0"),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router_static_routes_ipv4.example",
						tfjsonpath.New("static_routes").
							AtSliceIndex(1).
							AtMapKey("nexthop"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ip_address": knownvalue.Null(),
							"discard":    knownvalue.Null(),
							"fqdn":       knownvalue.Null(),
							"next_vr":    knownvalue.StringExact(fmt.Sprintf("%s-vr2", prefix)),
							"receive":    knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const virtualRouterStaticRoutesIpv4Tmpl1 = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_ethernet_interface" "example" {
  location = var.location

  name = "ethernet1/1"

  layer3 = {
    ips = [{name = "192.168.1.1/32"}]
  }
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]

  location = var.location

  name = format("%s-vr1", var.prefix)

  interfaces = [panos_ethernet_interface.example.name]
}

resource "panos_virtual_router" "example2" {
  depends_on = [panos_template.example]

  location = var.location

  name = format("%s-vr2", var.prefix)
}

resource "panos_virtual_router_static_routes_ipv4" "example" {
  location = var.location

  virtual_router = panos_virtual_router.example.name

  static_routes = [{
    name = var.prefix
    admin_dist = 15
    destination = "192.168.2.0/24"
    interface = panos_ethernet_interface.example.name
    metric = 100

    #bfd = {
    #  profile = "BFD-profile"
    #}

    nexthop = {
      ip_address = "192.168.1.254"
    }

    path_monitor = {
      enable = true
      failure_condition = "any"
      hold_time = 2
      monitor_destinations = [{
        name = "dest-1"
        enable = true
        source = "192.168.1.1/32"
        destination = "192.168.1.254"
        interval = 3
        count = 5
      }]
    }

    route_table = {
      unicast = {}
    }
  }]
}

resource "panos_virtual_router_static_routes_ipv4" "example2" {
  location = var.location

  virtual_router = panos_virtual_router.example2.name

  static_routes = [{
    name = var.prefix
    destination = "192.168.1.0/24"

    nexthop = {
      next_vr = panos_virtual_router.example.name
    }
  }]
}
`

const virtualRouterStaticRoutesIpv4Tmpl2 = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_ethernet_interface" "example" {
  location = var.location

  name = "ethernet1/1"

  layer3 = {
    ips = [{name = "192.168.1.1/32"}]
  }
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]

  location = var.location

  name = format("%s-vr1", var.prefix)

  interfaces = [panos_ethernet_interface.example.name]
}

resource "panos_virtual_router" "example2" {
  depends_on = [panos_template.example]

  location = var.location

  name = format("%s-vr2", var.prefix)
}

resource "panos_virtual_router_static_routes_ipv4" "example" {
  location = var.location

  virtual_router = panos_virtual_router.example.name

  static_routes = [
    {
      name = var.prefix
      admin_dist = 20
      destination = "192.168.2.0/24"
      interface = panos_ethernet_interface.example.name
      metric = 100

      nexthop = {
        ip_address = "192.168.1.254"
      }
    },
    {
      name = "default"
      destination = "0.0.0.0/0"
      nexthop = {
        next_vr = panos_virtual_router.example2.name
      }
    }
  ]
}
`
