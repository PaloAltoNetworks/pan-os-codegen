# Create a template
resource "panos_template" "multicast_template" {
  location = { panorama = {} }
  name     = "multicast-routing-template"
}

# MSDP Authentication Profile with secret
resource "panos_multicast_msdp_authentication_routing_profile" "with_secret" {
  location = {
    template = {
      name = panos_template.multicast_template.name
    }
  }

  name   = "msdp-auth-profile"
  secret = "mySecretKey123!"
}

# MSDP Authentication Profile without secret
resource "panos_multicast_msdp_authentication_routing_profile" "without_secret" {
  location = {
    template = {
      name = panos_template.multicast_template.name
    }
  }

  name = "msdp-auth-profile-no-secret"
}
