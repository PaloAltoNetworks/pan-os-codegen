resource "panos_wildfire_analysis_security_profile" "example" {
  location = {
    device_group = {
      name = "shared"
    }
  }
  name             = "example-profile"
  description      = "This is an example profile."
  disable_override = "no"
  rules = [
    {
      name        = "default-rule"
      application = ["any"]
      file_type   = ["any"]
      direction   = "both"
      analysis    = "public-cloud"
    }
  ]
}
