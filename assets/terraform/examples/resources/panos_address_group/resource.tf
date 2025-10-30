resource "panos_address_group" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name        = "example-address-group"
  description = "example address group"
  static      = [for k in panos_address.example : k.name]
}

resource "panos_address" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  for_each = tomap({
    "addr1" = {
      description = "example address 1"
      ip_netmask  = "10.0.0.1/32"
    }
    "addr2" = {
      description = "example address 2"
      fqdn        = "example.com"
    }
  })

  name        = each.key
  description = each.value.description
  ip_netmask  = lookup(each.value, "ip_netmask", null)
  fqdn        = lookup(each.value, "fqdn", null)
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
