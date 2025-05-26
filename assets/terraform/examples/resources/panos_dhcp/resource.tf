# Configure a template
resource "panos_template" "example" {
  name = "example-template"
}

# Configure ethernet interfaces
resource "panos_ethernet_interface" "eth1" {
  template = panos_template.example.name
  name     = "ethernet1/1"
  vsys     = "vsys1"
  mode     = "layer3"
}

resource "panos_ethernet_interface" "eth2" {
  template = panos_template.example.name
  name     = "ethernet1/2"
  vsys     = "vsys1"
  mode     = "layer3"
}

# DHCP Relay configuration
resource "panos_dhcp" "relay_example" {
  template = panos_template.example.name
  name     = panos_ethernet_interface.eth1.name

  relay = {
    ip = {
      enabled = true
      server  = ["10.0.0.1", "10.0.0.2"]
    }
  }
}

# DHCP Relay IPv6 configuration
resource "panos_dhcp" "relay_ipv6_example" {
  template = panos_template.example.name
  name     = panos_ethernet_interface.eth2.name

  relay = {
    ipv6 = {
      enabled = true
      server = [
        {
          address   = "2001:db8::1"
          interface = panos_ethernet_interface.eth1.name
        },
        {
          address = "2001:db8::2"
        }
      ]
    }
  }
}

# DHCP Server configuration with various options
resource "panos_dhcp" "server_example" {
  template = panos_template.example.name
  name     = "ethernet1/3"

  server = {
    ip_pool = ["192.168.1.0/24"]
    mode    = "enabled"
    option = {
      dns = {
        primary   = "8.8.8.8"
        secondary = "8.8.4.4"
      }
      dns_suffix = "example.com"
      gateway    = "192.168.1.1"
      lease = {
        timeout = 720
      }
      nis = {
        primary   = "192.168.1.10"
        secondary = "192.168.1.11"
      }
      ntp = {
        primary   = "192.168.1.20"
        secondary = "192.168.1.21"
      }
      pop3_server = "192.168.1.30"
      smtp_server = "192.168.1.25"
      subnet_mask = "255.255.255.0"
      wins = {
        primary   = "192.168.1.40"
        secondary = "192.168.1.41"
      }
      user_defined = [
        {
          name      = "custom_ip_option"
          code      = 200
          ip        = ["10.0.0.1", "10.0.0.2"]
          inherited = false
        },
        {
          name      = "custom_ascii_option"
          code      = 201
          ascii     = ["custom option"]
          inherited = false
        },
        {
          name      = "custom_hex_option"
          code      = 202
          hex       = ["0A0B0C"]
          inherited = false
        }
      ]
    }
    reserved = [
      {
        name = "reserved-printer"
        mac  = "00:11:22:33:44:55"
        ip   = "192.168.1.100"
      }
    ]
  }
}

# DHCP Server configuration with unlimited lease
resource "panos_dhcp" "server_unlimited_lease_example" {
  template = panos_template.example.name
  name     = "ethernet1/4"

  server = {
    ip_pool = ["192.168.2.0/24"]
    option = {
      lease = {
        unlimited = true
      }
    }
  }
}

# DHCP Server configuration with inheritance
resource "panos_dhcp" "server_inheritance_example" {
  template = panos_template.example.name
  name     = "ethernet1/5"

  server = {
    ip_pool = ["192.168.3.0/24"]
    option = {
      inheritance = {
        source = panos_ethernet_interface.eth1.name
      }
    }
  }
}
