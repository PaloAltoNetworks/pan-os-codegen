# Mange a group of security policy rules.

## Place the rule group at the top
resource "panos_security_policy_rules" "example-1" {
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
      name                  = "rule-1",
      source_zones          = ["any"],
      source_addresses      = ["1.1.1.1"],
      destination_zones     = ["any"],
      destination_addresses = ["172.0.0.0/8"],
      services              = ["any"],
      applications          = ["any"],
    }
  ]
}

## Place the rule group directly after rule-2
resource "panos_security_policy_rules" "example-2" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  position = { where = "after", directly = true, pivot = "rule-2" }

  rules = [ for k in [10, 11, 12]: 
    {
      name                  = "rule-${k}",
      source_zones          = ["any"],
      source_addresses      = ["1.1.1.1"],
      destination_zones     = ["any"],
      destination_addresses = ["172.0.0.0/8"],
      services              = ["any"],
      applications          = ["any"],
    }
  ]
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}