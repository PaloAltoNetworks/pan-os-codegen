resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

resource "panos_ethernet_interface" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name = "ethernet1/1"
  layer3 = {
    ips = [{ name = "192.168.1.1/32" }]
  }
}

resource "panos_virtual_router" "example" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name       = "example-vr1"
  interfaces = [panos_ethernet_interface.example.name]
}

resource "panos_virtual_router" "example2" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name = "example-vr2"
}

resource "panos_virtual_router_static_routes_ipv4" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  virtual_router = panos_virtual_router.example.name
  static_routes = [{
    name        = "example-route"
    admin_dist  = 15
    destination = "192.168.2.0/24"
    interface   = panos_ethernet_interface.example.name
    metric      = 100
    nexthop = {
      ip_address = "192.168.1.254"
    }
    path_monitor = {
      enable            = true
      failure_condition = "any"
      hold_time         = 2
      monitor_destinations = [{
        name        = "dest-1"
        enable      = true
        source      = "192.168.1.1/32"
        destination = "192.168.1.254"
        interval    = 3
        count       = 5
      }]
    }
    route_table = {
      unicast = {}
    }
  }]
}

resource "panos_virtual_router_static_routes_ipv4" "example2" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  virtual_router = panos_virtual_router.example2.name
  static_routes = [{
    name        = "example-route"
    destination = "192.168.1.0/24"
    nexthop = {
      next_vr = panos_virtual_router.example.name
    }
  }]
}
