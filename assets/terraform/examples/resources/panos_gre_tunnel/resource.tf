provider "panos" {
  # It is recommended to configure the provider through
  # environment variables PANOS_HOSTNAME, PANOS_USERNAME,
  # and PANOS_PASSWORD.
}

resource "panos_ethernet_interface" "ethif_1" {
  location = { ngfw = {} }
  name = "ethernet1/2"
  layer3 = {
    ips = [{
      name = "4.4.4.4/24"
    }]
  }
}

resource "panos_tunnel_interface" "tunnelif_1" {
  location = { ngfw = {} }
  name = "tunnel.2"
}

resource "panos_gre_tunnel" "gre_full" {
  location = { ngfw = {} }
  name               = "gre-tunnel-full"
  tunnel_interface   = panos_tunnel_interface.tunnelif_1.name
  copy_tos           = true
  disabled           = false
  ttl                = 200
  keep_alive = {
    enable     = true
    interval   = 30
    retry      = 4
    hold_timer = 15
  }
  local_address = {
    interface = panos_ethernet_interface.ethif_1.name
    ip        = panos_ethernet_interface.ethif_1.layer3.ips[0].name
  }
  peer_address = {
    ip = "5.5.5.5"
  }
}

resource "panos_ethernet_interface" "ethif_2" {
  location = { ngfw = {} }
  name = "ethernet1/3"
  layer3 = {
    ips = [{
      name = "8.8.8.8/24"
    }]
  }
}

resource "panos_gre_tunnel" "gre_floating_ip" {
  location = { ngfw = {} }
  name = "gre-tunnel-floating-ip"
  local_address = {
    interface   = panos_ethernet_interface.ethif_2.name
    floating_ip = "6.6.6.6"
  }
  peer_address = {
    ip = "7.7.7.7"
  }
}