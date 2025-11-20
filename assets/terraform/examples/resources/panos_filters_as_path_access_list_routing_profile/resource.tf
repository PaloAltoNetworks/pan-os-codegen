# Create a template for the AS Path Access List profiles
resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "as-path-acl-template"
}

# Basic AS Path Access List Profile
resource "panos_filters_as_path_access_list_routing_profile" "basic" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "basic-as-path-acl"
  description = "Basic AS Path Access List for filtering BGP routes"

  aspath_entries = [
    {
      name         = "1"
      action       = "deny"
      aspath_regex = "^65000_"
    }
  ]
}

# AS Path Access List with Permit Action
resource "panos_filters_as_path_access_list_routing_profile" "permit_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "permit-as-path-acl"
  description = "AS Path Access List permitting specific AS paths"

  aspath_entries = [
    {
      name         = "1"
      action       = "permit"
      aspath_regex = "^65100_"
    }
  ]
}

# AS Path Access List with Multiple Entries
resource "panos_filters_as_path_access_list_routing_profile" "multiple_entries" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "multi-entry-as-path-acl"
  description = "AS Path Access List with multiple filtering rules"

  aspath_entries = [
    {
      name         = "1"
      action       = "permit"
      aspath_regex = "^65000_"
    },
    {
      name         = "2"
      action       = "deny"
      aspath_regex = "_65100$"
    },
    {
      name         = "3"
      action       = "permit"
      aspath_regex = "^65200_.*_65300$"
    }
  ]
}

# Complex AS Path Access List with Advanced Regex Patterns
resource "panos_filters_as_path_access_list_routing_profile" "advanced" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "advanced-as-path-acl"
  description = "Advanced AS Path Access List with complex regex patterns"

  aspath_entries = [
    {
      name         = "1"
      action       = "permit"
      aspath_regex = "^$" # Empty AS path (locally originated routes)
    },
    {
      name         = "2"
      action       = "deny"
      aspath_regex = "_64512_" # Deny paths containing AS 64512
    },
    {
      name         = "3"
      action       = "permit"
      aspath_regex = "^65[0-9]{3}_" # Permit paths starting with AS 65xxx
    },
    {
      name         = "4"
      action       = "deny"
      aspath_regex = ".*" # Deny all other paths
    }
  ]
}

# AS Path Access List on NGFW Device
resource "panos_filters_as_path_access_list_routing_profile" "ngfw_example" {
  location = {
    ngfw = {
      ngfw_device = "localhost.localdomain"
    }
  }

  name        = "ngfw-as-path-acl"
  description = "AS Path Access List configured on NGFW device"

  aspath_entries = [
    {
      name         = "1"
      action       = "permit"
      aspath_regex = "^65001_65002_"
    },
    {
      name         = "2"
      action       = "deny"
      aspath_regex = "_65003_"
    }
  ]
}
