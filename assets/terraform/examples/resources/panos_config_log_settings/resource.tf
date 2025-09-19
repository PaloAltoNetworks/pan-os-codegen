resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

resource "panos_syslog_profile" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }
  
  name = "example-profile-1"
  servers = [{
    name   = "syslog-server1"
    server = "10.0.0.1"
  }]
}

resource "panos_config_log_settings" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name             = "example-config-settings"
  description      = "config log settings example"
  filter           = "(dgname eq default)"
  send_to_panorama = true
  syslog_profiles  = [panos_syslog_profile.example.name]
}
