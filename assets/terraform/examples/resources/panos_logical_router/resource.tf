# Example 1: Basic logical router with interfaces and static routes
resource "panos_template" "basic_lr" {
  location = {
    panorama = {}
  }
  name = "basic-lr-template"
}

resource "panos_ethernet_interface" "lr_eth1" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.basic_lr.name
    }
  }
  name = "ethernet1/1"
  layer3 = {
    mtu = 1500
    ips = [{ name = "10.1.1.1/24" }]
  }
}

resource "panos_ethernet_interface" "lr_eth2" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.basic_lr.name
    }
  }
  name = "ethernet1/2"
  layer3 = {
    mtu = 1500
    ips = [{ name = "10.1.2.1/24" }]
  }
}

resource "panos_logical_router" "basic" {
  location = {
    template = {
      name = panos_template.basic_lr.name
    }
  }

  name = "lr-basic"

  vrf = [
    {
      name = "default"
      interface = [
        panos_ethernet_interface.lr_eth1.name,
        panos_ethernet_interface.lr_eth2.name
      ]

      # Configure administrative distances for different routing protocols
      administrative_distances = {
        static       = 10
        static_ipv6  = 10
        ospf_inter   = 110
        ospf_intra   = 110
        ospf_ext     = 150
        bgp_internal = 200
        bgp_external = 20
      }

      # IPv4 static routes
      routing_table = {
        ip = {
          static_route = [
            {
              name        = "default-route"
              destination = "0.0.0.0/0"
              interface   = panos_ethernet_interface.lr_eth1.name
              nexthop = {
                ip_address = "10.1.1.254"
              }
              metric                  = 10
              administrative_distance = 10
            },
            {
              name        = "internal-network"
              destination = "192.168.0.0/16"
              interface   = panos_ethernet_interface.lr_eth2.name
              nexthop = {
                ip_address = "10.1.2.254"
              }
              metric                  = 10
              administrative_distance = 10
            }
          ]
        }
      }
    }
  ]
}

# Example 2: Logical router with comprehensive BGP configuration
resource "panos_template" "bgp_lr" {
  location = {
    panorama = {}
  }
  name = "bgp-lr-template"
}

resource "panos_logical_router" "bgp" {
  location = {
    template = {
      name = panos_template.bgp_lr.name
    }
  }

  name = "lr-bgp-edge"

  vrf = [
    {
      name      = "default"
      interface = ["ethernet1/1", "ethernet1/2"]

      # BGP configuration with peer groups
      bgp = {
        enable                         = true
        router_id                      = "192.0.2.1"
        local_as                       = "65001"
        install_route                  = true
        enforce_first_as               = true
        fast_external_failover         = true
        ecmp_multi_as                  = false
        default_local_preference       = 100
        always_advertise_network_route = false

        # MED configuration
        med = {
          always_compare_med          = true
          deterministic_med_comparison = true
        }

        # Graceful restart
        graceful_restart = {
          enable                   = true
          stale_route_time         = 120
          max_peer_restart_time    = 120
          local_restart_time       = 120
        }

        # BFD for fast convergence
        global_bfd = {
          profile = "bgp-bfd-profile"
        }

        # Advertise IPv4 networks
        advertise_network = {
          ipv4 = {
            network = [
              {
                name     = "10.0.0.0/8"
                unicast  = true
              },
              {
                name     = "192.168.0.0/16"
                unicast  = true
              }
            ]
          }
        }

        # EBGP peer group for external peers
        peer_group = [
          {
            name   = "upstream-providers"
            enable = true
            type = {
              ebgp = {}
            }
            address_family = {
              ipv4 = "bgp-ipv4-unicast"
            }
            connection_options = {
              multihop = 2
              timers   = "bgp-timers"
            }
            peer = [
              {
                name   = "isp-primary"
                enable = true
                peer_as = "64512"
                local_address = {
                  interface = "ethernet1/1"
                  ip        = "192.0.2.1"
                }
                peer_address = {
                  ip = "192.0.2.254"
                }
                connection_options = {
                  authentication = "bgp-md5-auth"
                }
                inherit = {
                  yes = {}
                }
              }
            ]
          },
          {
            name   = "internal-routers"
            enable = true
            type = {
              ibgp = {}
            }
            address_family = {
              ipv4 = "bgp-ipv4-unicast"
              ipv6 = "bgp-ipv6-unicast"
            }
            peer = [
              {
                name   = "core-router-1"
                enable = true
                peer_as = "65001"
                local_address = {
                  interface = "ethernet1/2"
                }
                peer_address = {
                  ip = "10.0.0.1"
                }
                inherit = {
                  yes = {}
                }
              }
            ]
          }
        ]

        # Route aggregation
        aggregate_routes = [
          {
            name    = "aggregate-10"
            enable  = true
            summary_only = true
            as_set  = true
            type = {
              ipv4 = {
                summary_prefix = "10.0.0.0/8"
              }
            }
          }
        ]
      }
    }
  ]
}

