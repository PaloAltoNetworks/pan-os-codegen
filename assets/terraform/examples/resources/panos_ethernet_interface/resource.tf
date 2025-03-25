resource "panos_ethernet_interface" "example" {
  location = {
    template = {
      vsys = "vsys1"
      name = panos_template.example.name
    }
  }

  name = "ethernet1/1"

  comment = "ethernet interface comment"

  layer3 = {
    cluster_interconnect = false
    decrypt_forward      = true
    df_ignore            = true
    mtu                  = 9000
    traffic_interconnect = true


    adjust_tcp_mss = {
      enable              = true
      ipv4_mss_adjustment = 40
      ipv6_mss_adjustment = 60
    }
    arp = [{
      name       = "192.168.1.100"
      hw_address = "aa:bb:cc:dd:ee:ff"
    }]
    bonjour = {
      enable    = true
      group_id  = 10
      ttl_check = true
    }
    ips = [{
      name = "192.168.1.10"
    }]
    ipv6 = {
      addresses = [{
        name                = "::1"
        enable_on_interface = true
        advertise = {
          auto_config_flag   = true
          enable             = true
          onlink_flag        = true
          preferred_lifetime = 1800
          valid_lifetime     = 3600
        }
      }]
      enabled = true
    }
    lldp = {
      enable = true
      high_availability = {
        passive_pre_negotiation = true
      }
    }
    ndp_proxy = {
      enabled = true
      addresses = [{
        name   = "10.0.0.0/24"
        negate = true
      }]
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
