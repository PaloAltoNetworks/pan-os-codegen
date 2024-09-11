resource "panos_address_group" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name        = "example-address-group"
  description = "example address group"
  static      = [for k, v in panos_addresses.example.addresses : k]
}

resource "panos_addresses" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  addresses = {
    "foo" = {
      description = "foo example"
      ip_netmask  = "1.1.1.1"
    }
    "bar" = {
      description = "bar example"
      ip_netmask  = "2.2.2.2"
    }
  }
}