# Example 3: Logical router with OSPF configuration
resource "panos_template" "ospf_lr" {
  location = {
    panorama = {}
  }
  name = "ospf-lr-template"
}

resource "panos_logical_router" "ospf" {
  location = {
    template = {
      name = panos_template.ospf_lr.name
    }
  }

  name = "lr-ospf"

  vrf = [
    {
      name      = "default"
      interface = ["ethernet1/1", "ethernet1/2", "ethernet1/3"]

      # OSPF configuration
      ospf = {
        enable             = true
        router_id          = "1.1.1.1"
        rfc1583            = false
        global_if_timer    = "ospf-timer-profile"
        redistribution_profile = "ospf-redist-profile"

        # BFD for OSPF
        global_bfd = {
          profile = "ospf-bfd-profile"
        }

        # Graceful restart
        graceful_restart = {
          enable                      = true
          grace_period                = 120
          helper_enable               = true
          strict_lsa_checking         = false
          max_neighbor_restart_time   = 140
        }

        # OSPF Area 0 (backbone)
        area = [
          {
            name = "0.0.0.0"
            type = {
              normal = {
                abr = {
                  import_list = "area0-import"
                  export_list = "area0-export"
                }
              }
            }
            interface = [
              {
                name     = "ethernet1/1"
                enable   = true
                passive  = false
                priority = 100
                metric   = 10
                link_type = {
                  broadcast = {}
                }
                bfd = {
                  profile = "Inherit-lr-global-setting"
                }
              },
              {
                name     = "ethernet1/2"
                enable   = true
                passive  = false
                priority = 50
                metric   = 20
                link_type = {
                  p2p = {}
                }
              }
            ]
          },
          {
            name = "0.0.0.1"
            type = {
              stub = {
                no_summary = true
                abr = {
                  import_list = "area1-import"
                }
                default_route_metric = 100
              }
            }
            interface = [
              {
                name     = "ethernet1/3"
                enable   = true
                passive  = true
                metric   = 10
              }
            ]
          }
        ]
      }
    }
  ]
}

# Example 4: Logical router with ECMP configuration
resource "panos_template" "ecmp_lr" {
  location = {
    panorama = {}
  }
  name = "ecmp-lr-template"
}

resource "panos_logical_router" "ecmp" {
  location = {
    template = {
      name = panos_template.ecmp_lr.name
    }
  }

  name = "lr-ecmp"

  vrf = [
    {
      name      = "default"
      interface = ["ethernet1/1", "ethernet1/2", "ethernet1/3", "ethernet1/4"]

      # ECMP configuration for load balancing
      ecmp = {
        enable              = true
        max_paths           = 4
        symmetric_return    = true
        strict_source_path  = false
        algorithm = {
          ip_hash = {
            src_only  = false
            use_port  = true
            hash_seed = 12345
          }
        }
      }

      # Multiple static routes to same destination for ECMP
      routing_table = {
        ip = {
          static_route = [
            {
              name        = "ecmp-path-1"
              destination = "0.0.0.0/0"
              interface   = "ethernet1/1"
              nexthop = {
                ip_address = "10.1.1.1"
              }
              metric = 10
            },
            {
              name        = "ecmp-path-2"
              destination = "0.0.0.0/0"
              interface   = "ethernet1/2"
              nexthop = {
                ip_address = "10.1.2.1"
              }
              metric = 10
            },
            {
              name        = "ecmp-path-3"
              destination = "0.0.0.0/0"
              interface   = "ethernet1/3"
              nexthop = {
                ip_address = "10.1.3.1"
              }
              metric = 10
            },
            {
              name        = "ecmp-path-4"
              destination = "0.0.0.0/0"
              interface   = "ethernet1/4"
              nexthop = {
                ip_address = "10.1.4.1"
              }
              metric = 10
            }
          ]
        }
      }

      # BGP with ECMP multi-AS support
      bgp = {
        enable        = true
        router_id     = "192.0.2.1"
        local_as      = "65001"
        install_route = true
        ecmp_multi_as = true
      }
    }
  ]
}

# Example 5: Logical router with multicast (PIM) configuration
resource "panos_template" "multicast_lr" {
  location = {
    panorama = {}
  }
  name = "multicast-lr-template"
}

