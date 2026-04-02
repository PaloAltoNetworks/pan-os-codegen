# Create a template for the OSPF routing configuration
resource "panos_template" "ospf_password_template" {
  location = { panorama = {} }
  name     = "ospf-password-template"
}

# OSPF Authentication Profile with simple password authentication
resource "panos_ospf_auth_routing_profile" "password_auth" {
  location = {
    template = {
      name = panos_template.ospf_password_template.name
    }
  }

  name     = "ospf-simple-password"
  password = "Palo@123"
}
