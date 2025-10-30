
resource "panos_administrative_tag" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }

  }

  name  = "foo"
  color = "color1"
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
