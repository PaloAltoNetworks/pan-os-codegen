# This example generates a throwaway self-signed certificate and key with the
# hashicorp/tls provider purely so the example applies out of the box. DO NOT use
# generated demo material in production - supply your own certificate and private
# key (for example from a file or a secrets manager) instead.
#
# The tls provider (hashicorp/tls) is resolved automatically from the registry;
# add it to your own terraform.required_providers block in a real configuration.

resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "example" {
  private_key_pem = tls_private_key.example.private_key_pem

  subject {
    common_name  = "example.org"
    organization = "Example, Inc."
  }

  validity_period_hours = 87600 # 10 years
  is_ca_certificate     = true

  allowed_uses = [
    "cert_signing",
    "crl_signing",
    "digital_signature",
  ]
}

resource "panos_certificate_import" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "EXAMPLE-CERT"

  local = {
    pem = {
      certificate = tls_self_signed_cert.example.cert_pem   # PEM-encoded certificate
      private_key = tls_private_key.example.private_key_pem # PEM-encoded private key
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
