# Push device group with template configuration
action "panos_push_to_devices" "device_group_with_template" {
  config {
    type             = "device_group"
    name             = "prod-dg-1"
    description      = "Push DG and template config"
    include_template = true
  }
}
