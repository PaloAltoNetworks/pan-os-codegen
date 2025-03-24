# Manages the entire NAT policy
resource "panos_nat_policy" "example" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "post-rulebase" // Options: pre-rulebase, post-rulebase
    }
  }


  rules = [
    {
      name = "rule-1"

      source_zones          = [panos_zone.trust.name]   # from
      source_addresses      = ["10.0.0.0/24"]           # source
      destination_zone      = [panos_zone.untrust.name] # to
      destination_addresses = ["172.16.0.0/16"]         # destination
      services              = ["any"]

      source_translation = {
        static_ip = {
          translated_address = "192.168.0.1"
        }
      }
    },
    {
      name = "rule-2"

      source_zones          = [panos_zone.trust.name]   # from
      source_addresses      = ["10.0.0.0/24"]           # source
      destination_zone      = [panos_zone.untrust.name] # to
      destination_addresses = ["172.16.0.0/16"]         # destination
      services              = ["any"]

      source_translation = {
        dynamic_ip_and_port = {
          interface_address = {
            interface = panos_ethernet_interface.example.name
            ip        = "10.1.0.0/24"
          }
        }
      }
    }
  ]
}

resource "panos_ethernet_interface" "example" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }

  name = "ethernet1/1"

  layer3 = {}
}

resource "panos_zone" "trust" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "Trust"
}

resource "panos_zone" "untrust" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "Untrust"
}

resource "panos_device_group" "example" {
  location = { panorama = {} }

  name = "example-device-group"
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
