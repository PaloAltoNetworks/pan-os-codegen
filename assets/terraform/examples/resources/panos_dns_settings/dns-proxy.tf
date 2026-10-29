resource "panos_dns_settings" "split_dns" {
  location = {
    template = {
      name = "firewall-template"
    }
  }

  dns_settings = {
    # Name of an existing shared DNS proxy in this template.
    dns_proxy_object = "existing-shared-dns-proxy"
  }
}
