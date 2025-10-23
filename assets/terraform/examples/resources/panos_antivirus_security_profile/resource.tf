provider "panos" {}

resource "panos_device_group" "example" {
  location = { panorama = {} }
  name     = "my-device-group"
}

resource "panos_antivirus_security_profile" "example" {
  depends_on = [panos_device_group.example]
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }
  name        = "example-profile"
  description = "Example antivirus security profile"
  packet_capture = true

  application_exceptions = [{
    name   = "panos-web-interface"
    action = "alert"
  }]

  decoders = [{
    name            = "http"
    action          = "drop"
    wildfire_action = "alert"
    ml_action       = "reset-client"
  }]

  machine_learning_models = [
    {
      name   = "Windows Executables"
      action = "enable(alert-only)"
    },
    {
      name   = "PowerShell Script 2"
      action = "disable"
    },
    {
      name   = "Executable Linked Format"
      action = "enable"
    }
  ]

  machine_learning_exceptions = [{
    name        = "ml_exception_1"
    filename    = "example.exe"
    description = "Example ML exception"
  }]

  threat_exceptions = [{
    name = "20036500"
  }]
}
