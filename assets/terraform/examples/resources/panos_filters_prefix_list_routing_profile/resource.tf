# Create a template for the prefix list routing profiles
resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "prefix-list-template"
}

# IPv4 Prefix List Routing Profile
resource "panos_filters_prefix_list_routing_profile" "ipv4_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "ipv4-prefix-list"
  description = "IPv4 prefix list for filtering BGP routes"

  type = {
    ipv4 = {
      ipv4_entries = [
        {
          # Permit any prefix
          name   = "10"
          action = "permit"
          prefix = {
            network = "any"
          }
        },
        {
          # Deny specific network block
          name   = "20"
          action = "deny"
          prefix = {
            entry = {
              network = "10.0.0.0/8"
            }
          }
        },
        {
          # Permit 192.168.0.0/16 with prefix length between /24 and /28
          name   = "30"
          action = "permit"
          prefix = {
            entry = {
              network                = "192.168.0.0/16"
              greater_than_or_equal = 24
              less_than_or_equal    = 28
            }
          }
        },
        {
          # Deny 172.16.0.0/12 with minimum prefix length of /16
          name   = "40"
          action = "deny"
          prefix = {
            entry = {
              network                = "172.16.0.0/12"
              greater_than_or_equal = 16
            }
          }
        }
      ]
    }
  }
}

# IPv6 Prefix List Routing Profile
resource "panos_filters_prefix_list_routing_profile" "ipv6_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "ipv6-prefix-list"
  description = "IPv6 prefix list for filtering BGP routes"

  type = {
    ipv6 = {
      ipv6_entries = [
        {
          # Permit any IPv6 prefix
          name   = "10"
          action = "permit"
          prefix = {
            network = "any"
          }
        },
        {
          # Deny documentation prefix 2001:db8::/32
          name   = "20"
          action = "deny"
          prefix = {
            entry = {
              network = "2001:db8::/32"
            }
          }
        },
        {
          # Permit ULA prefixes fd00::/8 with prefix length between /48 and /64
          name   = "30"
          action = "permit"
          prefix = {
            entry = {
              network                = "fd00::/8"
              greater_than_or_equal = 48
              less_than_or_equal    = 64
            }
          }
        },
        {
          # Deny fc00::/7 with maximum prefix length of /48
          name   = "40"
          action = "deny"
          prefix = {
            entry = {
              network             = "fc00::/7"
              less_than_or_equal = 48
            }
          }
        }
      ]
    }
  }
}
