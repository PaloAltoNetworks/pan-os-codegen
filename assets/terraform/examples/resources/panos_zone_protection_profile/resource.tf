resource "panos_zone_protection_profile" "example" {
  location = { ngfw = {} }
  name = "zpp1"
  description = "test description"
  asymmetric_path = "bypass"
  discard_icmp_error = true
  discard_icmp_frag = true
  discard_icmp_large_packet = true
  discard_icmp_ping_zero_id = true
  discard_ip_frag = true
  discard_ip_spoof = true
  discard_loose_source_routing = true
  discard_malformed_option = true
  discard_overlapping_tcp_segment_mismatch = true
  discard_record_route = true
  discard_security = true
  discard_stream_id = true
  discard_strict_source_routing = true
  discard_tcp_split_handshake = true
  discard_tcp_syn_with_data = true
  discard_tcp_synack_with_data = true
  discard_timestamp = true
  discard_unknown_option = true
  flood = {
    icmp = {
      enable = true
      red = {
        activate_rate = 100
        alarm_rate    = 200
        maximal_rate  = 300
      }
    }
    icmpv6 = {
      enable = true
      red = {
        activate_rate = 100
        alarm_rate    = 200
        maximal_rate  = 300
      }
    }
    other_ip = {
      enable = true
      red = {
        activate_rate = 100
        alarm_rate    = 200
        maximal_rate  = 300
      }
    }
    tcp_syn = {
      enable = true
      syn_cookies = {
        activate_rate = 100
        alarm_rate    = 200
        maximal_rate  = 300
      }
    }
    udp = {
      enable = true
      red = {
        activate_rate = 100
        alarm_rate    = 200
        maximal_rate  = 300
      }
    }
  }
}
