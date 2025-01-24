# Manages the entire security policy
resource "panos_security_policy" "name" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  rules = [
    {
      # rule_type             = "intrazone",
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
