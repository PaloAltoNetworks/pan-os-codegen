resource "panos_local_user" "user1" {
  location = {
    template_vsys = {
      template = panos_template.example.name
    }
  }

  name     = "user1"
  password = "SecurePassword123!"
}

resource "panos_local_user" "user2" {
  location = {
    template_vsys = {
      template = panos_template.example.name
    }
  }

  name     = "user2"
  password = "SecurePassword123!"
}

resource "panos_local_user_group" "example" {
  location = {
    template_vsys = {
      template = panos_template.example.name
    }
  }

  name  = "example-group"
  users = [panos_local_user.user1.name, panos_local_user.user2.name]
}

resource "panos_template" "example" {
  location = {
    panorama = {}
  }

  name = "example-template"
}
