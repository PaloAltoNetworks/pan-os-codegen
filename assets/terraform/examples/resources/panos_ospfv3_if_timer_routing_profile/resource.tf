# Create a template
resource "panos_template" "ospfv3_template" {
  location = { panorama = {} }
  name     = "ospfv3-routing-template"
}

# OSPFv3 Interface Timer Profile with custom timer values
resource "panos_ospfv3_if_timer_routing_profile" "custom_timers" {
  location = {
    template = {
      name = panos_template.ospfv3_template.name
    }
  }

  name                = "custom-if-timer-profile"
  hello_interval      = 30
  dead_counts         = 4
  retransmit_interval = 10
  transit_delay       = 2
  gr_delay            = 5
}
