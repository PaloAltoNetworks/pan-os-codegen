# Register tags on an IP directly on a firewall's virtual system.
resource "panos_ip_tag" "vsys" {
  location = {
    vsys = {
      name = "vsys1"
    }
  }

  ip   = "10.0.0.1"
  tags = ["web", "prod"]
}

# Register tags directly on Panorama's own User-ID table (no target firewall).
resource "panos_ip_tag" "panorama" {
  location = {
    panorama = {}
  }

  ip   = "10.0.0.2"
  tags = ["db"]
}

# Register tags on a Panorama-managed firewall, targeted by serial number.
resource "panos_ip_tag" "target_device" {
  location = {
    target_device = {
      serial = "0123456789"
      vsys   = "vsys1"
    }
  }

  ip   = "10.0.0.3"
  tags = ["dmz"]
}
