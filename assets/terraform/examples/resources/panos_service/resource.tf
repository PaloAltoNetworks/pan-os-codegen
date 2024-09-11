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
