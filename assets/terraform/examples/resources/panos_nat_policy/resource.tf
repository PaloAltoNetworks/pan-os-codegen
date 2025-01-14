# Manages the entire NAT policy
resource "panos_nat_policy" "name" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  rules = {
    name = "rule-1"

    source_zones          = ["any"]           # from
    source_addresses      = ["10.0.0.0/24"]   # source
    destination_zone      = ["any"]           # to
    destination_addresses = ["172.16.0.0/16"] # destination
    services              = ["any"]

    source_translation = {
      static_ip = {
        translated_address = ["192.168.0.1/24"]
      }
    }
  }
}
