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
    "bar" = {
      description = "foo"
      ip_netmask  = "2.2.2.2"
    }
  }
}

# Example 1: Get the value of a single address object.
output "foo_value" {
  value = provider::panos::address_value(panos_address_objects.name.addresses.foo)
}

# Example 2: Transform all the address objects into a map of values.
output "address_values" {
  value = { for k, v in panos_address_objects.name.addresses : k => provider::panos::address_value(panos_address_objects.name.addresses[k]) }
}
