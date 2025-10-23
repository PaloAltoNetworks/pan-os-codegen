resource "panos_file_blocking_security_profile" "example" {
    location = var.device_group
    name = "my-profile"
    description = "test description"
    disable_override = "yes"
    rules = [
        {
            name = "rule1"
            applications = ["any"]
            file_types = ["any"]
            direction = "both"
            action = "block"
        }
    ]
}

variable "device_group" {
    description = "The device group location for the profile."
    type = any
    default = {
        device_group = {
            name = "my-device-group"
        }
    }
}