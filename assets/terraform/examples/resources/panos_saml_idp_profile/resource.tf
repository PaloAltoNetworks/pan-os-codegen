// Create a template to hold the SAML IdP profile
resource "panos_template" "example" {
  location = { panorama = {} }

  name = "saml-idp-profile-example-tmpl"
}

// Import the IdP signing certificate
resource "panos_certificate_import" "idp_cert" {
  location = { template = { name = panos_template.example.name } }

  name = "saml-idp-example-cert"

  local = {
    pem = {
      certificate = file("idp-cert.pem")
    }
  }
}

// Create the SAML IdP profile
resource "panos_saml_idp_profile" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "saml-idp-profile-example"

  certificate = panos_certificate_import.idp_cert.name
  entity_id   = "my-panos-entity-id"
  sso_url     = "https://idp.example.com/sso"
  slo_url     = "https://idp.example.com/slo"

  max_clock_skew = 90

  // Optional: Set boolean flags
  admin_use_only            = false
  validate_idp_certificate  = true
  want_auth_requests_signed = false

  // Optional: Set attribute import names
  attribute_name_username_import  = "uid"
  attribute_name_usergroup_import = "memberOf"
}
