resource "panos_dns_settings" "example" {
  location = {
    system = {}
  }

  dns_settings = {
    servers = {
      primary   = "8.8.8.8"
      secondary = "1.1.1.1"
    }
  }
}
