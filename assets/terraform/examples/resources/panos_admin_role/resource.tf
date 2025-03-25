resource "panos_admin_role" "example" {
  location = { template = { name = panos_template.example.name } }

  name = "example-admin-role"

  description = "admin role description"

  role = {
    device = {
      cli = "superuser"
      restapi = {
        device = {
          email_server_profiles     = "read-only"
          http_server_profiles      = "read-only"
          ldap_server_profiles      = "read-only"
          log_interface_setting     = "read-only"
          snmp_trap_server_profiles = "read-only"
          syslog_server_profiles    = "read-only"
          virtual_systems           = "read-only"
        }
        network = {
          aggregate_ethernet_interfaces               = "read-only"
          bfd_network_profiles                        = "read-only"
          bgp_routing_profiles                        = "read-only"
          dhcp_relays                                 = "read-only"
          dhcp_servers                                = "read-only"
          dns_proxies                                 = "read-only"
          ethernet_interfaces                         = "read-only"
          globalprotect_clientless_app_groups         = "read-only"
          globalprotect_clientless_apps               = "read-only"
          globalprotect_gateways                      = "read-only"
          globalprotect_ipsec_crypto_network_profiles = "read-only"
          globalprotect_mdm_servers                   = "read-only"
          globalprotect_portals                       = "read-only"
          gre_tunnels                                 = "read-only"
          ike_crypto_network_profiles                 = "read-only"
          ike_gateway_network_profiles                = "read-only"
          interface_management_network_profiles       = "read-only"
          ipsec_crypto_network_profiles               = "read-only"
          ipsec_tunnels                               = "read-only"
          lldp                                        = "read-only"
          lldp_network_profiles                       = "read-only"
          logical_routers                             = "read-only"
          loopback_interfaces                         = "read-only"
          qos_interfaces                              = "read-only"
          qos_network_profiles                        = "read-only"
          sdwan_interface_profiles                    = "read-only"
          sdwan_interfaces                            = "read-only"
          tunnel_interfaces                           = "read-only"
          tunnel_monitor_network_profiles             = "read-only"
          virtual_routers                             = "read-only"
          virtual_wires                               = "read-only"
          vlan_interfaces                             = "read-only"
          vlans                                       = "read-only"
          zone_protection_network_profiles            = "read-only"
          zones                                       = "read-only"
        }
        objects = {
          address_groups                             = "read-only"
          addresses                                  = "read-only"
          anti_spyware_security_profiles             = "read-only"
          antivirus_security_profiles                = "read-only"
          application_filters                        = "read-only"
          application_groups                         = "read-only"
          applications                               = "read-only"
          authentication_enforcements                = "read-only"
          custom_data_patterns                       = "read-only"
          custom_spyware_signatures                  = "read-only"
          custom_url_categories                      = "read-only"
          custom_vulnerability_signatures            = "read-only"
          data_filtering_security_profiles           = "read-only"
          decryption_profiles                        = "read-only"
          devices                                    = "read-only"
          dos_protection_security_profiles           = "read-only"
          dynamic_user_groups                        = "read-only"
          external_dynamic_lists                     = "read-only"
          file_blocking_security_profiles            = "read-only"
          globalprotect_hip_objects                  = "read-only"
          globalprotect_hip_profiles                 = "read-only"
          gtp_protection_security_profiles           = "read-only"
          log_forwarding_profiles                    = "read-only"
          packet_broker_profiles                     = "read-only"
          regions                                    = "read-only"
          schedules                                  = "read-only"
          sctp_protection_security_profiles          = "read-only"
          sdwan_error_correction_profiles            = "read-only"
          sdwan_path_quality_profiles                = "read-only"
          sdwan_saas_quality_profiles                = "read-only"
          sdwan_traffic_distribution_profiles        = "read-only"
          security_profile_groups                    = "read-only"
          service_groups                             = "read-only"
          services                                   = "read-only"
          tags                                       = "read-only"
          url_filtering_security_profiles            = "read-only"
          vulnerability_protection_security_profiles = "read-only"
          wildfire_analysis_security_profiles        = "read-only"
        }
        policies = {
          application_override_rules    = "read-only"
          authentication_rules          = "read-only"
          decryption_rules              = "read-only"
          dos_rules                     = "read-only"
          nat_rules                     = "read-only"
          network_packet_broker_rules   = "read-only"
          policy_based_forwarding_rules = "read-only"
          qos_rules                     = "read-only"
          sdwan_rules                   = "read-only"
          security_rules                = "read-only"
          tunnel_inspection_rules       = "read-only"
        }
        system = {
          configuration = "read-only"
        }
      }
      webui = {
        acc       = "enable"
        dashboard = "enable"
        tasks     = "enable"
        validate  = "enable"
        commit = {
          commit_for_other_admins = "enable"
          device                  = "disable"
          object_level_changes    = "enable"
        }
        device = {
          access_domain           = "read-only"
          admin_roles             = "read-only"
          administrators          = "read-only"
          authentication_profile  = "read-only"
          authentication_sequence = "read-only"
          block_pages             = "read-only"
          config_audit            = "enable"
          data_redistribution     = "read-only"
          device_quarantine       = "read-only"
          # dhcp_syslog_server = "read-only"
          dynamic_updates       = "read-only"
          global_protect_client = "read-only"
          high_availability     = "read-only"
          licenses              = "read-only"
          log_fwd_card          = "read-only"
          master_key            = "read-only"
          plugins               = "disable"
          scheduled_log_export  = "enable"
          shared_gateways       = "read-only"
          software              = "read-only"
          support               = "read-only"
          troubleshooting       = "read-only"
          user_identification   = "read-only"
          virtual_systems       = "read-only"
          vm_info_source        = "read-only"
          certificate_management = {
            certificate_profile      = "read-only"
            certificates             = "read-only"
            ocsp_responder           = "read-only"
            scep                     = "read-only"
            ssh_service_profile      = "read-only"
            ssl_decryption_exclusion = "read-only"
            ssl_tls_service_profile  = "read-only"
          }
          local_user_database = {
            user_groups = "read-only"
            users       = "read-only"
          }
          log_settings = {
            cc_alarm      = "read-only"
            config        = "read-only"
            correlation   = "read-only"
            globalprotect = "read-only"
            hipmatch      = "read-only"
            iptag         = "read-only"
            manage_log    = "read-only"
            system        = "read-only"
            user_id       = "read-only"
          }
          policy_recommendations = {
            iot  = "read-only"
            saas = "read-only"
          }
          server_profile = {
            dns       = "read-only"
            email     = "read-only"
            http      = "read-only"
            kerberos  = "read-only"
            ldap      = "read-only"
            mfa       = "read-only"
            netflow   = "read-only"
            radius    = "read-only"
            saml_idp  = "read-only"
            scp       = "read-only"
            snmp_trap = "read-only"
            syslog    = "read-only"
            tacplus   = "read-only"
          }
          setup = {
            content_id = "read-only"
            hsm        = "read-only"
            interfaces = "read-only"
            management = "read-only"
            operations = "read-only"
            services   = "read-only"
            session    = "read-only"
            telemetry  = "read-only"
            wildfire   = "read-only"
          }
        }
        global = {
          system_alarms = "disable"
        }
        monitor = {
          app_scope             = "disable"
          application_reports   = "disable"
          block_ip_list         = "disable"
          botnet                = "disable"
          external_logs         = "disable"
          gtp_reports           = "disable"
          packet_capture        = "disable"
          sctp_reports          = "disable"
          session_browser       = "disable"
          threat_reports        = "disable"
          traffic_reports       = "disable"
          url_filtering_reports = "disable"
          view_custom_reports   = "disable"
          automated_correlation_engine = {
            correlated_events   = "disable"
            correlation_objects = "disable"
          }
          custom_reports = {
            application_statistics = "disable"
            auth                   = "disable"
            data_filtering_log     = "disable"
            decryption_log         = "disable"
            decryption_summary     = "disable"
            globalprotect          = "disable"
            gtp_log                = "disable"
            gtp_summary            = "disable"
            hipmatch               = "disable"
            iptag                  = "disable"
            sctp_log               = "disable"
            sctp_summary           = "disable"
            threat_log             = "disable"
            threat_summary         = "disable"
            traffic_log            = "disable"
            traffic_summary        = "disable"
            tunnel_log             = "disable"
            tunnel_summary         = "disable"
            url_log                = "disable"
            url_summary            = "disable"
            userid                 = "disable"
            wildfire_log           = "disable"
          }
          logs = {
            alarm          = "disable"
            authentication = "disable"
            configuration  = "disable"
            data_filtering = "disable"
            decryption     = "disable"
            globalprotect  = "disable"
            gtp            = "disable"
            hipmatch       = "disable"
            iptag          = "disable"
            sctp           = "disable"
            system         = "disable"
            threat         = "disable"
            traffic        = "disable"
            tunnel         = "disable"
            url            = "disable"
            userid         = "disable"
            wildfire       = "disable"
          }
          pdf_reports = {
            email_scheduler               = "disable"
            manage_pdf_summary            = "disable"
            pdf_summary_reports           = "disable"
            report_groups                 = "disable"
            saas_application_usage_report = "disable"
            user_activity_report          = "disable"
          }
        }
        network = {
          dhcp                    = "disable"
          dns_proxy               = "disable"
          gre_tunnels             = "disable"
          interfaces              = "disable"
          ipsec_tunnels           = "disable"
          lldp                    = "disable"
          qos                     = "disable"
          sdwan_interface_profile = "disable"
          secure_web_gateway      = "disable"
          virtual_routers         = "disable"
          virtual_wires           = "disable"
          vlans                   = "disable"
          zones                   = "disable"
          global_protect = {
            clientless_app_groups = "disable"
            clientless_apps       = "disable"
            gateways              = "disable"
            mdm                   = "disable"
            portals               = "disable"
          }
          network_profiles = {
            bfd_profile         = "disable"
            gp_app_ipsec_crypto = "disable"
            ike_crypto          = "disable"
            ike_gateways        = "disable"
            interface_mgmt      = "disable"
            ipsec_crypto        = "disable"
            lldp_profile        = "disable"
            qos_profile         = "disable"
            tunnel_monitor      = "disable"
            zone_protection     = "disable"
          }
          routing = {
            logical_routers = "disable"
            routing_profiles = {
              bfd       = "disable"
              bgp       = "disable"
              filters   = "disable"
              multicast = "disable"
              ospf      = "disable"
              ospfv3    = "disable"
              ripv2     = "disable"
            }
          }
        }
        objects = {
          address_groups          = "disable"
          addresses               = "disable"
          application_filters     = "disable"
          application_groups      = "disable"
          applications            = "disable"
          authentication          = "disable"
          devices                 = "disable"
          dynamic_block_lists     = "disable"
          dynamic_user_groups     = "disable"
          log_forwarding          = "disable"
          packet_broker_profile   = "disable"
          regions                 = "disable"
          schedules               = "disable"
          security_profile_groups = "disable"
          service_groups          = "disable"
          services                = "disable"
          tags                    = "disable"
          custom_objects = {
            data_patterns = "disable"
            spyware       = "disable"
            url_category  = "disable"
            vulnerability = "disable"
          }
          decryption = {
            decryption_profile = "disable"
          }
          global_protect = {
            hip_objects  = "disable"
            hip_profiles = "disable"
          }
          sdwan = {
            sdwan_dist_profile             = "disable"
            sdwan_error_correction_profile = "disable"
            sdwan_profile                  = "disable"
            sdwan_saas_quality_profile     = "disable"
          }
          security_profiles = {
            anti_spyware             = "disable"
            antivirus                = "disable"
            data_filtering           = "disable"
            dos_protection           = "disable"
            file_blocking            = "disable"
            gtp_protection           = "disable"
            sctp_protection          = "disable"
            url_filtering            = "disable"
            vulnerability_protection = "disable"
            wildfire_analysis        = "disable"
          }
        }
        operations = {
          download_core_files        = "disable"
          download_pcap_files        = "disable"
          generate_stats_dump_file   = "disable"
          generate_tech_support_file = "disable"
          reboot                     = "disable"
        }
        policies = {
          application_override_rulebase  = "disable"
          authentication_rulebase        = "disable"
          dos_rulebase                   = "disable"
          nat_rulebase                   = "disable"
          network_packet_broker_rulebase = "disable"
          pbf_rulebase                   = "disable"
          qos_rulebase                   = "disable"
          rule_hit_count_reset           = "disable"
          sdwan_rulebase                 = "disable"
          security_rulebase              = "disable"
          ssl_decryption_rulebase        = "disable"
          tunnel_inspect_rulebase        = "disable"
        }
        privacy = {
          show_full_ip_addresses              = "disable"
          show_user_names_in_logs_and_reports = "disable"
          view_pcap_files                     = "disable"
        }
        save = {
          object_level_changes  = "disable"
          partial_save          = "disable"
          save_for_other_admins = "disable"
        }
      }
      xmlapi = {
        commit  = "disable"
        config  = "disable"
        export  = "disable"
        import  = "disable"
        iot     = "disable"
        log     = "disable"
        op      = "disable"
        report  = "disable"
        user_id = "disable"
      }
    }
  }
}

resource "panos_template" "example" {
  location = { panorama = {} }

  name = "example-template"
}
