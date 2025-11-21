# A multicast MSDP authentication routing profile can be imported by providing the following base64 encoded object as the ID
# {
#   location = {
#     template = {
#       name            = "multicast-routing-template"
#       panorama_device = "localhost.localdomain"
#     }
#   }
#
#   name = "msdp-auth-profile"
# }
terraform import panos_multicast_msdp_authentication_routing_profile.example $(echo '{"location":{"template":{"name":"multicast-routing-template","panorama_device":"localhost.localdomain"}},"name":"msdp-auth-profile"}' | base64)
