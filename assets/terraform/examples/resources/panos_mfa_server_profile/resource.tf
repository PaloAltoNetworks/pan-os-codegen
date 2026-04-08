resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "mfa-template"
}

# Certificate profile for MFA server validation
resource "panos_certificate_profile" "mfa_ca" {
  location = { template = { name = panos_template.example.name } }
  name     = "mfa-server-ca"
}

# Okta Adaptive MFA Profile
# Okta provides adaptive MFA based on context and risk
resource "panos_mfa_server_profile" "okta" {
  location = { template = { name = panos_template.example.name } }

  name = "okta-mfa-profile"

  mfa_cert_profile = panos_certificate_profile.mfa_ca.name
  mfa_vendor_type  = "okta-adaptive-v1"

  mfa_config = [
    {
      name  = "okta-api-host"
      value = "company.okta.com"
    },
    {
      name  = "okta-baseuri"
      value = "/api/v1"
    },
    {
      name  = "okta-token"
      value = "00A1bCdEfGhIjKlMnOpQrStUvWxYz2345678" # Use a variable in production
    },
    {
      name  = "okta-org"
      value = "company"
    },
    {
      name  = "okta-timeout"
      value = "30"
    }
  ]
}

# Duo Security MFA Profile
# Duo provides push notifications, SMS, and phone call verification
resource "panos_mfa_server_profile" "duo" {
  location = { template = { name = panos_template.example.name } }

  name = "duo-mfa-profile"

  mfa_cert_profile = panos_certificate_profile.mfa_ca.name
  mfa_vendor_type  = "duo-security-v2"

  mfa_config = [
    {
      name  = "duo-api-host"
      value = "api-a1b2c3d4.duosecurity.com"
    },
    {
      name  = "duo-integration-key"
      value = "DIXXXXXXXXXXXXXXXXXX" # Use a variable in production
    },
    {
      name  = "duo-secret-key"
      value = "deadbeefcafebabe0123456789abcdef01234567" # Use a variable in production
    },
    {
      name  = "duo-timeout"
      value = "30"
    },
    {
      name  = "duo-baseuri"
      value = "https://api-a1b2c3d4.duosecurity.com"
    }
  ]
}

# PingIdentity MFA Profile
# PingIdentity provides enterprise-grade adaptive authentication
resource "panos_mfa_server_profile" "ping" {
  location = { template = { name = panos_template.example.name } }

  name = "ping-mfa-profile"

  mfa_cert_profile = panos_certificate_profile.mfa_ca.name
  mfa_vendor_type  = "ping-identity-v1"

  mfa_config = [
    {
      name  = "ping-api-host"
      value = "idpxnyl3m.pingidentity.com"
    },
    {
      name  = "ping-baseuri"
      value = "https://tenant.pingone.com"
    },
    {
      name  = "ping-token"
      value = "AbCdEfGhIjKlMnOpQrStUvWxYz0123456789" # Use a variable in production
    },
    {
      name  = "ping-org-alias"
      value = "12345678-1234-1234-1234-123456789abc"
    },
    {
      name  = "ping-timeout"
      value = "30"
    }
  ]
}

# RSA SecurID Access MFA Profile
# RSA provides hardware token and software token based authentication
resource "panos_mfa_server_profile" "rsa" {
  location = { template = { name = panos_template.example.name } }

  name = "rsa-mfa-profile"

  mfa_cert_profile = panos_certificate_profile.mfa_ca.name
  mfa_vendor_type  = "rsa-securid-access-v1"

  mfa_config = [
    {
      name  = "rsa-api-host"
      value = "cloud.securid.com"
    },
    {
      name  = "rsa-baseuri"
      value = "https://tenant.rsa.com"
    },
    {
      name  = "rsa-accesskey"
      value = "abcdef1234567890ABCDEF1234567890" # Use a variable in production
    },
    {
      name  = "rsa-accessid"
      value = "RSAID123456"
    },
    {
      name  = "rsa-assurancepolicyid"
      value = "policy-abc-123-def-456"
    },
    {
      name  = "rsa-timeout"
      value = "90"
    }
  ]
}
