# Basic firewall commit example
action "panos_commit" "firewall_basic" {
  config {
    description = "Commit all pending changes"
  }
}
