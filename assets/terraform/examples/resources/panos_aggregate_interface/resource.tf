resource "panos_aggregate_interface" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "ae1"

  comment = "aggregate interface comment"
  ha = {
    lacp = {
      enable            = true
      fast_failover     = true
      max_ports         = 4
      mode              = "active"
      system_priority   = 10
      transmission_rate = "fast"
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
