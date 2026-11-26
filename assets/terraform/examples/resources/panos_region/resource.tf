# Tag a classroom subnet with a geographic region so it plots correctly
# on PAN-OS's ACC traffic map, on a firewall's virtual system.
resource "panos_region" "classroom" {
  location = {
    vsys = {
      name = "vsys1"
    }
  }

  name      = "CR_BanyanC"
  address   = ["10.220.34.0/24"]
  latitude  = 1.283333
  longitude = 103.833333
}

# Region objects can also be defined in Panorama's shared scope (a provider
# targeting Panorama, not a standalone NGFW -- shared scope has no region
# node on a standalone firewall's own local config tree).
resource "panos_region" "shared" {
  location = {
    shared = {}
  }

  name      = "hq_region"
  address   = ["192.168.0.0/16"]
  latitude  = 37.774929
  longitude = -122.419418
}
