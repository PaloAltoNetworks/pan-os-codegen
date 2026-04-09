# Create a template
resource "panos_template" "ospfv3_template" {
  location = { panorama = {} }
  name     = "ospfv3-routing-template"
}

# OSPFv3 Authentication Profile using AH with SHA-256
resource "panos_ospfv3_auth_routing_profile" "ah_sha256" {
  location = {
    template = {
      name = panos_template.ospfv3_template.name
    }
  }

  name = "ospfv3-ah-sha256-profile"
  spi  = "00000101"

  ah = {
    sha256 = {
      key = "a1b2c3d4-e5f6a7b8-c9d0e1f2-a3b4c5d6-e7f8a9b0-c1d2e3f4-a5b6c7d8-e9f0a1b2"
    }
  }
}

# OSPFv3 Authentication Profile using ESP with authentication and encryption
resource "panos_ospfv3_auth_routing_profile" "esp_full" {
  location = {
    template = {
      name = panos_template.ospfv3_template.name
    }
  }

  name = "ospfv3-esp-secure-profile"
  spi  = "00000201"

  esp = {
    authentication = {
      sha512 = {
        key = "1a2b3c4d-5e6f7a8b-9c0d1e2f-3a4b5c6d-7e8f9a0b-1c2d3e4f-5a6b7c8d-9e0f1a2b-3c4d5e6f-7a8b9c0d-1e2f3a4b-5c6d7e8f-9a0b1c2d-3e4f5a6b-7c8d9e0f-1a2b3c4d"
      }
    }
    encryption = {
      algorithm = "aes-256-cbc"
      key       = "f1e2d3c4-b5a69788-9a0b1c2d-3e4f5a6b-7c8d9e0f-1a2b3c4d-5e6f7a8b-9c0d1e2f"
    }
  }
}

# OSPFv3 Authentication Profile using ESP with encryption only (no authentication)
resource "panos_ospfv3_auth_routing_profile" "esp_encrypt_only" {
  location = {
    template = {
      name = panos_template.ospfv3_template.name
    }
  }

  name = "ospfv3-esp-encrypt-only"
  spi  = "00000301"

  esp = {
    authentication = {
      none = {}
    }
    encryption = {
      algorithm = "aes-128-cbc"
      key       = "a1b2c3d4-e5f6a7b8-c9d0e1f2-a3b4c5d6"
    }
  }
}
