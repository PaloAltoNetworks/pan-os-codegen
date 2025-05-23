package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCertificateImport_Local_PEM_Certificate(t *testing.T) {
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
				Config: certificateImport_Local_PEM_Certificate_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"location":    location,
					"certificate": config.StringVariable(certPemInitial),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("local").AtMapKey("pem").AtMapKey("certificate"),
						knownvalue.StringExact(certPemInitial),
					),
				},
			},
			{
				Config: certificateImport_Local_PEM_Certificate_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"location":    location,
					"certificate": config.StringVariable(certPemUpdated),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			{
				Config: certificateImport_Local_PEM_Certificate_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"location":    location,
					"certificate": config.StringVariable(certPemUpdated),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("local").AtMapKey("pem").AtMapKey("certificate"),
						knownvalue.StringExact(certPemUpdated),
					),
				},
			},
		},
	})
}

func TestAccCertificateImport_Local_PEM_CertificateWithKey(t *testing.T) {
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
				Config: certificateImport_Local_PEM_CertificateWithKey_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"location":    location,
					"certificate": config.StringVariable(certPemInitial),
					"private_key": config.StringVariable(privateKeyPemInitial),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("local").AtMapKey("pem").AtMapKey("certificate"),
						knownvalue.StringExact(certPemInitial),
					),
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("local").AtMapKey("pem").AtMapKey("private_key"),
						knownvalue.StringExact(privateKeyPemInitial),
					),
				},
			},
			{
				Config: certificateImport_Local_PEM_Certificate_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"location":    location,
					"certificate": config.StringVariable(certPemUpdated),
					"private_key": config.StringVariable(privateKeyPemUpdated),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			{
				Config: certificateImport_Local_PEM_CertificateWithKey_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"location":    location,
					"certificate": config.StringVariable(certPemUpdated),
					"private_key": config.StringVariable(privateKeyPemUpdated),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("local").AtMapKey("pem").AtMapKey("certificate"),
						knownvalue.StringExact(certPemUpdated),
					),
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("local").AtMapKey("pem").AtMapKey("private_key"),
						knownvalue.StringExact(privateKeyPemUpdated),
					),
				},
			},
		},
	})
}

// func TestAccCertificateImport_Local_PKCS12_Certificate(t *testing.T) {
// 	t.Parallel()

// 	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
// 	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

// 	location := config.ObjectVariable(map[string]config.Variable{
// 		"template": config.ObjectVariable(map[string]config.Variable{
// 			"name": config.StringVariable(prefix),
// 		}),
// 	})

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: certificateImport_Local_PKCS12_Certificate_Tmpl,
// 				ConfigVariables: map[string]config.Variable{
// 					"prefix":   config.StringVariable(prefix),
// 					"location": location,
// 				},
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					statecheck.ExpectKnownValue(
// 						"panos_certificate_import.example",
// 						tfjsonpath.New("name"),
// 						knownvalue.StringExact(prefix),
// 					),
// 				},
// 			},
// 		},
// 	})
// }

// func TestAccCertificateImport_Local_PKCS12_CertificateWithKey(t *testing.T) {
// 	t.Parallel()

// 	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
// 	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

// 	// location := config.ObjectVariable(map[string]config.Variable{
// 	// 	"template": config.ObjectVariable(map[string]config.Variable{
// 	// 		"name": config.StringVariable(prefix),
// 	// 	}),
// 	// })
// 	location := config.ObjectVariable(map[string]config.Variable{
// 		"panorama": config.ObjectVariable(map[string]config.Variable{}),
// 	})

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: certificateImport_Local_PKCS12_CertificateWithKey_Tmpl,
// 				ConfigVariables: map[string]config.Variable{
// 					"prefix":   config.StringVariable(prefix),
// 					"location": location,
// 				},
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					statecheck.ExpectKnownValue(
// 						"panos_certificate_import.example",
// 						tfjsonpath.New("name"),
// 						knownvalue.StringExact(prefix),
// 					),
// 				},
// 			},
// 		},
// 	})
// }

const certificateImport_Local_PEM_Certificate_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "certificate" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_certificate_import" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  local = {
    pem = {
      certificate = var.certificate
    }
  }
}
`

const certificateImport_Local_PEM_CertificateWithKey_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }
variable "certificate" { type = string }
variable "private_key" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_certificate_import" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  local = {
    pem = {
      certificate = var.certificate
      private_key = var.private_key
    }
  }
}
`

const certificateImport_Local_PKCS12_Certificate_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_certificate_import" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  local = {
    pkcs12 = {}
  }
}
`

const certificateImport_Local_PKCS12_CertificateWithKey_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = var.prefix
}

resource "panos_certificate_import" "example" {
  depends_on = [panos_template.example]
  location = var.location

  name = var.prefix

  local = {
    pkcs12 = {}
  }
}
`

const certPemInitial = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`

const certPemUpdated = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`

const privateKeyPemInitial = `-----BEGIN PRIVATE KEY-----
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
7xFraUnOHD26eQ5+DBups9jL+Y+Hgw==
-----END PRIVATE KEY-----`

const privateKeyPemUpdated = `-----BEGIN PRIVATE KEY-----
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
oTGtUd5Icd7xkTm7dOJbUXfXuOy8Bps=
-----END PRIVATE KEY-----`

const certKeyPkcs12Initial = `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`

const certKeyPkcs12Updated = `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`
