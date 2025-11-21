resource "panos_tacacs_plus_profile" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name     = "example-tacacs-profile"
  protocol = "CHAP"

  servers = [
    {
      name    = "tacacs-server-1"
      address = "192.168.1.10"
      secret  = "shared-secret-1"
      port    = 49
    },
    {
      name    = "tacacs-server-2"
      address = "192.168.1.11"
      secret  = "shared-secret-2"
      port    = 49
    }
  ]

  timeout               = 5
  use_single_connection = true
}

resource "panos_template" "example" {
  location = {
    panorama = {}
  }

  name = "example-template"
}
