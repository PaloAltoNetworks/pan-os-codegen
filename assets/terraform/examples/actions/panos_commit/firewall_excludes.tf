# Firewall commit with exclusions (NGFW only)
action "panos_commit" "firewall_excludes" {
  config {
    description                = "Commit policy changes only"
    exclude_device_and_network = true
    exclude_shared_objects     = true
    exclude_policy_and_objects = false
  }
}
