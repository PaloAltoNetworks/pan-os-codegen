resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

resource "panos_ethernet_interface" "example" {
  location = panos_template.example.location
  name     = "ethernet1/1"
  layer3 = {
    ipv6 = {
      enabled   = true
      addresses = [{ name = "2001:db8:1::1/64" }]
    }
  }
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location   = panos_template.example.location
  name       = "example-vr1"
  interfaces = [panos_ethernet_interface.example.name]
}

resource "panos_virtual_router" "example2" {
  depends_on = [panos_template.example]
  location   = panos_template.example.location
  name       = "example-vr2"
}

resource "panos_virtual_router_static_route_ipv6" "example" {
  location       = panos_template.example.location
  virtual_router = panos_virtual_router.example.name
  name           = "example-route"
  admin_dist     = 15
  destination    = "2001:db8:2::/64"
  interface      = panos_ethernet_interface.example.name
  metric         = 100

  nexthop = {
    ipv6_address = "2001:db8:1::254"
  }

  path_monitor = {
    enable            = true
    failure_condition = "any"
    hold_time         = 2
    monitor_destinations = [{
      name        = "dest-1"
      enable      = true
      source      = "2001:db8:1::1/64"
      destination = "2001:db8:1::254"
      interval    = 3
      count       = 5
    }]
  }

  route_table = {
    unicast = {}
  }
}

resource "panos_virtual_router_static_route_ipv6" "example2" {
  location       = panos_template.example.location
  virtual_router = panos_virtual_router.example2.name
  name           = "example-route2"
  destination    = "2001:db8:1::/64"

  nexthop = {
    next_vr = panos_virtual_router.example.name
  }
}
