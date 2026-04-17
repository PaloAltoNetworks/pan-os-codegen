resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

# Email server profile forwarding security logs to a corporate SMTP relay
# with authenticated SMTP and custom log format strings.
resource "panos_email_server_profile" "example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name = "security-alerts-email"

  servers = [
    {
      name         = "corporate-smtp"
      display_name = "Corporate SMTP Relay"
      from         = "panos-alerts@corp.example.com"
      to           = "security-team@corp.example.com"
      and_also_to  = "noc@corp.example.com"
      gateway      = "smtp.corp.example.com"
      protocol     = "SMTP"
      port         = 587

      # Use Login authentication with SMTP credentials
      authentication_type = "Login"
      username            = "panos-svc"
      password            = "Str0ngP@ssw0rd!"
    }
  ]

  # Custom format strings for log types relevant to security operations.
  # These override the default PAN-OS email log format.
  format = {
    traffic = "$receive_time,$serial,$type,$subtype,$src,$dst,$proto,$action"
    threat  = "$receive_time,$serial,$type,$subtype,$src,$dst,$threat_name,$severity"
    system  = "$receive_time,$serial,$type,$subtype,$severity,$opaque"
    url     = "$receive_time,$serial,$type,$subtype,$src,$dst,$url"
    wildfire = "$receive_time,$serial,$type,$subtype,$src,$dst,$threat_name,$filetype"

    # Escape backslashes and double-quotes in log field values
    escaping = {
      escape_character   = "\\"
      escaped_characters = "\""
    }
  }
}
