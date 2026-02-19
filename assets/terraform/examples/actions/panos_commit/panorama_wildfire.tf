# Panorama commit for WildFire appliances and clusters
action "panos_commit" "panorama_wildfire" {
  config {
    description         = "Commit WildFire configuration"
    wildfire_appliances = ["wf-appliance-1"]
    wildfire_clusters   = ["wf-cluster-1"]
  }
}
