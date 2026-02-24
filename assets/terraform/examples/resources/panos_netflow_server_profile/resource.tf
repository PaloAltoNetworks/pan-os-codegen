resource "panos_template" "example" {
  location = { panorama = {} }
  name = "example-template"
}

resource "panos_netflow_server_profile" "example" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "example-profile"
  active_timeout = 10
  export_enterprise_fields = true
  servers = [
    {
      name = "server1",
      host = "192.168.1.1",
      port = 2055
    }
  ]
  template_refresh_rate = {
    minutes = 20
    packets = 30
  }
}