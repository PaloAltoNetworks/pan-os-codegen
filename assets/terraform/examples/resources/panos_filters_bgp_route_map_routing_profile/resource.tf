# Create a template for the BGP route map routing profile
resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "bgp-route-map-template"
}

# Create supporting resources for match conditions

# IPv4 Access List for matching
resource "panos_filters_access_list_routing_profile" "ipv4_acl" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "ipv4-acl-example"
  description = "IPv4 access list for route map matching"

  type = {
    ipv4 = {
      ipv4_entries = [
        {
          name   = "10"
          action = "permit"
          source_address = {
            address = "any"
          }
        }
      ]
    }
  }
}

# IPv4 Prefix List for matching
resource "panos_filters_prefix_list_routing_profile" "ipv4_prefix" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "ipv4-prefix-example"
  description = "IPv4 prefix list for route map matching"

  type = {
    ipv4 = {
      ipv4_entries = [
        {
          name   = "10"
          action = "permit"
          prefix = {
            entry = {
              network = "10.0.0.0/8"
            }
          }
        }
      ]
    }
  }
}

# Basic BGP Route Map - Simple Deny
resource "panos_filters_bgp_route_map_routing_profile" "basic_deny" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "basic-deny-route-map"
  description = "Simple route map that denies all routes"

  route_map = [
    {
      name        = "10"
      action      = "deny"
      description = "Deny all routes"
    }
  ]
}

# Advanced BGP Route Map - Match and Set Operations
resource "panos_filters_bgp_route_map_routing_profile" "advanced" {
  depends_on = [
    panos_filters_access_list_routing_profile.ipv4_acl,
    panos_filters_prefix_list_routing_profile.ipv4_prefix
  ]

  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "advanced-route-map"
  description = "Advanced route map with match and set operations"

  route_map = [
    {
      name        = "10"
      action      = "permit"
      description = "Match IPv4 address and set communities"

      match = {
        ipv4 = {
          address = {
            access_list = panos_filters_access_list_routing_profile.ipv4_acl.name
          }
        }
        origin           = "igp"
        metric           = 100
        local_preference = 200
      }

      set = {
        local_preference = 300
        weight           = 500
        metric = {
          value  = 50
          action = "add"
        }
        regular_community = ["65001:100", "65001:200"]
        large_community   = ["65001:100:200", "65001:300:400"]
        aspath_prepend    = [65001, 65001]
      }
    },
    {
      name        = "20"
      action      = "permit"
      description = "Match prefix list and modify AS path"

      match = {
        ipv4 = {
          address = {
            prefix_list = panos_filters_prefix_list_routing_profile.ipv4_prefix.name
          }
        }
      }

      set = {
        origin = "incomplete"
        metric = {
          value  = 100
          action = "set"
        }
        aspath_exclude = [65002]
      }
    },
    {
      name        = "30"
      action      = "deny"
      description = "Deny remaining routes"
    }
  ]
}

# Route Map with IPv4 Next-Hop Matching
resource "panos_filters_bgp_route_map_routing_profile" "nexthop_match" {
  depends_on = [panos_filters_access_list_routing_profile.ipv4_acl]

  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "nexthop-match-route-map"
  description = "Route map matching on next-hop"

  route_map = [
    {
      name        = "10"
      action      = "permit"
      description = "Match next-hop and set attributes"

      match = {
        ipv4 = {
          next_hop = {
            access_list = panos_filters_access_list_routing_profile.ipv4_acl.name
          }
        }
      }

      set = {
        local_preference = 150
        weight           = 200
      }
    }
  ]
}

# Route Map with Aggregator and Origin Settings
resource "panos_filters_bgp_route_map_routing_profile" "aggregator" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "aggregator-route-map"
  description = "Route map with aggregator and origin settings"

  route_map = [
    {
      name        = "10"
      action      = "permit"
      description = "Set aggregator and origin"

      set = {
        aggregator = {
          as        = 65001
          router_id = "192.0.2.1"
        }
        origin           = "egp"
        atomic_aggregate = true
        originator_id    = "192.0.2.2"
      }
    }
  ]
}

# Route Map with Boolean Flags
resource "panos_filters_bgp_route_map_routing_profile" "flags" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "flags-route-map"
  description = "Route map demonstrating boolean flags"

  route_map = [
    {
      name        = "10"
      action      = "permit"
      description = "Set various boolean flags"

      set = {
        atomic_aggregate            = true
        ipv6_nexthop_prefer_global  = true
        overwrite_regular_community = true
        overwrite_large_community   = true
      }
    }
  ]
}

# Multiple Match Conditions
resource "panos_filters_bgp_route_map_routing_profile" "multiple_match" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "multiple-match-route-map"
  description = "Route map with multiple match conditions"

  route_map = [
    {
      name        = "10"
      action      = "permit"
      description = "Match on multiple conditions"

      match = {
        origin           = "igp"
        metric           = 100
        tag              = 200
        local_preference = 150
        interface        = "ethernet1/1"
        peer             = "192.0.2.10"
      }

      set = {
        tag              = 300
        local_preference = 250
        weight           = 500
      }
    }
  ]
}
