# Firewall force commit example
action "panos_commit" "firewall_force" {
  config {
    description = "Force commit all changes"
    force       = true
  }
}
