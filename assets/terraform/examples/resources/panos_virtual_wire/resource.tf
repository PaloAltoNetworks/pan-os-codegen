resource "panos_template" "tmpl" {
  location = { panorama = {} }
  name     = "my-template"
}

resource "panos_ethernet_interface" "iface1" {
  location     = { template = { name = resource.panos_template.template.name, vsys = "vsys1" } }
  name         = var.interface1
  virtual_wire = {}
}

resource "panos_ethernet_interface" "iface2" {
  location     = { template = { name = resource.panos_template.template.name, vsys = "vsys1" } }
  name         = var.interface2
  virtual_wire = {}
}


resource "panos_virtual_wire" "example" {
  location   = { template = { name = panos_template.tmpl.name } }
  name       = "vw-1"
  interface1 = "ethernet1/1"
  interface2 = "ethernet1/2"
}
