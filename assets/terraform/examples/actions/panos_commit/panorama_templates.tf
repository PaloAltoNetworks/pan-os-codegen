# Panorama commit for specific templates
action "panos_commit" "panorama_templates" {
  config {
    description = "Commit template changes"
    templates   = ["network-template", "security-template"]
  }
}
