# Mange a group of security policy rules.
resource "panos_security_policy_rules" "name" {
  location = {
    device_group = {
      name = panos_device_group.parent.name
    }
  }


  position = {
    where = "first"
  }

  rules = [
    {
      name                  = "rule-1",
      source_zones          = ["any"],
      source_addresses      = ["1.1.1.1"],
      destination_zones     = ["any"],
      destination_addresses = ["172.16.0.0/8"],
      services              = ["any"],
      applications          = ["any"],
    }
  ]
}
