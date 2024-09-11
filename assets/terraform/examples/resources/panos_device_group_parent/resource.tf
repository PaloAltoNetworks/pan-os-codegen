resource "panos_device_group" "parent" {
  location = {
    panorama = {}
  }

  name = "parent device group"
}

resource "panos_device_group" "child" {
  location = {
    panorama = {}
  }

  name = "child device group"
}

resource "panos_device_group_parent" "example" {
  location = {
    panorama = {}
  }
  device_group = panos_device_group.child.name
  parent       = panos_device_group.parent.name
}
w