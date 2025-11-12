# Create a template
resource "panos_template" "bgp_template" {
  location = { panorama = {} }
  name     = "bgp-routing-template"
}

# BGP Timer Profile with custom timer values
resource "panos_bgp_timer_routing_profile" "custom_timers" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name                             = "custom-timer-profile"
  hold_time                        = "180"
  keep_alive_interval              = "60"
  min_route_advertisement_interval = 15
  open_delay_time                  = 5
  reconnect_retry_interval         = 30
}
