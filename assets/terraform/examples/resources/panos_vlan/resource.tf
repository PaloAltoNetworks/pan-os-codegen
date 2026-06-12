# This example demonstrates configuring a VLAN in a Panorama template.
# The VLAN binds L2 member interfaces together and, optionally, references
# a VLAN (L3 subinterface) for routed traffic.
#
# To use on a standalone NGFW instead of Panorama, replace the location block
# with:
#   location = { ngfw = {} }
# and remove the panos_template dependency.

resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "branch-office-template"
}

# The VLAN interface provides an L3 gateway for hosts on this VLAN.
# Assumption: virtual_interface is optional. It is included here to show
# how to bind an L3 VLAN interface to the VLAN entry.
resource "panos_vlan_interface" "vlan10" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name    = "vlan.10"
  comment = "L3 gateway for production VLAN 10"
  ip = [{
    name = "10.10.0.1/24"
  }]
}

# VLAN 10 — production segment.
# The interfaces list contains the L2 (layer2-mode) ethernet interfaces
# that are members of this VLAN.
# Assumption: interfaces is optional (zero or more members are allowed).
resource "panos_vlan" "production" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "vlan-10-production"

  # L2 interfaces that belong to this VLAN.
  interfaces = [
    "ethernet1/3",
    "ethernet1/4",
  ]

  # Bind the L3 VLAN interface so routed traffic exits via vlan.10.
  virtual_interface = {
    interface = panos_vlan_interface.vlan10.name
  }
}
