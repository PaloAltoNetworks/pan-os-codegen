# Authentication policy rules in shared location (Panorama or NGFW)
# This configuration applies to shared rulebase that is evaluated before vsys-specific rules

resource "panos_authentication_policy_rules" "shared_baseline" {
  location = {
    shared = {
      rulebase = "pre-rulebase"
    }
  }

  position = {
    where = "first"
  }

  rules = [
    {
      name                       = "global-admin-auth"
      description                = "Require authentication for administrative access from any location"
      source_zones               = ["any"]
      source_addresses           = ["any"]
      destination_zones          = ["management"]
      destination_addresses      = ["admin-servers"]
      services                   = ["service-https", "ssh"]
      source_users               = ["any"]
      authentication_enforcement = "admin-mfa-profile"
      timeout                    = 60
      log_authentication_timeout = true
      log_setting                = "admin-log-profile"
      tags                       = ["admin", "critical"]
    },
    {
      name                       = "remote-access-baseline"
      description                = "Baseline authentication for all remote access"
      source_zones               = ["remote-access"]
      source_addresses           = ["any"]
      destination_zones          = ["any"]
      destination_addresses      = ["corporate-resources"]
      services                   = ["any"]
      authentication_enforcement = "remote-access-profile"
      timeout                    = 360
      log_authentication_timeout = true
      disabled                   = false
    }
  ]
}
