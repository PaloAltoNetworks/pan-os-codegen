# Create a template
resource "panos_template" "gp_template" {
  location = { panorama = {} }
  name     = "global-protect-template"
}

# Create an ethernet interface
resource "panos_ethernet_interface" "gp_interface" {
  location = { template = { name = panos_template.gp_template.name } }
  name     = "ethernet1/1"
  layer3 = {
    ips = [{ name = "10.1.1.1/24" }]
  }
}

# Create a zone
resource "panos_zone" "gp_zone" {
  location = { template = { name = panos_template.gp_template.name } }
  name     = "gp-zone"
}

# GlobalProtect Portal with Client Config
resource "panos_globalprotect_portal" "client_config_example" {
  location = {
    template = {
      name = panos_template.gp_template.name
    }
  }
  name = "gp-portal-client-config"

  client_config = {
    agent_user_override_key = "key"
    configs = [
      {
        name                    = "config1"
        save_user_credentials   = "1"
        portal_2fa              = true
        internal_gateway_2fa    = true
        refresh_config          = true
        mdm_address             = "mdm.example.com"
        mdm_enrollment_port     = "443"
        source_user             = ["employee", "contractor"]
        third_party_vpn_clients = ["client1", "client2"]
        os                      = ["windows", "mac", "linux"]
        gateways = {
          external = {
            cutoff_time = 10
            list = [
              {
                name     = "external-gateway1"
                fqdn     = "gateway.example.com"
                priority = "1"
                manual   = true
              }
            ]
          }
        }
        internal_host_detection = {
          ip_address = "192.168.1.1"
          hostname   = "internal.example.com"
        }
        agent_ui = {
          passcode                    = "123456"
          uninstall_password          = "uninstall123"
          agent_user_override_timeout = 60
          max_agent_user_overrides    = 5
        }
        hip_collection = {
          max_wait_time    = 30
          collect_hip_data = true
        }
        authentication_override = {
          generate_cookie = true
          accept_cookie = {
            cookie_lifetime = {
              lifetime_in_hours = 24
            }
          }
        }
      }
    ]
  }
}

# GlobalProtect Portal with Clientless VPN
resource "panos_globalprotect_portal" "clientless_vpn_example" {
  location = {
    template = {
      name = panos_template.gp_template.name
    }
  }
  name = "gp-portal-clientless-vpn"

  clientless_vpn = {
    hostname = "clientless.example.com"
    inactivity_logout = {
      hours = 2
    }
    login_lifetime = {
      minutes = 90
    }
    max_user      = 1000
    security_zone = panos_zone.gp_zone.name
    crypto_settings = {
      server_cert_verification = {
        block_expired_certificate = true
        block_timeout_cert        = true
        block_unknown_cert        = true
        block_untrusted_issuer    = true
      }
    }
  }
}

# GlobalProtect Portal with Portal Config
resource "panos_globalprotect_portal" "portal_config_example" {
  location = {
    template = {
      name = panos_template.gp_template.name
    }
  }
  name = "gp-portal-config"

  portal_config = {
    certificate_profile = "portal-cert-profile"
    client_auth = [
      {
        name                                    = "client-auth1"
        os                                      = "Any"
        authentication_profile                  = "auth-profile1"
        auto_retrieve_passcode                  = true
        username_label                          = "Username"
        password_label                          = "Password"
        authentication_message                  = "Enter login credentials"
        user_credential_or_client_cert_required = "no"
      }
    ]
    local_address = {
      interface         = panos_ethernet_interface.gp_interface.name
      ip_address_family = "ipv4"
      ip = {
        ipv4 = "10.1.1.1"
      }
    }
    log_fail    = true
    log_success = true
  }
}

# GlobalProtect Portal with Satellite Config
resource "panos_globalprotect_portal" "satellite_config_example" {
  location = {
    template = {
      name = panos_template.gp_template.name
    }
  }
  name = "gp-portal-satellite"

  satellite_config = {
    configs = [
      {
        name        = "satellite-config1"
        devices     = ["device1", "device2"]
        source_user = ["user1", "user2"]
        gateways = [
          {
            name           = "gateway1"
            ipv6_preferred = true
            priority       = 1
            fqdn           = "gateway1.example.com"
          }
        ]
        config_refresh_interval = 24
      }
    ]
    client_certificate = {
      local = {
        certificate_life_time      = 30
        certificate_renewal_period = 7
        issuing_certificate        = "issuing-cert"
        ocsp_responder             = "ocsp-responder"
      }
    }
    root_ca = ["root-ca1", "root-ca2"]
  }
}
