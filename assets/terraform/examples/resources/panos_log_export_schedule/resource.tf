resource "panos_template" "example" {
  location = { panorama = {} }
  name     = "example-template"
}

# Example 1: Basic log export schedule with FTP protocol
resource "panos_log_export_schedule" "ftp_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "ftp-export-schedule"
  description = "Export traffic logs via FTP"
  enable      = true
  log_type    = "traffic"
  start_time  = "03:30"

  protocol = {
    ftp = {
      hostname     = "ftp.example.com"
      passive_mode = true
      password     = "secure-password"
      path         = "/logs/export"
      port         = 21
      username     = "ftpuser"
    }
  }
}

# Example 2: Log export schedule with SCP protocol
resource "panos_log_export_schedule" "scp_example" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "scp-export-schedule"
  description = "Export wildfire logs via SCP"
  enable      = true
  log_type    = "wildfire"
  start_time  = "02:00"

  protocol = {
    scp = {
      hostname = "scp.example.com"
      password = "secure-scp-password"
      path     = "/var/logs/wildfire"
      port     = 22
      username = "scpuser"
    }
  }
}

# Example 3: Multiple log export schedules for different log types
resource "panos_log_export_schedule" "threat_export" {
  location = {
    template = {
      name = panos_template.example.name
    }
  }

  name        = "threat-export"
  description = "Daily export of threat logs"
  enable      = true
  log_type    = "threat"
  start_time  = "01:00"

  protocol = {
    ftp = {
      hostname = "logs.example.com"
      username = "loguser"
      password = "logpass"
      path     = "/threat-logs"
      port     = 21
    }
  }
}
