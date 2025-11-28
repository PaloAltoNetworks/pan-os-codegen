# Manages the entire QoS policy
resource "panos_qos_policy" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  rules = [
    {
      name        = "qos-rule-1"
      description = "QoS rule for high priority traffic"

      source_zones          = ["trust"]
      source_addresses      = ["any"]
      destination_zones     = ["untrust"]
      destination_addresses = ["any"]
      applications          = ["ssl"]
      services              = ["application-default"]

      action = {
        class = "4"
      }

      dscp_tos = {
        codepoints = [
          {
            name = "ef-marking"
            ef = {
              codepoint = "ef"
            }
          }
        ]
      }
    },
    {
      name = "qos-rule-2"

      source_zones          = ["any"]
      source_addresses      = ["any"]
      destination_zones     = ["any"]
      destination_addresses = ["any"]
      applications          = ["any"]
      services              = ["any"]

      action = {
        class = "1"
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
