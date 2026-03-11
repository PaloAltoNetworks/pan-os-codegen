# Manage a group of QoS policy rules.

## Place the rule group at the top
resource "panos_qos_policy_rules" "example-1" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  position = {
    where = "first" # first, last, after, before
  }

  rules = [
    {
      name        = "qos-rule-1"
      description = "High priority VoIP traffic"

      source_zones          = ["trust"]
      source_addresses      = ["any"]
      destination_zones     = ["untrust"]
      destination_addresses = ["any"]
      applications          = ["sip", "h323"]
      services              = ["application-default"]

      action = {
        class = "7"
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
    }
  ]
}

## Place the rule group directly after an existing rule
resource "panos_qos_policy_rules" "example-2" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  position = { where = "after", directly = true, pivot = "existing-rule" }

  rules = [for k in ["web", "database", "default"] :
    {
      name = "qos-${k}"

      source_zones          = ["any"]
      source_addresses      = ["any"]
      destination_zones     = ["any"]
      destination_addresses = ["any"]
      applications          = ["any"]
      services              = ["any"]

      action = {
        class = k == "web" ? "5" : k == "database" ? "4" : "1"
      }

      dscp_tos = {
        codepoints = [
          {
            name = "${k}-codepoint"
            af = {
              codepoint = k == "web" ? "af21" : k == "database" ? "af31" : "af11"
            }
          }
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
