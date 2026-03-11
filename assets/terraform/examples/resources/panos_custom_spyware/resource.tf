resource "panos_custom_spyware" "example" {
  name = 6900001
  threatname = "my-custom-spyware"
  severity = "critical"
  location = {
    device_group = {
      name = "my-device-group"
    }
  }
}
