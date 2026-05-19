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

func TestAccDecryptionProfile_Basic(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_Basic_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("disable_override"),
						knownvalue.StringExact("yes"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("forwarded_only"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("interface"),
						knownvalue.StringExact("ethernet1/1"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_forward_proxy"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"auto_include_altname": knownvalue.Bool(true),
							"block_client_cert":    knownvalue.Bool(true),
							"block_expired_certificate":         knownvalue.Bool(true),
							"block_if_hsm_unavailable":          knownvalue.Bool(true),
							"block_if_no_resource":              knownvalue.Bool(true),
							// block_if_sni_mismatch has version constraint (11.0.2-11.0.3)
							"block_if_sni_mismatch":             knownvalue.Bool(true),
							"block_timeout_cert":                knownvalue.Bool(true),
							"block_tls13_downgrade_no_resource": knownvalue.Bool(true),
							"block_unknown_cert":                knownvalue.Bool(true),
							"block_unsupported_cipher":          knownvalue.Bool(true),
							"block_unsupported_version":         knownvalue.Bool(true),
							"block_untrusted_issuer":            knownvalue.Bool(true),
							"restrict_cert_exts":                knownvalue.Bool(true),
							"strip_alpn":                        knownvalue.Bool(true),
						}),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("min_version"),
						knownvalue.StringExact("tls1-1"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("max_version"),
						knownvalue.StringExact("tls1-3"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("auth_algo_sha256"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_aes_256_gcm"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("keyxchg_algo_ecdhe"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

const decryptionProfile_Basic_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	disable_override = "yes"
	forwarded_only = true
	interface = "ethernet1/1"
	ssl_forward_proxy = {
		auto_include_altname             = true
		block_client_cert                = true
		block_expired_certificate        = true
		block_if_hsm_unavailable         = true
		block_if_no_resource             = true
		# block_if_sni_mismatch has version constraint (11.0.2-11.0.3)
		block_if_sni_mismatch            = true
		block_timeout_cert               = true
		block_tls13_downgrade_no_resource = true
		block_unknown_cert               = true
		block_unsupported_cipher         = true
		block_unsupported_version        = true
		block_untrusted_issuer           = true
		restrict_cert_exts               = true
		strip_alpn                       = true
	}
	ssl_protocol_settings = {
		min_version      = "tls1-1"
		max_version      = "tls1-3"
		auth_algo_sha256 = true
		enc_algo_aes_256_gcm = true
		keyxchg_algo_ecdhe   = true
	}
}
`

func TestAccDecryptionProfile_SshProxy(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_SshProxy_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssh_proxy"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"block_if_no_resource":      knownvalue.Bool(true),
							"block_ssh_errors":          knownvalue.Bool(true),
							"block_unsupported_alg":     knownvalue.Bool(true),
							"block_unsupported_version": knownvalue.Bool(true),
						}),
					),
				},
			},
		},
	})
}

const decryptionProfile_SshProxy_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	ssh_proxy = {
		block_if_no_resource      = true
		block_ssh_errors          = true
		block_unsupported_alg     = true
		block_unsupported_version = true
	}
}
`

func TestAccDecryptionProfile_SslInboundProxy(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_SslInboundProxy_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_inbound_proxy"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"block_if_hsm_unavailable":          knownvalue.Bool(true),
							"block_if_no_resource":              knownvalue.Bool(true),
							"block_tls13_downgrade_no_resource": knownvalue.Bool(true),
							"block_unsupported_cipher":          knownvalue.Bool(true),
							"block_unsupported_version":         knownvalue.Bool(true),
						}),
					),
				},
			},
		},
	})
}

const decryptionProfile_SslInboundProxy_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	ssl_inbound_proxy = {
		block_if_hsm_unavailable          = true
		block_if_no_resource              = true
		block_tls13_downgrade_no_resource = true
		block_unsupported_cipher          = true
		block_unsupported_version         = true
	}
}
`

func TestAccDecryptionProfile_SslNoProxy(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_SslNoProxy_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_no_proxy"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"block_expired_certificate": knownvalue.Bool(true),
							"block_untrusted_issuer":    knownvalue.Bool(true),
						}),
					),
				},
			},
		},
	})
}

const decryptionProfile_SslNoProxy_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	ssl_no_proxy = {
		block_expired_certificate = true
		block_untrusted_issuer    = true
	}
}
`

// --- Enum value coverage tests for ssl_protocol_settings min/max versions ---

func TestAccDecryptionProfile_SslProtocolSettings_Versions(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_SslProtocolSettings_Versions_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("min_version"),
						knownvalue.StringExact("tls1-0"),
					),
					// "max" means no upper bound restriction
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("max_version"),
						knownvalue.StringExact("max"),
					),
				},
			},
		},
	})
}

const decryptionProfile_SslProtocolSettings_Versions_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	ssl_protocol_settings = {
		min_version = "tls1-0"
		max_version = "max"
	}
}
`

func TestAccDecryptionProfile_SslProtocolSettings_Complete(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_SslProtocolSettings_Complete_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("min_version"),
						knownvalue.StringExact("tls1-2"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("max_version"),
						knownvalue.StringExact("tls1-3"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("auth_algo_md5"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("auth_algo_sha1"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("auth_algo_sha256"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("auth_algo_sha384"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_3des"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_aes_128_cbc"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_aes_128_gcm"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_aes_256_cbc"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_aes_256_gcm"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_chacha20_poly1305"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("enc_algo_rc4"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("keyxchg_algo_dhe"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("keyxchg_algo_ecdhe"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("keyxchg_algo_rsa"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

const decryptionProfile_SslProtocolSettings_Complete_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	ssl_protocol_settings = {
		min_version              = "tls1-2"
		max_version              = "tls1-3"
		auth_algo_md5            = true
		auth_algo_sha1           = true
		auth_algo_sha256         = true
		auth_algo_sha384         = true
		enc_algo_3des            = true
		enc_algo_aes_128_cbc     = true
		enc_algo_aes_128_gcm     = true
		enc_algo_aes_256_cbc     = true
		enc_algo_aes_256_gcm     = true
		enc_algo_chacha20_poly1305 = true
		enc_algo_rc4             = true
		keyxchg_algo_dhe         = true
		keyxchg_algo_ecdhe       = true
		keyxchg_algo_rsa         = true
	}
}
`

func TestAccDecryptionProfile_SslProtocolSettings_MinVersionTls13(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"device_group": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(prefix),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: decryptionProfile_SslProtocolSettings_MinVersionTls13_Tmpl,
				ConfigVariables: map[string]config.Variable{
					"prefix":   config.StringVariable(prefix),
					"location": location,
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("min_version"),
						knownvalue.StringExact("tls1-3"),
					),
					statecheck.ExpectKnownValue(
						"panos_decryption_profile.example",
						tfjsonpath.New("ssl_protocol_settings").AtMapKey("max_version"),
						knownvalue.StringExact("tls1-3"),
					),
				},
			},
		},
	})
}

const decryptionProfile_SslProtocolSettings_MinVersionTls13_Tmpl = `
variable "prefix" { type = string }
variable "location" { type = any }

resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = var.prefix
}

resource "panos_decryption_profile" "example" {
	depends_on = [panos_device_group.example]
	location = var.location
	name = var.prefix
	ssl_protocol_settings = {
		min_version = "tls1-3"
		max_version = "tls1-3"
	}
}
`
