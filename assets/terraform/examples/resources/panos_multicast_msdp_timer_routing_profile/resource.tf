# Create a template
resource "panos_template" "multicast_template" {
  location = { panorama = {} }
  name     = "multicast-routing-template"
}

# MSDP Timer Profile with custom values
resource "panos_multicast_msdp_timer_routing_profile" "custom_timers" {
  location = {
    template = {
      name = panos_template.multicast_template.name
    }
  }

  name                      = "msdp-timer-profile-custom"
  connection_retry_interval = 15
  keep_alive_interval       = 30
  message_timeout           = 45
}

# MSDP Timer Profile with default values
resource "panos_multicast_msdp_timer_routing_profile" "default_timers" {
  location = {
    template = {
      name = panos_template.multicast_template.name
    }
  }

  name = "msdp-timer-profile-default"
  # Uses default values:
  # connection_retry_interval = 30
  # keep_alive_interval = 60
  # message_timeout = 75
}

# MSDP Timer Profile with maximum values
resource "panos_multicast_msdp_timer_routing_profile" "max_timers" {
  location = {
    template = {
      name = panos_template.multicast_template.name
    }
  }

  name                      = "msdp-timer-profile-max"
  connection_retry_interval = 60
  keep_alive_interval       = 60
  message_timeout           = 75
}

# MSDP Timer Profile with minimum values
resource "panos_multicast_msdp_timer_routing_profile" "min_timers" {
  location = {
    template = {
      name = panos_template.multicast_template.name
    }
  }

  name                      = "msdp-timer-profile-min"
  connection_retry_interval = 1
  keep_alive_interval       = 1
  message_timeout           = 1
}
