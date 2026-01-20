# Example 1: Basic virtual router with interfaces and administrative distances
resource "panos_template" "basic" {
  location = {
    panorama = {}
  }
  name = "basic-vr-template"
}

resource "panos_ethernet_interface" "eth1" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.basic.name
    }
  }
  name = "ethernet1/1"
  layer3 = {
    mtu = 1500
    ips = [{ name = "10.1.1.1/24" }]
  }
}

resource "panos_ethernet_interface" "eth2" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.basic.name
    }
  }
  name = "ethernet1/2"
  layer3 = {
    mtu = 1500
    ips = [{ name = "10.1.2.1/24" }]
  }
}

resource "panos_virtual_router" "basic" {
  location = {
    template = {
      name = panos_template.basic.name
    }
  }

  name = "production-vr"

  interfaces = [
    panos_ethernet_interface.eth1.name,
    panos_ethernet_interface.eth2.name
  ]

  administrative_distances = {
    static      = 10
    static_ipv6 = 10
    ospf_int    = 30
    ospf_ext    = 110
    ibgp        = 200
    ebgp        = 20
    rip         = 120
  }
}

# Example 2: Virtual router with comprehensive BGP configuration
resource "panos_template" "bgp" {
  location = {
    panorama = {}
  }
  name = "bgp-vr-template"
}

resource "panos_virtual_router" "bgp" {
  location = {
    template = {
      name = panos_template.bgp.name
    }
  }

  name = "bgp-edge-router"

  protocol = {
    bgp = {
      # Core BGP settings
      enable                     = true
      router_id                  = "192.168.100.1"
      local_as                   = "65100"
      install_route              = true
      reject_default_route       = false
      allow_redist_default_route = true
      ecmp_multi_as              = false
      enforce_first_as           = true

      # BGP authentication profile
      auth_profile = [
        {
          name   = "bgp-auth-main"
          secret = "bgp-secure-password-2024"
        }
      ]

      # Peer groups: EBGP for external peers and IBGP for internal mesh
      peer_group = [
        {
          name   = "upstream-providers"
          enable = true
          type = {
            ebgp = {
              export_nexthop = "use-self"
              import_nexthop = "original"
            }
          }
          peer = [
            {
              name     = "isp-primary"
              enable   = true
              local_ip = "192.168.100.1"
              peer_ip  = "192.168.100.254"
              peer_as  = "65000"
            },
            {
              name     = "isp-backup"
              enable   = true
              local_ip = "192.168.100.1"
              peer_ip  = "192.168.100.253"
              peer_as  = "65001"
            }
          ]
        },
        {
          name   = "internal-mesh"
          enable = true
          type = {
            ibgp = {}
          }
          peer = [
            {
              name     = "core-router-1"
              enable   = true
              local_ip = "192.168.100.1"
              peer_ip  = "192.168.101.1"
              peer_as  = "65100"
            }
          ]
        }
      ]

      # BGP routing policies
      policy = {
        # Export rules: control what routes we advertise to peers
        export = {
          rules = [
            {
              name   = "advertise-local-networks"
              enable = true
              match = {
                # Match locally originated routes (empty AS path)
                as_path = {
                  regex = "^$"
                }
              }
              action = {
                allow = {
                  update = {
                    origin           = "igp"
                    med              = 100
                    local_preference = 150
                    # Prepend our AS twice for path manipulation
                    as_path = {
                      prepend = 2
                    }
                    # Tag with community for route tracking
                    community = {
                      append = ["65100:1000", "65100:2000"]
                    }
                  }
                }
              }
            },
            {
              name   = "block-private-as"
              enable = true
              match = {
                # Block routes with private AS numbers
                as_path = {
                  regex = "^6500[0-9]"
                }
              }
              action = {
                deny = {}
              }
            }
          ]
        }

        # Import rules: control what routes we accept from peers
        import = {
          rules = [
            {
              name   = "prefer-customer-routes"
              enable = true
              match = {
                # Match customer routes by AS path and community
                as_path = {
                  regex = ".*65200.*"
                }
                community = {
                  regex = "65200:.*"
                }
              }
              action = {
                allow = {
                  update = {
                    # Increase local preference for customer routes
                    local_preference = 200
                  }
                }
              }
            }
          ]
        }

        # Route aggregation for summarizing address blocks
        aggregation = {
          address = [
            {
              name    = "datacenter-summary"
              prefix  = "10.0.0.0/8"
              enable  = true
              summary = true
              aggregate_route_attributes = {
                origin = "incomplete"
                med    = 50
                as_path = {
                  prepend = 1
                }
                community = {
                  argument = ["65100:3000", "65100:4000"]
                }
              }
            }
          ]
        }
      }

      # Redistribute connected and static routes into BGP
      redist_rules = [
        {
          name   = "redist-connected"
          enable = true
        },
        {
          name   = "redist-static"
          enable = true
        }
      ]
    }
  }
}
