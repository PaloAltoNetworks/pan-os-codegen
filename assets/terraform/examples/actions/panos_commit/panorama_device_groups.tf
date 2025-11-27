# Panorama commit for specific device groups
action "panos_commit" "panorama_device_groups" {
  config {
    description   = "Commit changes to production device groups"
    device_groups = ["prod-dg-1", "prod-dg-2"]
  }
}
