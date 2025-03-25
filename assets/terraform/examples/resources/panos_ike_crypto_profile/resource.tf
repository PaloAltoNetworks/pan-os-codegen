resource "panos_ike_crypto_profile" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "example-ike-profile"

  authentication_multiple = 50
  dh_group                = ["group1", "group2"]
  encryption              = ["3des", "aes-256-gcm"]
  hash                    = ["md5", "sha256"]
  lifetime = {
    seconds = 3600
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
