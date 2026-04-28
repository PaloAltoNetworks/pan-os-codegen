# Create a template
resource "panos_template" "ospf_simple_template" {
  location = { panorama = {} }
  name     = "ospf-simple-template"
}

# OSPF Authentication Profile using simple password
resource "panos_ospf_auth_routing_profile" "simple_password" {
  location = {
    template = {
      name = panos_template.ospf_simple_template.name
    }
  }

  name     = "ospf-simple-auth"
  password = "ospf-pass"
}
