# Create a template
resource "panos_template" "bgp_template" {
  location = { panorama = {} }
  name     = "bgp-routing-template"
}

# Basic IPv4 Unicast BGP Address Family Profile
resource "panos_bgp_address_family_routing_profile" "ipv4_unicast_basic" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-unicast-profile"

  ipv4 = {
    unicast = {
      enable                         = true
      default_originate              = true
      route_reflector_client         = false
      soft_reconfig_with_stored_info = true
    }
  }
}

# IPv4 Unicast with Community Attributes
resource "panos_bgp_address_family_routing_profile" "ipv4_with_community" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-with-community"

  ipv4 = {
    unicast = {
      enable            = true
      as_override       = true
      default_originate = true
      send_community = {
        all = {}
      }
    }
  }
}

# IPv6 Unicast BGP Address Family Profile
resource "panos_bgp_address_family_routing_profile" "ipv6_unicast" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv6-unicast-profile"

  ipv6 = {
    unicast = {
      enable                         = true
      default_originate              = true
      soft_reconfig_with_stored_info = true
      add_path = {
        tx_all_paths = true
      }
    }
  }
}

# IPv4 Multicast BGP Address Family Profile
resource "panos_bgp_address_family_routing_profile" "ipv4_multicast" {
  location = {
    template = {
      name = panos_template.bgp_template.name
    }
  }

  name = "ipv4-multicast-profile"

  ipv4 = {
    multicast = {
      enable                 = true
      route_reflector_client = true
      orf = {
        orf_prefix_list = "both"
      }
    }
  }
}
