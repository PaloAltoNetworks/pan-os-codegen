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

func TestAccNetflowServerProfile_Basic(t *testing.T) {
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
				Config: netflowServerProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("active_timeout"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("export_enterprise_fields"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("servers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("server1"),
								"host": knownvalue.StringExact("192.168.1.1"),
								"port": knownvalue.Int64Exact(2055),
							}),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("template_refresh_rate"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"minutes": knownvalue.Int64Exact(20),
							"packets": knownvalue.Int64Exact(30),
						}),
					),
				},
			},
		},
	})
}

const netflowServerProfile_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_netflow_server_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  active_timeout = 10
  export_enterprise_fields = true
  servers = [
    {
      name = "server1"
      host = "192.168.1.1"
      port = 2055
    }
  ]
  template_refresh_rate = {
    minutes = 20
    packets = 30
  }
}
`

func TestAccNetflowServerProfile_MultipleServers(t *testing.T) {
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
				Config: netflowServerProfile_MultipleServers_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("servers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("server1"),
								"host": knownvalue.StringExact("192.168.1.10"),
								"port": knownvalue.Int64Exact(2055),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("server2"),
								"host": knownvalue.StringExact("192.168.1.20"),
								"port": knownvalue.Int64Exact(9996),
							}),
						}),
					),
				},
			},
		},
	})
}

const netflowServerProfile_MultipleServers_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_netflow_server_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  servers = [
    {
      name = "server1"
      host = "192.168.1.10"
      port = 2055
    },
    {
      name = "server2"
      host = "192.168.1.20"
      port = 9996
    }
  ]
}
`

func TestAccNetflowServerProfile_MinimalConfig(t *testing.T) {
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
				Config: netflowServerProfile_MinimalConfig_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("active_timeout"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("export_enterprise_fields"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("servers"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"panos_netflow_server_profile.example",
						tfjsonpath.New("template_refresh_rate"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

const netflowServerProfile_MinimalConfig_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }
  name = var.prefix
}

resource "panos_netflow_server_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
}
`