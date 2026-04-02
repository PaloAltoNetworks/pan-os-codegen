# Manage a group of Policy Based Forwarding rules with positioning

## Place the rule group at the top of the pre-rulebase
resource "panos_pbf_policy_rules" "priority_routing" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "pre-rulebase"
    }
  }

  position = {
    where = "first"
  }

  rules = [
    {
      name                  = "route-voip-traffic"
      description           = "Route VoIP traffic through low-latency path"
      source_addresses      = ["corporate-network"]
      destination_addresses = ["voip-servers"]
      services              = ["sip", "rtp"]
      applications          = ["sip", "voip"]

      from = {
        zone = ["trust"]
      }

      action = {
        forward = {
          egress_interface = "ethernet1/2"
          nexthop = {
            ip_address = "10.10.0.1"
          }
          monitor = {
            ip_address             = "10.10.0.1"
            profile                = "high-availability"
            disable_if_unreachable = true
          }
        }
      }

      enforce_symmetric_return = {
        enabled = true
      }

      tags = ["voip", "priority"]
    }
  ]
}

## Place the rule group after a specific rule with forward action using FQDN
resource "panos_pbf_policy_rules" "application_routing" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "pre-rulebase"
    }
  }

  position = {
    where    = "after"
    directly = true
    pivot    = "route-voip-traffic"
  }

  rules = [
    {
      name                  = "route-backup-traffic"
      description           = "Route backup traffic through dedicated backup link"
      source_addresses      = ["backup-servers"]
      destination_addresses = ["backup-storage"]
      services              = ["any"]
      applications          = ["backup"]

      from = {
        zone = ["dmz"]
      }

      action = {
        forward = {
          egress_interface = "ethernet1/4"
          nexthop = {
            fqdn = "backup-gateway.example.com"
          }
          monitor = {
            ip_address             = "10.30.0.1"
            disable_if_unreachable = false
          }
        }
      }

      schedule = "backup-window"
      tags     = ["backup", "scheduled"]
    },
    {
      name                  = "block-suspicious-traffic"
      description           = "Discard traffic from untrusted sources"
      source_addresses      = ["suspicious-network"]
      destination_addresses = ["any"]
      services              = ["any"]
      applications          = ["any"]

      from = {
        zone = ["untrust"]
      }

      action = {
        discard = {}
      }

      disabled = false
      tags     = ["security", "block"]
    }
  ]
}

## Advanced rule with interface-based source and specific target devices
resource "panos_pbf_policy_rules" "interface_routing" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "post-rulebase"
    }
  }

  position = {
    where = "last"
  }

  rules = [
    {
      name                  = "route-iot-devices"
      description           = "Route IoT device traffic through isolated network segment"
      source_addresses      = ["iot-network"]
      destination_addresses = ["cloud-services"]
      negate_destination    = false
      services              = ["service-https"]
      applications          = ["any"]
      source_users          = ["any"]

      from = {
        interface = ["ethernet1/5", "ethernet1/6"]
      }

      action = {
        forward = {
          egress_interface = "ethernet1/7"
          nexthop = {
            ip_address = "10.40.0.1"
          }
        }
      }

      active_active_device_binding = "both"

      # Target specific devices in the device group
      target = {
        devices = [
          {
            name = "fw-branch-01"
            vsys = [
              { name = "vsys1" }
            ]
          }
        ]
        negate = false
        tags   = ["branch-office"]
      }

      tags = ["iot", "isolated"]
    },
    {
      name                  = "route-to-virtual-system"
      description           = "Route traffic to a different virtual system for processing"
      source_addresses      = ["cross-vsys-network"]
      destination_addresses = ["shared-resources"]
      services              = ["any"]
      applications          = ["any"]

      from = {
        zone = ["trust"]
      }

      action = {
        forward_to_vsys = "vsys2"
      }

      tags = ["cross-vsys"]
    }
  ]
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
