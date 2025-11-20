# Create a template
resource "panos_template" "bgp_template" {
  location = { panorama = {} }
  name     = "bgp-routing-template"
}

# BGP Dampening Profile with custom values
resource "panos_bgp_dampening_routing_profile" "custom" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name               = "custom-dampening-profile"
  description        = "BGP dampening profile with custom timer values"
  half_life          = 10
  max_suppress_limit = 120
  reuse_limit        = 500
  suppress_limit     = 1500
}
