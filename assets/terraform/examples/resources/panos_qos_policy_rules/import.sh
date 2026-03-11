# A set of QoS rules can be imported by providing the following base64 encoded object as the ID
# {
#     location = {
#         device_group = {
#         name = "example-device-group"
#         rulebase = "pre-rulebase"
#         panorama_device = "localhost.localdomain"
#         }
#     }
#
#     position = { where = "after", directly = true, pivot = "existing-rule" }
#
#     names = [
#         "qos-rule-8",
#         "qos-rule-9"
#     ]
# }
terraform import panos_qos_policy_rules.example $(echo '{"location":{"device_group":{"name":"example-device-group","panorama_device":"localhost.localdomain","rulebase":"pre-rulebase"}},"names":["qos-rule-8","qos-rule-9"],"position":{"directly":true,"pivot":"existing-rule","where":"after"}}' | base64)
