# Panorama partial commit with multiple filters
action "panos_commit" "panorama_partial" {
  config {
    description            = "Partial Panorama commit"
    admins                 = ["admin1"]
    device_groups          = ["prod-dg"]
    templates              = ["base-template"]
    exclude_shared_objects = true
  }
}
