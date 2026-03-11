# Authentication policy rule on NGFW in vsys location
resource "panos_authentication_policy" "ngfw_vsys" {
  location = {
    vsys = {
      name         = "vsys1"
      ngfw_device  = "localhost.localdomain"
    }
  }

  rules = [
    {
      name                       = "vpn-user-auth"
      description                = "Require authentication for VPN users accessing internal resources"
      source_zones               = ["vpn-zone"]
      source_addresses           = ["vpn-pool"]
      source_users               = ["any"]
      destination_zones          = ["internal"]
      destination_addresses      = ["internal-servers"]
      services                   = ["any"]
      authentication_enforcement = "vpn-auth-profile"
      timeout                    = 480
      log_authentication_timeout = true
      log_setting                = "vpn-log-profile"
      tags                       = ["vpn", "remote-access"]
    },
    {
      name                       = "internet-access-auth"
      description                = "Require authentication for internet access"
      source_zones               = ["trust"]
      source_addresses           = ["private-networks"]
      destination_zones          = ["untrust"]
      destination_addresses      = ["any"]
      services                   = ["web-browsing", "ssl"]
      category                   = ["any"]
      authentication_enforcement = "internet-auth-profile"
      timeout                    = 1440
      log_authentication_timeout = false
      group_tag                  = "internet-users"
    }
  ]
}
