# A multicast MSDP timer routing profile can be imported by providing the following base64 encoded object as the ID
# {
#   location = {
#     template = {
#       name            = "multicast-routing-template"
#       panorama_device = "localhost.localdomain"
#     }
#   }
#
#   name = "msdp-timer-profile-custom"
# }
terraform import panos_multicast_msdp_timer_routing_profile.example $(echo '{"location":{"template":{"name":"multicast-routing-template","panorama_device":"localhost.localdomain"}},"name":"msdp-timer-profile-custom"}' | base64)
