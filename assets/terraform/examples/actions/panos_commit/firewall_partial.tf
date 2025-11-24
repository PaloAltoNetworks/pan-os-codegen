# Firewall partial commit example - commit specific admin changes
action "panos_commit" "firewall_partial" {
  config {
    description = "Commit changes from admin1 and admin2"
    admins      = ["admin1", "admin2"]
  }
}
