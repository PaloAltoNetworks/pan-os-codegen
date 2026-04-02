# Create a template for OSPF redistribution profiles
resource "panos_template" "ospf_template" {
  location = { panorama = {} }
  name     = "ospf-routing-template"
}

# Redistribute connected routes into OSPF with basic configuration
resource "panos_ospf_redistribution_routing_profile" "connected" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name = "ospf-redistribute-connected"

  connected = {
    enable      = true
    metric      = 10
    metric_type = "type-1"
  }
}

# Redistribute BGP routes into OSPF with type-2 metric
resource "panos_ospf_redistribution_routing_profile" "bgp" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name = "ospf-redistribute-bgp"

  bgp = {
    enable      = true
    metric      = 100
    metric_type = "type-2"
  }
}

# Redistribute static routes with route-map filtering
resource "panos_ospf_redistribution_routing_profile" "static_with_map" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name = "ospf-redistribute-static-filtered"

  static = {
    enable    = true
    route_map = "static-route-filter"
    # Note: metric and metric_type are ignored when route_map is configured
  }
}

# Redistribute multiple sources into OSPF with different configurations
resource "panos_ospf_redistribution_routing_profile" "multiple" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name = "ospf-redistribute-multiple"

  connected = {
    enable      = true
    metric      = 10
    metric_type = "type-1"
  }

  static = {
    enable      = true
    metric      = 20
    metric_type = "type-1"
  }

  bgp = {
    enable      = true
    metric      = 100
    metric_type = "type-2"
  }

  rip = {
    enable      = true
    metric      = 50
    metric_type = "type-2"
    route_map   = "rip-filter-map"
  }
}

# Default route redistribution with always option
# The 'always' option generates a default route even if one doesn't exist
resource "panos_ospf_redistribution_routing_profile" "default_route" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name = "ospf-redistribute-default"

  default_route = {
    enable      = true
    always      = true
    metric      = 1
    metric_type = "type-1"
  }
}
