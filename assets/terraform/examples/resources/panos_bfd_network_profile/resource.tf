resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "bfd-network-template"
}

# BFD Network Profile with active mode and custom timing parameters.
# Suitable for single-hop BFD sessions where fast failure detection is needed.
resource "panos_bfd_network_profile" "active" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "bfd-active-fast"

  # Use active mode so this device initiates BFD control packets
  mode = "active"

  # Increase the detection multiplier to allow 5 missed packets before
  # declaring the session down (default is 3)
  detection_multiplier = 5

  # Require BFD control packets every 300 ms (default is 1000 ms)
  min_rx_interval = 300
  min_tx_interval = 300

  # Hold BFD sessions for 500 ms after a link event before re-negotiating
  hold_time = 500
}

# BFD Network Profile configured for multihop sessions in passive mode.
# Multihop BFD traverses multiple IP hops (e.g., across a routed WAN) so
# a minimum accepted TTL guards against spoofed packets from closer peers.
resource "panos_bfd_network_profile" "multihop_passive" {
  depends_on = [panos_template.example]
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "bfd-multihop-passive"

  # Passive mode: wait for the remote peer to initiate BFD control packets
  mode = "passive"

  detection_multiplier = 3
  min_rx_interval      = 500
  min_tx_interval      = 500

  # Enable multihop and set the minimum TTL accepted on incoming BFD packets.
  # A value of 200 rejects packets that have traversed more than 55 hops,
  # limiting the reach of potential spoofed BFD packets.
  multihop = {
    min_received_ttl = 200
  }
}
