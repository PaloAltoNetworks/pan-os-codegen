resource "panos_multicast_igmp_interface_query_routing_profile" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  for_each = tomap({
    "igmp-query-profile1" = {
      description                = "IGMP query profile with immediate leave"
      immediate_leave            = true
      last_member_query_interval = 5
      max_query_response_time    = 15
      query_interval             = 200
    }
    "igmp-query-profile2" = {
      description                = "IGMP query profile with default values"
      immediate_leave            = false
      last_member_query_interval = 1
      max_query_response_time    = 10
      query_interval             = 125
    }
  })

  name                       = each.key
  immediate_leave            = each.value.immediate_leave
  last_member_query_interval = each.value.last_member_query_interval
  max_query_response_time    = each.value.max_query_response_time
  query_interval             = each.value.query_interval
}

resource "panos_template" "example" {
  location = {
    panorama = {}
  }

  name = "example-template"
}
