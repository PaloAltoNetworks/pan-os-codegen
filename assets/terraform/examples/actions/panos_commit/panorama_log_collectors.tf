# Panorama commit for log collectors
action "panos_commit" "panorama_log_collectors" {
  config {
    description          = "Commit log collector configuration changes"
    log_collectors       = ["lc-1", "lc-2"]
    log_collector_groups = ["lcg-prod"]
  }
}
