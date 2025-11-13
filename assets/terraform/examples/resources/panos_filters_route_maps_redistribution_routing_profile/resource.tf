# Create a template
resource "panos_template" "redistribution_template" {
  location = { panorama = {} }
  name     = "redistribution-routing-template"
}

# BGP to OSPF Redistribution with Match and Set Attributes
resource "panos_filters_route_maps_redistribution_routing_profile" "bgp_to_ospf" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "bgp-to-ospf-redistribution"
  description = "Redistribute BGP routes into OSPF with filtering and metric manipulation"

  bgp = {
    ospf = {
      route_map = [
        {
          name        = "10"
          action      = "permit"
          description = "Permit BGP routes with specific attributes"
          match = {
            metric           = 100
            tag              = 200
            origin           = "igp"
            local_preference = 150
          }
          set = {
            metric = {
              value  = 50
              action = "set"
            }
            metric_type = "type-1"
            tag         = 300
          }
        },
        {
          name        = "20"
          action      = "deny"
          description = "Deny all other BGP routes"
        }
      ]
    }
  }
}

# OSPF to BGP Redistribution with IPv4 Address Matching
resource "panos_filters_route_maps_redistribution_routing_profile" "ospf_to_bgp" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "ospf-to-bgp-redistribution"
  description = "Redistribute OSPF routes into BGP with route filtering"

  ospf = {
    bgp = {
      route_map = [
        {
          name        = "10"
          action      = "permit"
          description = "Permit OSPF routes and set BGP attributes"
          match = {
            metric    = 10
            tag       = 100
            interface = "ethernet1/1"
          }
          set = {
            metric = {
              value  = 200
              action = "set"
            }
            as_path_prepend  = "65001 65001"
            local_preference = 150
            origin           = "igp"
          }
        }
      ]
    }
  }
}

# Connected/Static to BGP Redistribution
resource "panos_filters_route_maps_redistribution_routing_profile" "connected_to_bgp" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "connected-to-bgp-redistribution"
  description = "Redistribute connected and static routes into BGP"

  connected_static = {
    bgp = {
      route_map = [
        {
          name        = "10"
          action      = "permit"
          description = "Permit connected/static routes"
          match = {
            interface = "ethernet1/2"
            metric    = 0
          }
          set = {
            metric = {
              value  = 100
              action = "set"
            }
            local_preference = 200
            origin           = "incomplete"
          }
        }
      ]
    }
  }
}

# BGP to RIP Redistribution with Metric Manipulation
resource "panos_filters_route_maps_redistribution_routing_profile" "bgp_to_rip" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "bgp-to-rip-redistribution"
  description = "Redistribute BGP routes into RIP with hop count control"

  bgp = {
    rip = {
      route_map = [
        {
          name        = "10"
          action      = "permit"
          description = "Set RIP metric for BGP routes"
          match = {
            metric = 50
            tag    = 100
          }
          set = {
            metric = {
              value  = 5
              action = "set"
            }
            next_hop = "10.0.0.1"
            tag      = 200
          }
        }
      ]
    }
  }
}

# BGP to OSPFv3 Redistribution (IPv6)
resource "panos_filters_route_maps_redistribution_routing_profile" "bgp_to_ospfv3" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "bgp-to-ospfv3-redistribution"
  description = "Redistribute BGP routes into OSPFv3 for IPv6"

  bgp = {
    ospfv3 = {
      route_map = [
        {
          name        = "10"
          action      = "permit"
          description = "Permit BGP routes for OSPFv3"
          match = {
            metric = 100
            tag    = 50
          }
          set = {
            metric = {
              value  = 75
              action = "add"
            }
            metric_type = "type-2"
            tag         = 150
          }
        }
      ]
    }
  }
}

# BGP to RIB Redistribution (Simple)
resource "panos_filters_route_maps_redistribution_routing_profile" "bgp_to_rib" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "bgp-to-rib-redistribution"
  description = "Install BGP routes into routing table"

  bgp = {
    rib = {
      route_map = [
        {
          name   = "10"
          action = "permit"
        }
      ]
    }
  }
}

# Multi-Protocol Redistribution Example
resource "panos_filters_route_maps_redistribution_routing_profile" "rip_to_ospf" {
  location = {
    template = {
      name = panos_template.redistribution_template.name
    }
  }

  name        = "rip-to-ospf-redistribution"
  description = "Redistribute RIP routes into OSPF"

  rip = {
    ospf = {
      route_map = [
        {
          name        = "10"
          action      = "permit"
          description = "Convert RIP metrics to OSPF"
          match = {
            metric    = 3
            tag       = 100
            interface = "ethernet1/3"
          }
          set = {
            metric = {
              value  = 20
              action = "set"
            }
            metric_type = "type-2"
            tag         = 200
          }
        }
      ]
    }
  }
}
