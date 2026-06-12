# Look up every tag currently registered against an IP address.
data "panos_ip_tag" "example" {
  location = {
    panorama = {}
  }

  ip = "10.0.0.1"
}

# The data source reports all tags present on the IP, regardless of whether they
# are managed by a panos_ip_tag resource.
output "registered_tags" {
  value = data.panos_ip_tag.example.tags
}
