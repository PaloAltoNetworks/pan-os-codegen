# A multicast IGMP interface query routing profile can be imported by providing the following base64 encoded object as the ID
# {
#   location = {
#     template = {
#       name            = "example-template"
#       panorama_device = "localhost.localdomain"
#     }
#   }
#
#   name = "igmp-query-profile1"
# }
terraform import panos_multicast_igmp_interface_query_routing_profile.example $(echo '{"location":{"template":{"name":"example-template","panorama_device":"localhost.localdomain"}},"name":"igmp-query-profile1"}' | base64)