resource "panos_logical_router" "multicast" {
  location = {
    template = {
      name = panos_template.multicast_lr.name
    }
  }

  name = "lr-multicast"

  vrf = [
    {
      name      = "default"
      interface = ["ethernet1/1", "ethernet1/2"]

      # Multicast configuration with PIM
      multicast = {
        enable = true

        # Multicast static routes
        static_route = [
          {
            name        = "mcast-route-1"
            destination = "239.0.0.0/8"
            interface   = "ethernet1/1"
            nexthop = {
              ip_address = "10.1.1.1"
            }
            preference = 100
          }
        ]

        # PIM sparse mode configuration
        pim = {
          enable             = true
          rpf_lookup_mode    = "mrib-then-urib"
          route_ageout_time  = 210
          if_timer_global    = "pim-timer-global"
          group_permission   = "pim-group-acl"

          # SSM address space
          ssm_address_space = {
            group_list = "ssm-groups"
          }

          # Rendezvous Point configuration
          rp = {
            local_rp = {
              static_rp = {
                interface  = "ethernet1/1"
                address    = "10.1.1.1"
                override   = false
                group_list = "rp-groups"
              }
            }
            external_rp = [
              {
                name       = "192.0.2.1"
                group_list = "external-rp-groups"
                override   = false
              }
            ]
          }

          # SPT threshold
          spt_threshold = [
            {
              name      = "239.0.0.0/8"
              threshold = "0"
            }
          ]

          # PIM interfaces
          interface = [
            {
              name           = "ethernet1/1"
              dr_priority    = 100
              send_bsm       = true
              neighbor_filter = "pim-neighbor-filter"
            },
            {
              name        = "ethernet1/2"
              dr_priority = 50
            }
          ]
        }

        # IGMP configuration
        igmp = {
          enable = true
          dynamic = {
            interface = [
              {
                name               = "ethernet1/1"
                version            = "3"
                robustness         = "2"
                group_filter       = "igmp-group-filter"
                max_groups         = "500"
                max_sources        = "unlimited"
                router_alert_policing = true
              }
            ]
          }
          static = [
            {
              name           = "static-group-1"
              interface      = "ethernet1/2"
              group_address  = "239.1.1.1/32"
              source_address = "10.1.1.100"
            }
          ]
        }
      }
    }
  ]
}

# Example 6: Logical router with IPv6 and OSPFv3
resource "panos_template" "ipv6_lr" {
  location = {
    panorama = {}
  }
  name = "ipv6-lr-template"
}

resource "panos_logical_router" "ipv6" {
  location = {
    template = {
      name = panos_template.ipv6_lr.name
    }
  }

  name = "lr-ipv6"

  vrf = [
    {
      name      = "default"
      interface = ["ethernet1/1", "ethernet1/2"]

      # IPv6 static routes
      routing_table = {
        ipv6 = {
          static_route = [
            {
              name        = "ipv6-default"
              destination = "::/0"
              interface   = "ethernet1/1"
              nexthop = {
                ipv6_address = "2001:db8::1"
              }
              metric                  = 10
              administrative_distance = 10
            },
            {
              name        = "ipv6-internal"
              destination = "2001:db8:1::/48"
              interface   = "ethernet1/2"
              nexthop = {
                ipv6_address = "2001:db8:2::1"
              }
              metric = 10
            }
          ]
        }
      }

      # OSPFv3 for IPv6
      ospfv3 = {
        enable                   = true
        router_id                = "1.1.1.1"
        disable_transit_traffic  = false
        global_if_timer          = "ospfv3-timer"
        redistribution_profile   = "ospfv3-redist"

        global_bfd = {
          profile = "ospfv3-bfd-profile"
        }

        graceful_restart = {
          enable                    = true
          grace_period              = 120
          helper_enable             = true
          strict_lsa_checking       = false
          max_neighbor_restart_time = 140
        }

        area = [
          {
            name = "0.0.0.0"
            type = {
              normal = {}
            }
            interface = [
              {
                name     = "ethernet1/1"
                enable   = true
                passive  = false
                priority = 100
                metric   = 10
                link_type = {
                  broadcast = {}
                }
              }
            ]
          }
        ]
      }

      # BGP with IPv6 support
      bgp = {
        enable        = true
        router_id     = "192.0.2.1"
        local_as      = "65001"
        install_route = true

        advertise_network = {
          ipv6 = {
            network = [
              {
                name    = "2001:db8::/32"
                unicast = true
              }
            ]
          }
        }
      }
    }
  ]
}

# Example 7: Using vsys location
resource "panos_logical_router" "vsys_location" {
  location = {
    vsys = {
      name = "vsys1"
    }
  }

  name = "lr-vsys"

  vrf = [
    {
      name      = "default"
      interface = ["ethernet1/1"]

      routing_table = {
        ip = {
          static_route = [
            {
              name        = "default"
              destination = "0.0.0.0/0"
              interface   = "ethernet1/1"
              nexthop = {
                ip_address = "192.168.1.1"
              }
            }
          ]
        }
      }
    }
  ]
}
