# Create a template
resource "panos_template" "ospfv3_template" {
  location = { panorama = {} }
  name     = "ospfv3-routing-template"
}

# OSPFv3 SPF Timer Profile with custom timing values
resource "panos_ospfv3_spf_timer_routing_profile" "custom_timers" {
  location = {
    template = {
      name = panos_template.ospfv3_template.name
    }
  }

  name                  = "custom-spf-timer-profile"
  spf_calculation_delay = 10
  initial_hold_time     = 15
  max_hold_time         = 30
  lsa_interval          = 8
}
