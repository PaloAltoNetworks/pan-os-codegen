resource "panos_password_profile" "example" {
  location = { template = { name = "my-template" } }

  name = "example-password-profile"

  password_change = {
    expiration_period                  = 90
    expiration_warning_period          = 7
    post_expiration_admin_login_count  = 3
    post_expiration_grace_period       = 5
  }
}
