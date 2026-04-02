# Create a template for the OSPF routing configuration
resource "panos_template" "ospf_template" {
  location = { panorama = {} }
  name     = "ospf-routing-template"
}

# OSPF Authentication Profile with MD5 authentication using multiple keys
# This allows for key rotation - the preferred key is used for sending packets
# while all keys can validate incoming packets
resource "panos_ospf_auth_routing_profile" "md5_auth" {
  location = {
    template = {
      name = panos_template.ospf_template.name
    }
  }

  name = "ospf-md5-auth"

  md5 = [
    {
      name      = "key-1"
      key       = "SecureKey123456"
      preferred = true
    },
    {
      name = "key-2"
      key  = "BackupKey987654"
    }
  ]
}
