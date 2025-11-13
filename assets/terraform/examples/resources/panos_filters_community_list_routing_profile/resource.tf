# Create a template for the community list profiles
resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "community-list-template"
}

# Extended Community List Profile
resource "panos_filters_community_list_routing_profile" "extended" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "extended-community-list"
  description = "Extended BGP community list for filtering routes"

  type = {
    extended = {
      extended_entries = [
        {
          name                       = "1"
          action                     = "permit"
          extended_community_regexes = ["^100:.*", "^200:.*"]
        },
        {
          name                       = "2"
          action                     = "deny"
          extended_community_regexes = ["^300:.*"]
        }
      ]
    }
  }
}

# Large Community List Profile
resource "panos_filters_community_list_routing_profile" "large" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "large-community-list"
  description = "Large BGP community list for filtering routes"

  type = {
    large = {
      large_entries = [
        {
          name                      = "1"
          action                    = "permit"
          large_community_regexes = ["^1000:.*:.*", "^2000:.*:.*"]
        },
        {
          name                      = "2"
          action                    = "deny"
          large_community_regexes = ["^3000:.*:.*"]
        }
      ]
    }
  }
}

# Regular Community List Profile
resource "panos_filters_community_list_routing_profile" "regular" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "regular-community-list"
  description = "Regular BGP community list for filtering routes"

  type = {
    regular = {
      regular_entries = [
        {
          name        = "1"
          action      = "permit"
          communities = ["100:200", "300:400"]
        },
        {
          name        = "2"
          action      = "deny"
          communities = ["500:600"]
        }
      ]
    }
  }
}
