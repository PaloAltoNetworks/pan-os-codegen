# Basic authentication policy rule in device group
resource "panos_authentication_policy" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  rules = [
    {
      name                     = "require-auth-web-access"
      source_zones             = ["trust"]
      source_addresses         = ["any"]
      destination_zones        = ["untrust"]
      destination_addresses    = ["any"]
      services                 = ["service-http", "service-https"]
      authentication_enforcement = "auth-profile-captive-portal"
      timeout                  = 120
      log_authentication_timeout = true
    }
  ]
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
