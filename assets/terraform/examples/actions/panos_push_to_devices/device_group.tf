# Push device group configuration to all managed devices
action "panos_push_to_devices" "device_group" {
  config {
    type        = "device_group"
    name        = "prod-dg-1"
    description = "Push device group configuration to all devices"
  }
}
