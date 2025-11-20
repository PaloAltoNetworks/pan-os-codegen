resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

# Example SSL/TLS Service Profile with just a certificate
resource "panos_ssl_tls_service_profile" "basic" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "basic-ssl-profile"
  certificate = "my-certificate"
}

# Example SSL/TLS Service Profile with protocol settings
resource "panos_ssl_tls_service_profile" "advanced" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "advanced-ssl-profile"
  certificate = "my-certificate"

  protocol_settings = {
    # Encryption algorithms
    allow_algorithm_aes_128_cbc = true
    allow_algorithm_aes_128_gcm = true
    allow_algorithm_aes_256_cbc = true
    allow_algorithm_aes_256_gcm = true

    # Key exchange algorithms
    allow_algorithm_dhe   = true
    allow_algorithm_ecdhe = true
    allow_algorithm_rsa   = true

    # Authentication algorithms
    allow_authentication_sha1   = true
    allow_authentication_sha256 = true
    allow_authentication_sha384 = true

    # TLS version constraints
    min_version = "tls1-1"
    max_version = "tls1-2"
  }
}

# Example SSL/TLS Service Profile with minimal secure configuration
resource "panos_ssl_tls_service_profile" "secure" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "secure-ssl-profile"
  certificate = "my-certificate"

  protocol_settings = {
    # Modern encryption only
    allow_algorithm_aes_128_gcm = true
    allow_algorithm_aes_256_gcm = true

    # Modern key exchange
    allow_algorithm_ecdhe = true

    # Strong authentication
    allow_authentication_sha256 = true
    allow_authentication_sha384 = true

    # TLS 1.2 only
    min_version = "tls1-2"
    max_version = "tls1-2"
  }
}
