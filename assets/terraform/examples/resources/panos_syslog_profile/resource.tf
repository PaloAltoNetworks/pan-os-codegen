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

  name = "example-syslog-profile"

  servers = [
    {
      name      = "server1"
      server    = "10.0.0.1"
      transport = "UDP"
      port      = 514
      facility  = "LOG_USER"
      format    = "IETF"
    },
    {
      name      = "server2"
      server    = "syslog.example.com"
      transport = "SSL"
      port      = 6514
      facility  = "LOG_LOCAL1"
      format    = "BSD"
    }
  ]

  format = {
    auth    = "auth-format"
    traffic = "traffic-format"
    escaping = {
      escape_character   = "\\"
      escaped_characters = "'"
    }
  }
}
