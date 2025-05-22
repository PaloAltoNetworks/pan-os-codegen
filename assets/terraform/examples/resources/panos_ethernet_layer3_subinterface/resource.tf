resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-tmpl"
}

resource "panos_ethernet_interface" "parent" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  name   = "ethernet1/1"
  layer3 = {}
}

resource "panos_ethernet_layer3_subinterface" "example1" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent          = panos_ethernet_interface.parent.name
  name            = "ethernet1/1.1"
  tag             = 1
  comment         = "Basic subinterface"
  netflow_profile = "NetflowProfile1"
  mtu             = 1500
  adjust_tcp_mss  = { enable = true, ipv4_mss_adjustment = 1300, ipv6_mss_adjustment = 1300 }
  arp             = [{ name = "192.168.0.1", hw_address = "00:1a:2b:3c:4d:5e" }]
  bonjour         = { enable = true, group_id = 5, ttl_check = true }
  decrypt_forward = true
  df_ignore       = true
  ndp_proxy       = { enabled = true, address = [{ name = "10.0.0.1", negate = false }] }
  ip              = [{ name = "192.168.1.1", sdwan_gateway = "192.168.1.1" }]
}

resource "panos_ethernet_layer3_subinterface" "example2" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.2"
  tag     = 2
  comment = "IPv6 GUA subinterface"
  ipv6 = {
    enabled = true
    inherited = {
      assign_addr = [
        {
          name = "gua_config"
          type = {
            gua = {
              enable_on_interface = true
              prefix_pool         = "my-gua-pool"
            }
          }
        }
      ]
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example3" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.3"
  tag     = 3
  comment = "IPv6 ULA subinterface"
  ipv6 = {
    enabled = true
    inherited = {
      assign_addr = [
        {
          name = "ula_config"
          type = {
            ula = {
              enable_on_interface = true
              address             = "fd00:1234:5678::/48"
            }
          }
        }
      ]
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example4" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.4"
  tag     = 4
  comment = "SDWAN DDNS subinterface"
  sdwan_link_settings = {
    enable                  = true
    sdwan_interface_profile = "SdwanProfile1"
    upstream_nat = {
      enable = true
      ddns   = {}
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example5" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.5"
  tag     = 5
  comment = "SDWAN Static IP FQDN subinterface"
  sdwan_link_settings = {
    enable                  = true
    sdwan_interface_profile = "SdwanProfile1"
    upstream_nat = {
      enable = true
      static_ip = {
        fqdn = "example.com"
      }
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example6" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.6"
  tag     = 6
  comment = "SDWAN Static IP Address subinterface"
  sdwan_link_settings = {
    enable                  = true
    sdwan_interface_profile = "SdwanProfile1"
    upstream_nat = {
      enable = true
      static_ip = {
        ip_address = "203.0.113.1"
      }
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example7" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.7"
  tag     = 7
  comment = "DHCP Client subinterface"
  dhcp_client = {
    create_default_route = true
    default_route_metric = 10
    enable               = true
    send_hostname = {
      enable   = true
      hostname = "dhcp-client-hostname"
    }
  }
  interface_management_profile = "dhcp-client-profile"
  ipv6                         = { enabled = false }
  sdwan_link_settings          = { enable = false }
}

resource "panos_ethernet_layer3_subinterface" "example8" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.8"
  tag     = 8
  comment = "IPv6 DHCP Client subinterface"
  ipv6 = {
    enabled = true
    dhcp_client = {
      accept_ra_route      = true
      default_route_metric = 10
      enable               = true
      neighbor_discovery = {
        dad_attempts       = 1
        enable_dad         = true
        enable_ndp_monitor = true
        ns_interval        = 1000
        reachable_time     = 30000
      }
      preference = "high"
      prefix_delegation = {
        enable = {
          yes = {
            pfx_pool_name   = "prefix-pool-1"
            prefix_len      = 64
            prefix_len_hint = true
          }
        }
      }
      v6_options = {
        duid_type = "duid-type-llt"
        enable = {
          yes = {
            non_temp_addr = true
            temp_addr     = false
          }
        }
        rapid_commit          = true
        support_srvr_reconfig = true
      }
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example9" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.9"
  tag     = 9
  comment = "IPv6 Neighbor Discovery subinterface"
  ipv6 = {
    enabled = true
    neighbor_discovery = {
      dad_attempts       = 1
      enable_dad         = true
      enable_ndp_monitor = true
      ns_interval        = 1000
      reachable_time     = 30000
      neighbor = [
        {
          name       = "2001:DB8::1/128"
          hw_address = "00:1a:2b:3c:4d:5e"
        }
      ]
      router_advertisement = {
        enable                   = true
        enable_consistency_check = true
        hop_limit                = "64"
        lifetime                 = 1800
        link_mtu                 = "1500"
        managed_flag             = true
        max_interval             = 600
        min_interval             = 200
        other_flag               = true
        reachable_time           = "0"
        retransmission_timer     = "0"
        router_preference        = "Medium"
        dns_support = {
          enable = true
          server = [
            {
              name     = "2001:DB8::1/128"
              lifetime = 1200
            }
          ]
          suffix = [
            {
              name     = "suffix1"
              lifetime = 1200
            }
          ]
        }
      }
    }
  }
}

resource "panos_ethernet_layer3_subinterface" "example10" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  parent  = panos_ethernet_interface.parent.name
  name    = "ethernet1/1.10"
  tag     = 10
  comment = "PPPoE subinterface"
  pppoe = {
    access_concentrator  = "ac-1"
    authentication       = "auto"
    create_default_route = true
    default_route_metric = 10
    enable               = true
    passive = {
      enable = true
    }
    password = "pppoe-password"
    service  = "pppoe-service"
    static_address = {
      ip = "192.168.2.1"
    }
    username = "pppoe-user"
  }
  ipv6                = { enabled = false }
  sdwan_link_settings = { enable = false }
}
