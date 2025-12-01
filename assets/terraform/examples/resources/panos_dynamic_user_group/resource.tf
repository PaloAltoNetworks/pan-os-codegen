resource "panos_dynamic_user_group" "example" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name        = "developers"
  description = "Dynamic user group for developers"
  filter      = "'department-dev' and 'location-hq'"
  tags = [
    panos_administrative_tag.dev.name,
    panos_administrative_tag.hq.name
  ]
}

resource "panos_administrative_tag" "dev" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name  = "department-dev"
  color = "color3"
}

resource "panos_administrative_tag" "hq" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name  = "location-hq"
  color = "color5"
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
