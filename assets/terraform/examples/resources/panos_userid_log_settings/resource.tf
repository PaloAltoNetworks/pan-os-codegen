resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

resource "panos_userid_log_settings" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name             = "example-userid-settings"
  description      = "userid log settings example"
  filter           = "(severity eq high)"
  send_to_panorama = true
  quarantine       = false

  actions = [
    {
      name = "tag-action"
      type = {
        tagging = {
          action = "add-tag"
          target = "source-address"
          tags   = ["tag1", "tag2"]
        }
      }
    }
  ]
}
