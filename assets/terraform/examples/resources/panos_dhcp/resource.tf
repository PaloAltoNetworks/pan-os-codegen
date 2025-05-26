# Configure a template
resource "panos_template" "example" {
  location = {
    panorama = {}
  }
  name = "example-template"
}

# Configure ethernet interfaces
resource "panos_ethernet_interface" "eth1" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = "ethernet1/1"
  layer3   = {}
}

resource "panos_ethernet_interface" "eth2" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = "ethernet1/2"
  layer3   = {}
}

resource "panos_ethernet_interface" "eth3" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = "ethernet1/3"
  layer3   = {}
}

resource "panos_ethernet_interface" "eth4" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = "ethernet1/4"
  layer3   = {}
}

resource "panos_ethernet_interface" "eth5" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = "ethernet1/5"
  layer3   = {}
}

# DHCP Relay configuration
resource "panos_dhcp" "relay_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
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
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = panos_ethernet_interface.eth2.name

  relay = {
    ipv6 = {
      enabled = true
      server = [
        {
          name      = "2001:db8::1"
          interface = panos_ethernet_interface.eth1.name
        },
        {
          name = "2001:db8::2"
        }
      ]
    }
  }
}

# DHCP Server configuration with various options
resource "panos_dhcp" "server_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = panos_ethernet_interface.eth3.name

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
        name        = "192.168.1.100"
        mac         = "00:11:22:33:44:55"
        description = "reserved-printer"
      }
    ]
  }
}

# DHCP Server configuration with unlimited lease
resource "panos_dhcp" "server_unlimited_lease_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = panos_ethernet_interface.eth4.name

  server = {
    ip_pool = ["192.168.2.0/24"]
    option = {
      lease = {
        unlimited = {}
      }
    }
  }
}

# DHCP Server configuration with inheritance
resource "panos_dhcp" "server_inheritance_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  name     = panos_ethernet_interface.eth5.name

  server = {
    ip_pool = ["192.168.3.0/24"]
    option = {
      inheritance = {
        source = panos_ethernet_interface.eth1.name
      }
    }
  }
}
