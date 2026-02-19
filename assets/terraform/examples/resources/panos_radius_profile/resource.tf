resource "panos_radius_profile" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "radius-basic"
  retries = 3
  timeout = 5
  servers = [
    {
      name       = "radius1"
      ip_address = "10.0.1.10"
      secret     = "secret123"
      port       = 1812
    },
    {
      name       = "radius2"
      ip_address = "10.0.1.11"
      secret     = "secret456"
      port       = 1812
    }
  ]
}

resource "panos_template" "example" {
  location = {
    panorama = {}
  }

  name = "example-template"
}
