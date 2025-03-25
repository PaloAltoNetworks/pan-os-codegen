resource "panos_dynamic_updates" "example" {
  location = { template = { name = panos_template.example.name } }

  update_schedule = {
    anti_virus = {
      recurring = {
        sync_to_peer = true
        threshold    = 10
        daily        = { action = "download-and-install", at = "20:10" }
      }
    }

    app_profile = {
      recurring = {
        sync_to_peer = true
        threshold    = 10
        daily        = { action = "download-and-install", at = "20:10" }
      }
    }

    global_protect_clientless_vpn = {
      recurring = {
        daily = { action = "download-and-install", at = "20:10" }
      }
    }

    global_protect_datafile = {
      recurring = {
        daily = { action = "download-and-install", at = "20:10" }
      }
    }

    statistics_service = {
      url_reports                 = true
      application_reports         = false
      file_identification_reports = true
      health_performance_reports  = true
      passive_dns_monitoring      = true

      threat_prevention_information = true
      threat_prevention_pcap        = true
      threat_prevention_reports     = true
    }

    threats = {
      recurring = {
        sync_to_peer      = true
        threshold         = 10
        new_app_threshold = 10

        daily = {
          disable_new_content = true
          action              = "download-and-install"
          at                  = "20:10"
        }
      }
    }

    wf_private = {
      recurring = {
        sync_to_peer  = true
        every_15_mins = { action = "download-and-install", at = 10 }
      }
    }

    wildfire = {
      recurring = {
        every_15_mins = {
          sync_to_peer = true
          action       = "download-and-install"
          at           = 10
        }
      }
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
