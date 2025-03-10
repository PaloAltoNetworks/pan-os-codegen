# Manages the entire NAT policy
resource "panos_nat_policy" "name" {
  location = {
    device_group = {
      name     = "example device group"
      rulebase = "post-rulebase" // Options: pre-rulebase, post-rulebase
    }
  }


  rules = [{
    name = "rule-1"

    source_zones          = ["Trust"]         # from
    source_addresses      = ["10.0.0.0/24"]   # source
    destination_zone      = ["Untrust"]       # to
    destination_addresses = ["172.16.0.0/16"] # destination
    services              = ["any"]

    source_translation = {
      static_ip = {
        translated_address = "192.168.0.1"
      }
    }
    }
  ]
}
