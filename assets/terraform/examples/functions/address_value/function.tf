resource "panos_address_objects" "name" {
  location = {
    device_group = {
      name = "<device_group_name>"
    }
  }

  addresses = {
    "foo" = {
      description = "foo"
      ip_netmask  = "1.1.1.1"
    }
  }
}

output "address_value" {
  value = provider::panos::address_value(panos_address_objects.name.addresses.foo)
}
