resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-tmpl"
}

resource "panos_ldap_profile" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "example-ldap-profile"
}

# Authentication Profile with LDAP method
resource "panos_authentication_profile" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "example-auth-profile"

  allow_list = ["all"]

  lockout = {
    failed_attempts = 5
    lockout_time    = 30
  }

  method = {
    ldap = {
      login_attribute = "sAMAccountName"
      passwd_exp_days = 14
      server_profile  = panos_ldap_profile.example.name
    }
  }
}
