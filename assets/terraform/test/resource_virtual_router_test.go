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

func TestAccVirtualRouter_Basic(t *testing.T) {
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
				Config: virtualRouter_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":     config.StringVariable(prefix),
					"location":   location,
					"interfaces": config.ListVariable(config.StringVariable("ethernet1/1")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("interfaces"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("ethernet1/1"),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("administrative_distances"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ebgp":        knownvalue.Int64Exact(21),
							"ibgp":        knownvalue.Int64Exact(201),
							"ospf_ext":    knownvalue.Int64Exact(111),
							"ospf_int":    knownvalue.Int64Exact(31),
							"ospfv3_ext":  knownvalue.Int64Exact(112),
							"ospfv3_int":  knownvalue.Int64Exact(32),
							"rip":         knownvalue.Int64Exact(121),
							"static":      knownvalue.Int64Exact(11),
							"static_ipv6": knownvalue.Int64Exact(12),
						}),
					),
				},
			},
			{
				Config: virtualRouter_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
					"interfaces": config.ListVariable(
						config.StringVariable("ethernet1/1"),
						config.StringVariable("ethernet1/2"),
					),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("interfaces"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("ethernet1/1"),
							knownvalue.StringExact("ethernet1/2"),
						}),
					),
				},
			},
		},
	})
}

func TestAccVirtualRouter_Ecmp(t *testing.T) {
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
				Config: virtualRouter_Ecmp_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("ecmp"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"enable":             knownvalue.Bool(true),
							"max_paths":          knownvalue.Int64Exact(3),
							"strict_source_path": knownvalue.Bool(true),
							"symmetric_return":   knownvalue.Bool(true),
							"algorithm":          knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

func TestAccVirtualRouter_Ecmp_BalancedRoundRobin(t *testing.T) {
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
				Config: virtualRouter_Ecmp_BalancedRoundRobin_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("ecmp").AtMapKey("algorithm"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"balanced_round_robin": knownvalue.ObjectExact(map[string]knownvalue.Check{}),
							"ip_hash":              knownvalue.Null(),
							"ip_modulo":            knownvalue.Null(),
							"weighted_round_robin": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

func TestAccVirtualRouter_Ecmp_IpHash(t *testing.T) {
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
				Config: virtualRouter_Ecmp_IpHash_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("ecmp").AtMapKey("algorithm"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ip_hash": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"hash_seed": knownvalue.Int64Exact(123),
								"src_only":  knownvalue.Bool(true),
								"use_port":  knownvalue.Bool(true),
							}),
							"balanced_round_robin": knownvalue.Null(),
							"ip_modulo":            knownvalue.Null(),
							"weighted_round_robin": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

func TestAccVirtualRouter_Ecmp_IpModulo(t *testing.T) {
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
				Config: virtualRouter_Ecmp_IpModulo_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("ecmp").AtMapKey("algorithm"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ip_modulo":            knownvalue.ObjectExact(map[string]knownvalue.Check{}),
							"ip_hash":              knownvalue.Null(),
							"balanced_round_robin": knownvalue.Null(),
							"weighted_round_robin": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

func TestAccVirtualRouter_Ecmp_WeightedRoundRobin(t *testing.T) {
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
				Config: virtualRouter_Ecmp_WeightedRoundRobin_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("ecmp").AtMapKey("algorithm"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"weighted_round_robin": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"interface": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":   knownvalue.StringExact("ethernet1/1"),
										"weight": knownvalue.Int64Exact(150),
									}),
								}),
							}),
							"ip_hash":              knownvalue.Null(),
							"balanced_round_robin": knownvalue.Null(),
							"ip_modulo":            knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

func TestAccVirtualRouter_Multicast(t *testing.T) {
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
				Config: virtualRouter_Multicast_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_virtual_router.example",
						tfjsonpath.New("multicast"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"enable":            knownvalue.Bool(true),
							"interface_group":   knownvalue.Null(),
							"route_ageout_time": knownvalue.Int64Exact(210),
							"rp":                knownvalue.Null(),
							"spt_threshold":     knownvalue.Null(),
							"ssm_address_space": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const virtualRouter_Multicast_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name = var.prefix
  multicast = {
    enable = true
  }
}
`

const virtualRouter_Ecmp_WeightedRoundRobin_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_ethernet_interface" "iface1" {
  location = { template = { name = panos_template.example.name } }
  name = "ethernet1/1"
  layer3 = {
    ips = [{ name = "192.168.1.1/24" }]
  }
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example, panos_ethernet_interface.iface1]
  location = var.location
  name = var.prefix
  interfaces = [panos_ethernet_interface.iface1.name]
  ecmp = {
    enable = true
    algorithm = {
      weighted_round_robin = {
        interface = [
          {
            name = panos_ethernet_interface.iface1.name
            weight = 150
          }
        ]
      }
    }
  }
}
`

const virtualRouter_Ecmp_IpModulo_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name = var.prefix
  ecmp = {
    enable = true
    algorithm = {
      ip_modulo = {}
    }
  }
}
`

const virtualRouter_Ecmp_IpHash_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name = var.prefix
  ecmp = {
    enable = true
    algorithm = {
      ip_hash = {
        hash_seed = 123
        src_only = true
        use_port = true
      }
    }
  }
}
`

const virtualRouter_Ecmp_BalancedRoundRobin_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name = var.prefix
  ecmp = {
    enable = true
    algorithm = {
      balanced_round_robin = {}
    }
  }
}
`

const virtualRouter_Ecmp_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name = var.prefix
  ecmp = {
    enable = true
    max_paths = 3
    strict_source_path = true
    symmetric_return = true
  }
}
`

const virtualRouter_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "interfaces" { type = list(string) }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_ethernet_interface" "iface1" {
  location = { template = { name = panos_template.example.name } }
  name = "ethernet1/1"
  layer3 = { 
    ips = [{ name = "192.168.1.1/24" }]
  }
}

resource "panos_ethernet_interface" "iface2" {
  location = { template = { name = panos_template.example.name } }
  name = "ethernet1/2"
  layer3 = { 
    ips = [{ name = "192.168.2.1/24" }]
  }
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_ethernet_interface.iface1, panos_ethernet_interface.iface2]
  location = var.location
  name = var.prefix
  interfaces = var.interfaces
  administrative_distances = {
    ebgp = 21
    ibgp = 201
    ospf_ext = 111
    ospf_int = 31
    ospfv3_ext = 112
    ospfv3_int = 32
    rip = 121
    static = 11
    static_ipv6 = 12
  }
}
`
