resource "panos_template_stack" "example" {
  location = {
    panorama = {}
  }

  name        = "example-template-stack"
  description = "example template stack"

  templates = [
    panos_template.example.name
  ]

  # devices = [
  #   { name = panos_firewall_device.fw1.name },
  #   { name = panos_firewall_device.fw2.name }
  # ]

  # default_vsys = "vsys1"
}

resource "panos_template" "example" {
  location = {
    panorama = {}
  }

  name        = "example-template"
  description = "example template"
}

# resource "panos_firewall_device" "fw1" {
#   location = {
#     panorama = {}
#   }
#
#   name     = "007200001234"
#   hostname = "fw1.example.com"
#   ip       = "192.0.2.1"
# }
#
# resource "panos_firewall_device" "fw2" {
#   location = {
#     panorama = {}
#   }
#
#   name     = "007200005678"
#   hostname = "fw2.example.com"
#   ip       = "192.0.2.2"
# }
