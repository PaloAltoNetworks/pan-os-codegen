resource "panos_virtual_router" "example" {
  name = "vr-1"
  location = {
    template = { name = panos_template.tmpl.name }
  }
  interfaces = [panos_ethernet_interface.eth1.name, panos_ethernet_interface.eth2.name]
  administrative_distances = {
    ebgp = 20
    ibgp = 200
  }
}

resource "panos_ethernet_interface" "eth1" {
  name = "ethernet1/1"
  location = {
    template = { name = panos_template.tmpl.name }
  }
  layer3 = {
    ips = [{ name = "10.1.1.1/24" }]
  }
}

resource "panos_ethernet_interface" "eth2" {
  name = "ethernet1/2"
  location = {
    template = { name = panos_template.tmpl.name }
  }
  layer3 = {
    ips = [{ name = "10.1.2.1/24" }]
  }
}

resource "panos_template" "tpl" {
  name = "my-template"
  location = {
    panorama = true
  }
}
