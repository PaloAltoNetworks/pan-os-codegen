resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}


resource "panos_general_settings" "example" {
  location = { template = { name = panos_template.example.name } }

  hostname = "device"
  domain   = "example.com"
  geo_location = {
    latitude  = "40.7128"
    longitude = "-74.0060"
  }
  login_banner = "Example Login Banner"
}
