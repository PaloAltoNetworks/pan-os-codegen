# Create a template
resource "panos_template" "ospf_template" {
  location = { panorama = {} }
  name     = "ospf-routing-template"
}

# OSPF SPF Timer Profile with custom timing values
resource "panos_ospf_spf_timer_routing_profile" "custom_timers" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name                  = "custom-spf-timer-profile"
  spf_calculation_delay = 10
  initial_hold_time     = 15
  max_hold_time         = 30
  lsa_interval          = 8
}
