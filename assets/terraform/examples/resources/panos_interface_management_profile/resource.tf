resource "panos_interface_management_profile" "name" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "foo"
  http = true
  ping = true

  permitted_ips = [
    "1.1.1.1",
  ]

}

resource "panos_template" "example" {

  location = {
    panorama = {}
  }
  name        = "template-example"
  description = "example template stack"

}