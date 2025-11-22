provider "panos" {}

resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

# Example 1: Monitor profile with default values (minimal configuration)
resource "panos_monitor_profile" "minimal" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "minimal-monitor-profile"
}

# Example 2: Monitor profile with fail-over action
resource "panos_monitor_profile" "failover" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name      = "failover-monitor-profile"
  action    = "fail-over"
  interval  = 5
  threshold = 3
}

# Example 3: Monitor profile with wait-recover action and custom settings
resource "panos_monitor_profile" "custom" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name      = "custom-monitor-profile"
  action    = "wait-recover"
  interval  = 10
  threshold = 7
}
