resource "panos_ike_gateway" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "example-ike-gateway"

  comment  = "ike gateway comment"
  disabled = false
  ipv6     = true

  authentication = {
    certificate = {
      allow_id_payload_mismatch    = true
      strict_validation_revocation = true
      use_management_as_source     = true
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
