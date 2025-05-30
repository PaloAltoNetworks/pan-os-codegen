resource "panos_template" "example" {
  name     = "example-template"
  location = { panorama = {} }
}

resource "panos_aggregate_interface" "parent" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }
  name   = "ae1"
  layer3 = {}
}

resource "panos_interface_management_profile" "profile" {
  location = { template = { name = panos_template.example.name } }
  name     = "example-profile"
}

resource "panos_aggregate_layer3_subinterface" "example1" {
  location = { template = { name = panos_template.example.name, vsys = "vsys1" } }
  parent   = panos_aggregate_interface.parent.name
  name     = "ae1.1"
  tag      = 1
  comment  = "Basic aggregate layer3 subinterface"

  adjust_tcp_mss = {
    enable              = true
    ipv4_mss_adjustment = 40
    ipv6_mss_adjustment = 60
  }

  arp = [{
    name       = "192.0.2.1"
    hw_address = "00:1a:2b:3c:4d:5e"
  }]

  bonjour = {
    enable    = true
    group_id  = 0
    ttl_check = true
  }

  decrypt_forward = false
  df_ignore       = true

  interface_management_profile = panos_interface_management_profile.profile.name

  ip = [{
    name          = "192.0.2.1/24"
    sdwan_gateway = "10.0.0.1"
  }]

  mtu = 1500
}

resource "panos_aggregate_layer3_subinterface" "example2" {
  location = { template = { name = panos_template.example.name, vsys = "vsys1" } }
  parent   = panos_aggregate_interface.parent.name
  name     = "ae1.2"
  tag      = 2
  comment  = "Aggregate layer3 subinterface with DHCP client"

  dhcp_client = {
    enable               = true
    create_default_route = true
    default_route_metric = 10
    send_hostname = {
      enable   = true
      hostname = "system-hostname"
    }
  }

  interface_management_profile = panos_interface_management_profile.profile.name
}

resource "panos_aggregate_layer3_subinterface" "example3" {
  location = { template = { name = panos_template.example.name, vsys = "vsys1" } }
  parent   = panos_aggregate_interface.parent.name
  name     = "ae1.3"
  tag      = 3
  comment  = "Aggregate layer3 subinterface with IPv6 address"

  ipv6 = {
    enabled      = true
    interface_id = "EUI-64"
    address = [{
      name                = "2001:db8::1/64"
      enable_on_interface = true
      advertise = {
        enable             = true
        valid_lifetime     = "2592000"
        preferred_lifetime = "604800"
        onlink_flag        = true
        auto_config_flag   = true
      }
    }]
  }

  mtu = 1500
  ndp_proxy = {
    enabled = true
    address = [{
      name   = "2001:db8::/64"
      negate = false
    }]
  }

  interface_management_profile = panos_interface_management_profile.profile.name
}

resource "panos_aggregate_layer3_subinterface" "example4" {
  location = { template = { name = panos_template.example.name, vsys = "vsys1" } }
  parent   = panos_aggregate_interface.parent.name
  name     = "ae1.4"
  tag      = 4
  comment  = "Aggregate layer3 subinterface with IPv6 inherited configuration"

  ipv6 = {
    enabled      = true
    interface_id = "EUI-64"
    inherited = {
      enable = true
      assign_addr = [
        {
          name = "gua-address-dynamic"
          type = {
            gua = {
              enable_on_interface = true
              pool_type = {
                dynamic = {}
              }
              advertise = {
                enable           = true
                onlink_flag      = true
                auto_config_flag = true
              }
            }
          }
        },
        {
          name = "ula-address"
          type = {
            ula = {
              enable_on_interface = true
              address             = "fd00:1234:5678::/48"
              prefix              = true
              anycast             = false
              advertise = {
                enable             = true
                valid_lifetime     = "2592000"
                preferred_lifetime = "604800"
                onlink_flag        = true
                auto_config_flag   = true
              }
            }
          }
        }
      ]
      neighbor_discovery = {
        dad_attempts       = 1
        enable_dad         = true
        ns_interval        = 1000
        reachable_time     = 30000
        enable_ndp_monitor = true
        router_advertisement = {
          enable                   = true
          hop_limit                = "64"
          lifetime                 = 1800
          managed_flag             = true
          max_interval             = 600
          min_interval             = 200
          other_flag               = true
          router_preference        = "Medium"
          enable_consistency_check = true
        }
      }
    }
  }

  interface_management_profile = panos_interface_management_profile.profile.name
}

resource "panos_aggregate_layer3_subinterface" "example5" {
  location = { template = { name = panos_template.example.name, vsys = "vsys1" } }
  parent   = panos_aggregate_interface.parent.name
  name     = "ae1.5"
  tag      = 5
  comment  = "Aggregate layer3 subinterface with IPv6 neighbor discovery"

  ipv6 = {
    enabled      = true
    interface_id = "EUI-64"
    neighbor_discovery = {
      dad_attempts       = 1
      enable_dad         = true
      ns_interval        = 1000
      reachable_time     = 30000
      enable_ndp_monitor = true
      neighbor = [{
        name       = "2001:db8::1"
        hw_address = "00:1a:2b:3c:4d:5e"
      }]
      router_advertisement = {
        enable                   = true
        hop_limit                = "64"
        lifetime                 = 1800
        managed_flag             = false
        max_interval             = 600
        min_interval             = 200
        other_flag               = false
        router_preference        = "Medium"
        enable_consistency_check = true
        dns_support = {
          enable = true
          server = [{
            name     = "2001:db8::53"
            lifetime = 1200
          }]
          suffix = [{
            name     = "example.com"
            lifetime = 1200
          }]
        }
      }
    }
  }

  mtu = 1500

  interface_management_profile = panos_interface_management_profile.profile.name
}
