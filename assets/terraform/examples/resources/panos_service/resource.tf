resource "panos_service" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name        = "example-service"
  description = "example service"

  protocol = {
    tcp = {
      destination_port = "80"
      override = {
        timeout           = 600
        halfclose_timeout = 300
        timewait_timeout  = 60
      }
    }
  }
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
