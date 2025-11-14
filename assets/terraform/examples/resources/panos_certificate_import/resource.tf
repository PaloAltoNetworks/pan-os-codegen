resource "panos_certificate_import" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "EXAMPLE-CERT"

  local = {
    pem = {
      certificate = file("cert.pem") # PEM-encoded certificate
      private_key = file("cert.key") # PEM-encoded private key
      passphrase  = "paloalto"       # passphrase used to decrypt private key
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
