
resource "panos_administrative_tag" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }

  }

  name  = "foo"
  color = "color1"
}
