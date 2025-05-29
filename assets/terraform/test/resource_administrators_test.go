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

func TestAccAdministrator_Password_Hashing(t *testing.T) {
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
				Config: panosAdministrators_Password_Hashing_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
					"password": config.StringVariable("initial"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("password"),
						knownvalue.StringExact("initial"),
					),
				},
			},
			{
				Config: panosAdministrators_Password_Hashing_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
					"password": config.StringVariable("updated"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("password"),
						knownvalue.StringExact("updated"),
					),
				},
			},
		},
	})
}

const panosAdministrators_Password_Hashing_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }
variable "password" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}

resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = { panorama = {} }

  name = var.prefix

  password = var.password
}
`

func TestAccAdministrator_Basic(t *testing.T) {
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
				Config: panosAdministrators_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					// statecheck.ExpectKnownValue(
					// 	"panos_administrator.example",
					// 	tfjsonpath.New("authentication_profile"),
					// 	knownvalue.StringExact("auth_profile"),
					// ),
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("client_certificate_only"),
						knownvalue.Bool(false),
					),
					// statecheck.ExpectKnownValue(
					// 	"panos_administrator.example",
					// 	tfjsonpath.New("password_profile"),
					// 	knownvalue.StringExact("password_profile"),
					// ),
					// statecheck.ExpectKnownValue(
					// 	"panos_administrator.example",
					// 	tfjsonpath.New("public_key"),
					// 	knownvalue.StringExact("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."),
					// ),
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("preferences").AtMapKey("disable_dns"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("preferences").AtMapKey("saved_log_query").AtMapKey("traffic").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":  knownvalue.StringExact("Example Query"),
							"query": knownvalue.StringExact("addr.src in 10.0.0.0/8"),
						}),
					),
				},
			},
		},
	})
}

const panosAdministrators_Basic_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix
  password = "admin123"

  #authentication_profile = "auth_profile"
  client_certificate_only = false
  #password_profile = "password_profile"
  #public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."

  preferences = {
    disable_dns = true
    saved_log_query = {
      traffic = [
        {
          name = "Example Query"
          query = "addr.src in 10.0.0.0/8"
        }
      ]
    }
  }
}
`

func TestAccAdministrator_RoleBased_Custom(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_Custom_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("custom"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"profile": knownvalue.StringExact("custom_profile"),
							"vsys": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("vsys1"),
								knownvalue.StringExact("vsys2"),
							}),
						}),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_Custom_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      custom = {
        profile = "custom_profile"
        vsys    = ["vsys1", "vsys2"]
      }
    }
  }
}
`

func TestAccAdministrator_RoleBased_DeviceAdmin(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_DeviceAdmin_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("deviceadmin"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("device1"),
							knownvalue.StringExact("device2"),
						}),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_DeviceAdmin_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      deviceadmin = ["device1", "device2"]
    }
  }
}
`

func TestAccAdministrator_RoleBased_DeviceReader(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_DeviceReader_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("devicereader"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("device1"),
							knownvalue.StringExact("device2"),
						}),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_DeviceReader_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      devicereader = ["device1", "device2"]
    }
  }
}
`

func TestAccAdministrator_RoleBased_PanoramaAdmin(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_PanoramaAdmin_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("panorama_admin"),
						knownvalue.StringExact("yes"),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_PanoramaAdmin_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      panorama_admin = "yes"
    }
  }
}
`

func TestAccAdministrator_RoleBased_SuperReader(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_SuperReader_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("superreader"),
						knownvalue.StringExact("yes"),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_SuperReader_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      superreader = "yes"
    }
  }
}
`

func TestAccAdministrator_RoleBased_SuperUser(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_SuperUser_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("superuser"),
						knownvalue.StringExact("yes"),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_SuperUser_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" example {
  location = { panorama = {} }
  name =  var.prefix
}

resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      superuser = "yes"
    }
  }
}
`

func TestAccAdministrator_RoleBased_VsysAdmin(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"panorama": config.ObjectVariable(map[string]config.Variable{}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: panosAdministrators_RoleBased_VsysAdmin_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("vsysadmin"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name": knownvalue.StringExact("vsys_admin1"),
							"vsys": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("vsys1"),
								knownvalue.StringExact("vsys2"),
							}),
						}),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_VsysAdmin_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }
  name =  var.prefix
}

resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      vsysadmin = [
        {
          name = "vsys_admin1"
          vsys = ["vsys1", "vsys2"]
        }
      ]
    }
  }
}
`

func TestAccAdministrator_RoleBased_VsysReader(t *testing.T) {
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
				Config: panosAdministrators_RoleBased_VsysReader_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("permissions").AtMapKey("role_based").AtMapKey("vsysreader").AtSliceIndex(0),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name": knownvalue.StringExact("vsys_reader1"),
							"vsys": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("vsys1"),
								knownvalue.StringExact("vsys2"),
							}),
						}),
					),
				},
			},
		},
	})
}

const panosAdministrators_RoleBased_VsysReader_Tmpl = `
variable "location" { type = any }
variable "prefix" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name =  var.prefix
}


resource "panos_administrator" "example" {
  depends_on = [panos_template.example]
  location = var.location
  name     = var.prefix
  password = "admin123"

  permissions = {
    role_based = {
      vsysreader = [
        {
          name = "vsys_reader1"
          vsys = ["vsys1", "vsys2"]
        }
      ]
    }
  }
}
`
