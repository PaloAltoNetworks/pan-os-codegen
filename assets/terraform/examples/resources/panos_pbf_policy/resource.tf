# Manages the entire Policy Based Forwarding policy
resource "panos_pbf_policy" "example" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "pre-rulebase" # Options: pre-rulebase, post-rulebase
    }
  }

  rules = [
    {
      name                  = "route-guest-traffic"
      description           = "Route guest network traffic through dedicated gateway"
      source_addresses      = ["guest-network"]
      destination_addresses = ["any"]
      services              = ["any"]
      applications          = ["web-browsing", "ssl"]

      from = {
        zone = ["guest"]
      }

      action = {
        forward = {
          egress_interface = "ethernet1/3"
          nexthop = {
            ip_address = "10.20.0.1"
          }
          monitor = {
            ip_address             = "10.20.0.1"
            profile                = "default"
            disable_if_unreachable = true
          }
        }
      }

      enforce_symmetric_return = {
        enabled = true
        nexthop_address_list = [
          { name = "10.20.0.1" }
        ]
      }
    }
  ]
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
