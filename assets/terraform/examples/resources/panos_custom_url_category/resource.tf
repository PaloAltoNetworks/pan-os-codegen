resource "panos_custom_url_category" "name" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name = "foo"
  type = "URL List"
  list = [
    "test.com",
    "hello.com"
  ]

}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}