resource "panos_interface_management_profile" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "example"
  
  http = true
  ping = true

  permitted_ips = [
    { name = "1.1.1.1" },
    { name = "2.2.2.2" }
  ]

}

resource "panos_template" "example" {

  location = {
    panorama = {}
  }
  name        = "template-example"
  description = "example template"

}