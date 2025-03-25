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
    }
  }

}

resource "panos_service_group" "example" {

  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name = "example-service-group"
  # description = "example service group"

  members = [
    panos_service.example.name
  ]
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}