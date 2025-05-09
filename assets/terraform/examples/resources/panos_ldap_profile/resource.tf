resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-tmpl"
}

resource "panos_ldap_profile" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "LDAP"

  base                      = "dc=example,dc=com"
  bind_dn                   = "cn=admin,dc=example,dc=com"
  bind_password             = "admin_password"
  bind_timelimit            = 30
  disabled                  = false
  ldap_type                 = "active-directory"
  retry_interval            = 60
  ssl                       = true
  timelimit                 = 30
  verify_server_certificate = true

  servers = [
    {
      name    = "ADSRV1"
      address = "ldap.example.com"
      port    = 389
    }
  ]
}
