# OSPFv3 Redistribution Profile - Connected Routes
# This example redistributes directly connected routes into OSPFv3
# with Type-1 metric, making OSPFv3 prefer these routes based on
# accumulated cost from the redistributing router
resource "panos_ospfv3_redistribution_routing_profile" "connected_routes" {
  location = {
    template = {
      name = "production-template"
    }
  }

  name = "ospfv3-redistribute-connected"

  connected = {
    enable      = true
    metric      = 10
    metric_type = "type-1"
  }
}

# OSPFv3 Redistribution Profile - Default Route with Always Option
# This example advertises a default route into OSPFv3 even if the
# router doesn't have one in its routing table. The 'always' flag
# is critical for ensuring default route availability
resource "panos_ospfv3_redistribution_routing_profile" "default_route" {
  location = {
    template = {
      name = "production-template"
    }
  }

  name = "ospfv3-default-originate"

  default_route = {
    enable      = true
    always      = true
    metric      = 1
    metric_type = "type-1"
  }
}

# OSPFv3 Redistribution Profile - Multiple Sources with Route Map
# This example redistributes multiple route sources (connected, static, BGP)
# into OSPFv3. BGP routes use a route-map for selective redistribution and
# metric manipulation, while connected/static use direct metrics
resource "panos_ospfv3_redistribution_routing_profile" "multi_source" {
  location = {
    template = {
      name = "production-template"
    }
  }

  name = "ospfv3-redistribute-all"

  # Redistribute connected interfaces with low metric
  connected = {
    enable      = true
    metric      = 10
    metric_type = "type-1"
  }

  # Redistribute static routes with medium metric
  static = {
    enable      = true
    metric      = 50
    metric_type = "type-2"
  }

  # Redistribute BGP routes using route-map for filtering
  # Note: When route-map is configured, metric and metric_type are ignored
  bgp = {
    enable    = true
    route_map = "bgp-to-ospfv3-filter"
  }
}
