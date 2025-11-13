resource "panos_dos_protection_profile" "example" {
  location = {
    device_group = {
      name = "my_device_group"
    }
  }
  name = "example-profile"
  description = "test description"
  disable_override = "no"
  type = "aggregate"
  resource = {
    sessions = {
      enabled = true
      max_concurrent_limit = 1234
    }
  }
  flood = {
    icmp = {
      enable = true
      red = {
        activate_rate = 123
        alarm_rate = 1234
        block = {
          duration = 12345
        }
        maximal_rate = 123456
      }
    }
    tcp_syn = {
      enable = true
      red = {
        activate_rate = 123
        alarm_rate = 1234
        block = {
          duration = 12345
        }
        maximal_rate = 123456
      }
    }
  }
}