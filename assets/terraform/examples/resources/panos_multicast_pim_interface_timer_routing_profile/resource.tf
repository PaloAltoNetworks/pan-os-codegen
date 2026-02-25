resource "panos_multicast_pim_interface_timer_routing_profile" "example" {
  location = {
    template = {
      name = "my-template"
    }
  }

  name                = "example-pim-timer-profile"
  assert_interval     = 200
  hello_interval      = 60
  join_prune_interval = 120
}
