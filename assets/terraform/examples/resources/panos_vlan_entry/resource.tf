# This example demonstrates adding static MAC address entries to a VLAN.
# Each panos_vlan_entry binds a specific MAC address to an L2 interface,
# allowing the firewall to forward frames for that MAC without flooding.
#
# The example builds the full dependency chain:
#   panos_template -> panos_vlan_interface -> panos_vlan -> panos_vlan_entry
#
# To use on a standalone NGFW, replace every template location block with:
#   location = { ngfw = {} }

resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "branch-office-template"
}

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

resource "panos_vlan" "production" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "vlan-10-production"

  interfaces = [
    "ethernet1/3",
    "ethernet1/4",
  ]

  virtual_interface = {
    interface = panos_vlan_interface.vlan10.name
  }
}

# Static MAC entry for a server whose frames always arrive on ethernet1/3.
# Using panos_vlan.production.name as the parent reference ensures Terraform
# creates the VLAN before these entries and replaces entries if the VLAN
# is renamed.
resource "panos_vlan_entry" "server_web" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  vlan = panos_vlan.production.name

  # The entry name is the static MAC address in xx:xx:xx:xx:xx:xx format.
  name = "00:1a:2b:3c:4d:5e"

  # The L2 member interface where frames from this MAC are expected.
  # Must be one of the interfaces listed in the parent panos_vlan resource.
  interface = "ethernet1/3"
}

# Static MAC entry for a second host connected via ethernet1/4.
resource "panos_vlan_entry" "server_db" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  vlan = panos_vlan.production.name

  name = "00:1a:2b:3c:4d:6f"

  interface = "ethernet1/4"
}
