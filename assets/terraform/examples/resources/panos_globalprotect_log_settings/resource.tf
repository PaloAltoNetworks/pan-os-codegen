resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

resource "panos_globalprotect_log_settings" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name             = "example-gp-settings"
  description      = "globalprotect log settings example"
  filter           = "(severity eq high)"
  send_to_panorama = true

  actions = [
    {
      name = "tag-action"
      type = {
        tagging = {
          action = "add-tag"
          target = "source-address"
          tags   = ["tag1", "tag2"]
          registration = {
            panorama = {}
          }
        }
      }
    }
  ]
}
