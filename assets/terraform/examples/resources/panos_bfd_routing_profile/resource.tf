# Create a template
resource "panos_template" "bfd_template" {
  location = { panorama = {} }
  name     = "bfd-routing-template"
}

# BFD Profile with custom timing values
resource "panos_bfd_routing_profile" "custom_timers" {
  location = {
    template = {
      name = panos_template.bfd_template.name
    }
  }

  name                 = "custom-bfd-profile"
  detection_multiplier = 5
  hold_time            = 1000
  min_rx_interval      = 500
  min_tx_interval      = 500
  mode                 = "active"
}

# BFD Profile with multihop configuration
resource "panos_bfd_routing_profile" "multihop" {
  location = {
    template = {
      name = panos_template.bfd_template.name
    }
  }

  name                 = "multihop-bfd-profile"
  detection_multiplier = 3
  min_rx_interval      = 300
  min_tx_interval      = 300
  mode                 = "passive"

  multihop = {
    min_received_ttl = 128
  }
}
