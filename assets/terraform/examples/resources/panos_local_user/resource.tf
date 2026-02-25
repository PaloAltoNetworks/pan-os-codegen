resource "panos_local_user" "example" {
  location = {
    template_vsys = {
      template = panos_template.example.name
    }
  }

  name     = "example-user"
  password = "SecurePassword123!"
  disabled = false
}

resource "panos_template" "example" {
  location = {
    panorama = {}
  }

  name = "example-template"
}
