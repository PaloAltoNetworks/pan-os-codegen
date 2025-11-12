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

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Basic(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Test redistribution routing profile"),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("bgp"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"rib": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"route_map": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":        knownvalue.StringExact("1"),
										"action":      knownvalue.StringExact("permit"),
										"description": knownvalue.Null(),
										"match":       knownvalue.Null(),
										"set":         knownvalue.Null(),
									}),
								}),
							}),
							"ospf":   knownvalue.Null(),
							"ospfv3": knownvalue.Null(),
							"rip":    knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  description = "Test redistribution routing profile"

  bgp = {
    rib = {
      route_map = [
        {
          name = "1"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Bgp_Ospf(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Bgp_Ospf_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("bgp"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ospf": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"route_map": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":   knownvalue.StringExact("10"),
										"action": knownvalue.StringExact("permit"),
										"description": knownvalue.StringExact("Redistribute BGP to OSPF"),
										"match": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"metric":              knownvalue.Int64Exact(100),
											"tag":                 knownvalue.Int64Exact(200),
											"origin":              knownvalue.StringExact("igp"),
											"local_preference":    knownvalue.Int64Exact(150),
											"as_path_access_list": knownvalue.Null(),
											"regular_communities": knownvalue.Null(),
											"large_communities":   knownvalue.Null(),
											"extended_communities": knownvalue.Null(),
											"interface":           knownvalue.Null(),
											"peer":                knownvalue.Null(),
											"ipv4":                knownvalue.Null(),
										}),
										"set": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"metric": knownvalue.ObjectExact(map[string]knownvalue.Check{
												"value":  knownvalue.Int64Exact(50),
												"action": knownvalue.StringExact("set"),
											}),
											"metric_type": knownvalue.StringExact("type-1"),
											"tag":         knownvalue.Int64Exact(300),
										}),
									}),
								}),
							}),
							"ospfv3": knownvalue.Null(),
							"rib":    knownvalue.Null(),
							"rip":    knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Bgp_Ospf_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  bgp = {
    ospf = {
      route_map = [
        {
          name = "10"
          action = "permit"
          description = "Redistribute BGP to OSPF"
          match = {
            metric = 100
            tag = 200
            origin = "igp"
            local_preference = 150
          }
          set = {
            metric = {
              value = 50
              action = "set"
            }
            metric_type = "type-1"
            tag = 300
          }
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Bgp_Ospfv3(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Bgp_Ospfv3_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("bgp"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ospfv3": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"route_map": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":        knownvalue.StringExact("20"),
										"action":      knownvalue.StringExact("deny"),
										"description": knownvalue.StringExact("Redistribute BGP to OSPFv3"),
										"match": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"metric":              knownvalue.Int64Exact(50),
											"tag":                 knownvalue.Int64Exact(100),
											"as_path_access_list": knownvalue.Null(),
											"regular_communities": knownvalue.Null(),
											"large_communities":   knownvalue.Null(),
											"extended_communities": knownvalue.Null(),
											"interface":           knownvalue.Null(),
											"origin":              knownvalue.Null(),
											"local_preference":    knownvalue.Null(),
											"peer":                knownvalue.Null(),
											"ipv6":                knownvalue.Null(),
										}),
										"set": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"metric": knownvalue.ObjectExact(map[string]knownvalue.Check{
												"value":  knownvalue.Int64Exact(75),
												"action": knownvalue.StringExact("add"),
											}),
											"metric_type": knownvalue.StringExact("type-2"),
											"tag":         knownvalue.Int64Exact(150),
										}),
									}),
								}),
							}),
							"ospf": knownvalue.Null(),
							"rib":  knownvalue.Null(),
							"rip":  knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Bgp_Ospfv3_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  bgp = {
    ospfv3 = {
      route_map = [
        {
          name = "20"
          action = "deny"
          description = "Redistribute BGP to OSPFv3"
          match = {
            metric = 50
            tag = 100
          }
          set = {
            metric = {
              value = 75
              action = "add"
            }
            metric_type = "type-2"
            tag = 150
          }
        }
      ]
    }
  }
}
`
func TestAccFiltersRouteMapsRedistributionRoutingProfile_Bgp_Rib(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Bgp_Rib_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("bgp"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"rib": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"route_map": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":        knownvalue.StringExact("30"),
										"action":      knownvalue.StringExact("permit"),
										"description": knownvalue.Null(),
										"match":       knownvalue.Null(),
										"set":         knownvalue.Null(),
									}),
								}),
							}),
							"ospf":   knownvalue.Null(),
							"ospfv3": knownvalue.Null(),
							"rip":    knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Bgp_Rib_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  bgp = {
    rib = {
      route_map = [
        {
          name = "30"
          action = "permit"
        }
      ]
    }
  }
}
`
func TestAccFiltersRouteMapsRedistributionRoutingProfile_Bgp_Rip(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Bgp_Rip_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("bgp"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"rip": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"route_map": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":   knownvalue.StringExact("40"),
										"action": knownvalue.StringExact("permit"),
										"description": knownvalue.StringExact("Redistribute BGP to RIP"),
										"match": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"metric":              knownvalue.Int64Exact(5),
											"tag":                 knownvalue.Int64Exact(50),
											"as_path_access_list": knownvalue.Null(),
											"regular_communities": knownvalue.Null(),
											"large_communities":   knownvalue.Null(),
											"extended_communities": knownvalue.Null(),
											"interface":           knownvalue.Null(),
											"origin":              knownvalue.Null(),
											"local_preference":    knownvalue.Null(),
											"peer":                knownvalue.Null(),
											"ipv4":                knownvalue.Null(),
										}),
										"set": knownvalue.ObjectExact(map[string]knownvalue.Check{
											"metric": knownvalue.ObjectExact(map[string]knownvalue.Check{
												"value":  knownvalue.Int64Exact(10),
												"action": knownvalue.StringExact("set"),
											}),
											"next_hop": knownvalue.StringExact("10.0.0.1"),
											"tag":      knownvalue.Int64Exact(100),
										}),
									}),
								}),
							}),
							"ospf":   knownvalue.Null(),
							"ospfv3": knownvalue.Null(),
							"rib":    knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Bgp_Rip_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  bgp = {
    rip = {
      route_map = [
        {
          name = "40"
          action = "permit"
          description = "Redistribute BGP to RIP"
          match = {
            metric = 5
            tag = 50
          }
          set = {
            metric = {
              value = 10
              action = "set"
            }
            next_hop = "10.0.0.1"
            tag = 100
          }
        }
      ]
    }
  }
}
`
func TestAccFiltersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Bgp(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Bgp_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("connected_static"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"bgp": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"route_map": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"name":        knownvalue.StringExact("50"),
										"action":      knownvalue.StringExact("permit"),
										"description": knownvalue.Null(),
										"match":       knownvalue.Null(),
										"set":         knownvalue.Null(),
									}),
								}),
							}),
							"ospf":   knownvalue.Null(),
							"ospfv3": knownvalue.Null(),
							"rib":    knownvalue.Null(),
							"rip":    knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Bgp_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  connected_static = {
    bgp = {
      route_map = [
        {
          name = "50"
          action = "permit"
        }
      ]
    }
  }
}
`
func TestAccFiltersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Ospf(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Ospf_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Ospf_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  connected_static = {
    ospf = {
      route_map = [
        {
          name = "60"
          action = "deny"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Ospfv3(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Ospfv3_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Ospfv3_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  connected_static = {
    ospfv3 = {
      route_map = [
        {
          name = "70"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Rib(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Rib_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Rib_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  connected_static = {
    rib = {
      route_map = [
        {
          name = "80"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Rip(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Rip_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_ConnectedStatic_Rip_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  connected_static = {
    rip = {
      route_map = [
        {
          name = "90"
          action = "permit"
        }
      ]
    }
  }
}
`
func TestAccFiltersRouteMapsRedistributionRoutingProfile_Ospf_Bgp(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Ospf_Bgp_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Ospf_Bgp_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  ospf = {
    bgp = {
      route_map = [
        {
          name = "100"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Ospf_Rib(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Ospf_Rib_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Ospf_Rib_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  ospf = {
    rib = {
      route_map = [
        {
          name = "110"
          action = "deny"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Ospf_Rip(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Ospf_Rip_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Ospf_Rip_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  ospf = {
    rip = {
      route_map = [
        {
          name = "120"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Ospfv3_Bgp(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Ospfv3_Bgp_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Ospfv3_Bgp_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  ospfv3 = {
    bgp = {
      route_map = [
        {
          name = "130"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Ospfv3_Rib(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Ospfv3_Rib_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Ospfv3_Rib_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  ospfv3 = {
    rib = {
      route_map = [
        {
          name = "140"
          action = "deny"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Rip_Bgp(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Rip_Bgp_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Rip_Bgp_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  rip = {
    bgp = {
      route_map = [
        {
          name = "150"
          action = "permit"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Rip_Ospf(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Rip_Ospf_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Rip_Ospf_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  rip = {
    ospf = {
      route_map = [
        {
          name = "160"
          action = "deny"
        }
      ]
    }
  }
}
`

func TestAccFiltersRouteMapsRedistributionRoutingProfile_Rip_Rib(t *testing.T) {
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
				Config: filtersRouteMapsRedistributionRoutingProfile_Rip_Rib_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_filters_route_maps_redistribution_routing_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const filtersRouteMapsRedistributionRoutingProfile_Rip_Rib_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_filters_route_maps_redistribution_routing_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  rip = {
    rib = {
      route_map = [
        {
          name = "170"
          action = "permit"
        }
      ]
    }
  }
}
`
