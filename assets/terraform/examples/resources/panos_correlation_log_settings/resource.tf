resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

resource "panos_correlation_log_settings" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "example-correlation-settings"
  description = "correlation log settings example"
  filter      = "(severity eq high)"
  quarantine  = false

  actions = [
    {
      name = "integration-action"
      type = {
        integration = {
          action = "Azure-Security-Center-Integration"
        }
      }
    }
  ]
}
