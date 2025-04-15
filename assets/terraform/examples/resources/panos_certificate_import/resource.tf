resource "panos_certificate_import" "example1" {
  location = { template = { name = panos_template.example.name } }

  local = {
    pem = {
      certificate = file("cert.pem")     # PEM-encoded certificate
      private_key = file("cert.key")     # PEM-encoded private key
      passphrase  = "example-passphrase" # passphrase used to decrypt private key
    }
  }
}

resource "panos_template" "example2" {
  location = { panorama = {} }

  name = "example-template"

  local = {
    pcks12 = {
      certificate = base64encode(file("cert.pkcs12"))
      passphrase  = "example-passphrase"
    }
  }
}
