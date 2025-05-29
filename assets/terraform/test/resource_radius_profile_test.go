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

func TestAccServerRadiusProfile_Basic(t *testing.T) {
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
				Config: panosServerRadiusProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const panosServerRadiusProfile_Basic_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_radius_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
}
`

const panosServerRadiusProfile_Chap_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_radius_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  protocol = {
    chap = {}
  }
  retries = 3
  timeout = 3
  server = [
    {
      name = "server1"
      ip_address = "192.168.1.1"
      secret = "test-secret-1"
      port = 1812
    },
    {
      name = "server2"
      ip_address = "192.168.1.2"
      secret = "test-secret-2"
      port = 1813
    }
  ]
}
`

func TestAccServerRadiusProfile_Chap(t *testing.T) {
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
				Config: panosServerRadiusProfile_Chap_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("protocol").AtMapKey("chap"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{}),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("server"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name":       knownvalue.StringExact("server1"),
								"ip_address": knownvalue.StringExact("192.168.1.1"),
								"secret":     knownvalue.StringExact("test-secret-1"),
								"port":       knownvalue.Int64Exact(1812),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name":       knownvalue.StringExact("server2"),
								"ip_address": knownvalue.StringExact("192.168.1.2"),
								"secret":     knownvalue.StringExact("test-secret-2"),
								"port":       knownvalue.Int64Exact(1813),
							}),
						}),
					),
				},
			},
		},
	})
}

const panosServerRadiusProfile_EAP_TTLS_with_PAP_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_radius_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  protocol = {
    eap_ttls_with_pap = {
      anon_outer_id = true
      #radius_cert_profile = "test-cert-profile"
    }
  }
  retries = 4
  timeout = 5
  server = [
    {
      name = "server1"
      ip_address = "192.168.1.1"
      secret = "test-secret-1"
      port = 1812
    }
  ]
}
`

func TestAccServerRadiusProfile_EAP_TTLS_with_PAP(t *testing.T) {
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
				Config: panosServerRadiusProfile_EAP_TTLS_with_PAP_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("protocol").AtMapKey("eap_ttls_with_pap"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"anon_outer_id": knownvalue.Bool(true),
							//"radius_cert_profile": knownvalue.StringExact("test-cert-profile"),
							"radius_cert_profile": knownvalue.Null(),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("retries"),
						knownvalue.Int64Exact(4),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("server"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name":       knownvalue.StringExact("server1"),
								"ip_address": knownvalue.StringExact("192.168.1.1"),
								"secret":     knownvalue.StringExact("test-secret-1"),
								"port":       knownvalue.Int64Exact(1812),
							}),
						}),
					),
				},
			},
		},
	})
}

const panosServerRadiusProfile_PAP_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_radius_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  protocol = {
    pap = {}
  }
}
`

func TestAccServerRadiusProfile_PAP(t *testing.T) {
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
				Config: panosServerRadiusProfile_PAP_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("protocol").AtMapKey("pap"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{}),
					),
				},
			},
		},
	})
}

const panosServerRadiusProfile_PEAP_MSCHAPv2_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_radius_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  protocol = {
    peap_mschapv2 = {
      allow_pwd_change = true
      anon_outer_id = true
      #radius_cert_profile = "test-cert-profile"
    }
  }
}
`

func TestAccServerRadiusProfile_PEAP_MSCHAPv2(t *testing.T) {
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
				Config: panosServerRadiusProfile_PEAP_MSCHAPv2_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("protocol").AtMapKey("peap_mschapv2"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"allow_pwd_change": knownvalue.Bool(true),
							"anon_outer_id":    knownvalue.Bool(true),
							//"radius_cert_profile": knownvalue.StringExact("test-cert-profile"),
							"radius_cert_profile": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}

const panosServerRadiusProfile_PEAP_with_GTC_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_radius_profile" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  protocol = {
    peap_with_gtc = {
      anon_outer_id = true
      #radius_cert_profile = "test-cert-profile"
    }
  }
}
`

func TestAccServerRadiusProfile_PEAP_with_GTC(t *testing.T) {
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
				Config: panosServerRadiusProfile_PEAP_with_GTC_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_radius_profile.example",
						tfjsonpath.New("protocol").AtMapKey("peap_with_gtc"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"anon_outer_id": knownvalue.Bool(true),
							//"radius_cert_profile": knownvalue.StringExact("test-cert-profile"),
							"radius_cert_profile": knownvalue.Null(),
						}),
					),
				},
			},
		},
	})
}
