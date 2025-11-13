# Create a template
resource "panos_template" "bgp_template" {
  location = { panorama = {} }
  name     = "bgp-routing-template"
}

# IPv4 Unicast - Redistribute Connected Routes
resource "panos_bgp_redistribution_routing_profile" "ipv4_connected" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-redistribute-connected"

  ipv4 = {
    unicast = {
      connected = {
        enable = true
        metric = 100
      }
    }
  }
}

# IPv4 Unicast - Redistribute OSPF Routes
resource "panos_bgp_redistribution_routing_profile" "ipv4_ospf" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-redistribute-ospf"

  ipv4 = {
    unicast = {
      ospf = {
        enable = true
        metric = 200
      }
    }
  }
}

# IPv4 Unicast - Redistribute Static Routes
resource "panos_bgp_redistribution_routing_profile" "ipv4_static" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-redistribute-static"

  ipv4 = {
    unicast = {
      static = {
        enable = true
        metric = 150
      }
    }
  }
}

# IPv4 Unicast - Redistribute RIP Routes
resource "panos_bgp_redistribution_routing_profile" "ipv4_rip" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-redistribute-rip"

  ipv4 = {
    unicast = {
      rip = {
        enable = true
        metric = 175
      }
    }
  }
}

# IPv4 Unicast - Redistribute Multiple Sources
resource "panos_bgp_redistribution_routing_profile" "ipv4_multiple" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-redistribute-multiple"

  ipv4 = {
    unicast = {
      connected = {
        enable = true
        metric = 100
      }
      ospf = {
        enable = true
        metric = 200
      }
      static = {
        enable = true
        metric = 150
      }
    }
  }
}

# IPv6 Unicast - Redistribute Connected Routes
resource "panos_bgp_redistribution_routing_profile" "ipv6_connected" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv6-redistribute-connected"

  ipv6 = {
    unicast = {
      connected = {
        enable = true
        metric = 100
      }
    }
  }
}

# IPv6 Unicast - Redistribute OSPFv3 Routes
resource "panos_bgp_redistribution_routing_profile" "ipv6_ospfv3" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv6-redistribute-ospfv3"

  ipv6 = {
    unicast = {
      ospfv3 = {
        enable = true
        metric = 200
      }
    }
  }
}

# IPv6 Unicast - Redistribute Static Routes
resource "panos_bgp_redistribution_routing_profile" "ipv6_static" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv6-redistribute-static"

  ipv6 = {
    unicast = {
      static = {
        enable = true
        metric = 150
      }
    }
  }
}

# IPv6 Unicast - Redistribute Multiple Sources
resource "panos_bgp_redistribution_routing_profile" "ipv6_multiple" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv6-redistribute-multiple"

  ipv6 = {
    unicast = {
      connected = {
        enable = true
        metric = 100
      }
      ospfv3 = {
        enable = true
        metric = 200
      }
      static = {
        enable = true
        metric = 150
      }
    }
  }
}

# Using template-stack location
resource "panos_template_stack" "bgp_stack" {
  location = { panorama = {} }
  name     = "bgp-routing-stack"
}

resource "panos_bgp_redistribution_routing_profile" "template_stack_example" {
  location = {
    template_stack = {
      name = panos_template_stack.bgp_stack.name
    }
  }

  name = "stack-redistribute-profile"

  ipv4 = {
    unicast = {
      connected = {
        enable = true
        metric = 100
      }
    }
  }
}
