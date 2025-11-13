
resource "panos_zone" "zone" {
  location = {
    template = {
      name = panos_template.example.name
      vsys = "vsys1"
    }
  }

  name = "example-zone"

  device_acl = {
    # exclude_list = ["device-1"]
    # include_list = ["device-2"]
  }

  enable_device_identification = true
  enable_user_identification   = true

  network = {
    layer3 = [
      panos_ethernet_interface.iface1.name,
      panos_ethernet_interface.iface2.name,
      panos_tunnel_interface.iface3.name,
    ]
    enable_packet_buffer_protection = true
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}


resource "panos_ethernet_interface" "iface1" {

  location = {
    template = {
      name = panos_template.example.name
      vsys = "vsys1"
    }
  }

  name = "ethernet1/1"

  layer3 = {}
}

resource "panos_ethernet_interface" "iface2" {
  location = {
    template = {
      name = panos_template.example.name
      vsys = "vsys1"
    }
  }

  name = "ethernet1/2"

  layer3 = {}
}

resource "panos_tunnel_interface" "iface3" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "tunnel.1"
}
