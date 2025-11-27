# Push device group configuration to specific devices
action "panos_push_to_devices" "device_group_specific" {
  config {
    type        = "device_group"
    name        = "prod-dg-1"
    description = "Push to specific firewalls"
    devices     = ["007951000012345", "007951000012346"]
  }
}
