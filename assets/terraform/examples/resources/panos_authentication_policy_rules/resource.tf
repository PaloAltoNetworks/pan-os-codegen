# Manage a group of authentication policy rules with positioning

## Place the rule group at the top of the pre-rulebase
resource "panos_authentication_policy_rules" "guest_network" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "pre-rulebase"
    }
  }

  position = {
    where = "first"
  }

  rules = [
    {
      name                       = "guest-wifi-auth"
      description                = "Require authentication for guest WiFi users"
      source_zones               = ["guest-zone"]
      source_addresses           = ["guest-network"]
      destination_zones          = ["untrust"]
      destination_addresses      = ["any"]
      services                   = ["any"]
      authentication_enforcement = "guest-captive-portal"
      timeout                    = 480
      log_authentication_timeout = true
      log_setting                = "authentication-log-profile"
    }
  ]
}

## Place the rule group after a specific rule
resource "panos_authentication_policy_rules" "corporate_users" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "pre-rulebase"
    }
  }

  position = {
    where    = "after"
    directly = true
    pivot    = "guest-wifi-auth"
  }

  rules = [
    {
      name                       = "employee-byod-auth"
      description                = "Authentication for employee BYOD devices"
      source_zones               = ["byod-zone"]
      source_addresses           = ["byod-subnet"]
      source_users               = ["any"]
      destination_zones          = ["internal", "dmz"]
      destination_addresses      = ["corporate-apps"]
      services                   = ["any"]
      category                   = ["business-and-economy", "computer-and-internet-info"]
      authentication_enforcement = "corporate-auth-profile"
      timeout                    = 1440
      log_authentication_timeout = false
      tags                       = ["byod", "corporate"]
    },
    {
      name                       = "contractor-limited-access"
      description                = "Authentication for contractors with restricted access"
      source_zones               = ["contractor-zone"]
      source_addresses           = ["contractor-subnet"]
      source_users               = ["contractor-group"]
      destination_zones          = ["dmz"]
      destination_addresses      = ["contractor-apps"]
      services                   = ["service-https"]
      authentication_enforcement = "contractor-auth-profile"
      timeout                    = 240
      log_authentication_timeout = true
      log_setting                = "authentication-log-profile"
      tags                       = ["contractor", "restricted"]
    }
  ]
}

## Advanced rule with HIP checks and target restrictions
resource "panos_authentication_policy_rules" "hip_based_auth" {
  location = {
    device_group = {
      name     = panos_device_group.example.name
      rulebase = "post-rulebase"
    }
  }

  position = {
    where = "last"
  }

  rules = [
    {
      name                       = "hip-compliant-devices"
      description                = "Allow authenticated access only for HIP-compliant devices"
      source_zones               = ["trust"]
      source_addresses           = ["corporate-subnets"]
      source_hip                 = ["compliant-hip-profile"]
      destination_zones          = ["dmz", "internal"]
      destination_addresses      = ["sensitive-servers"]
      destination_hip            = ["any"]
      services                   = ["any"]
      source_users               = ["domain\\authenticated-users"]
      authentication_enforcement = "mfa-auth-profile"
      timeout                    = 720
      log_authentication_timeout = true
      log_setting                = "security-log-profile"

      # Target specific devices in the device group
      target = {
        devices = [
          {
            name = "fw-datacenter-01"
            vsys = [
              { name = "vsys1" }
            ]
          },
          {
            name = "fw-datacenter-02"
            vsys = [
              { name = "vsys1" },
              { name = "vsys2" }
            ]
          }
        ]
        negate = false
        tags   = ["production"]
      }

      tags = ["hip-required", "production", "authenticated"]
    },
    {
      name                       = "non-compliant-redirect"
      description                = "Redirect non-compliant devices to remediation portal"
      source_zones               = ["trust"]
      source_addresses           = ["corporate-subnets"]
      negate_source              = false
      destination_zones          = ["remediation"]
      destination_addresses      = ["remediation-portal"]
      negate_destination         = false
      services                   = ["service-http", "service-https"]
      authentication_enforcement = "remediation-auth-profile"
      timeout                    = 60
      log_authentication_timeout = true
      disabled                   = false
      tags                       = ["remediation"]
    }
  ]
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
