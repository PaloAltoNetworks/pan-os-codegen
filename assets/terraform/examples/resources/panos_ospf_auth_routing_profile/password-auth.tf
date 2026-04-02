# Create a template for the OSPF routing configuration
resource "panos_template" "ospf_template" {
  location = { panorama = {} }
  name     = "ospf-routing-template"
}

# OSPF Authentication Profile with simple password authentication
resource "panos_ospf_auth_routing_profile" "password_auth" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name     = "ospf-simple-password"
  password = "Palo@123"
}
