resource "panos_ntp_settings" "example" {
  location = {
    system = {}
  }

  ntp_servers = {
    primary_ntp_server = {
      ntp_server_address = "1.1.1.1"
    }
    secondary_ntp_server = {
      ntp_server_address = "2.2.2.2"
    }

  }
}