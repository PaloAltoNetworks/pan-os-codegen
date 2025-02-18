# Generate the VM Auth Key on Panorama
# https://docs.paloaltonetworks.com/vm-series/11-0/vm-series-deployment/bootstrap-the-vm-series-firewall/generate-the-vm-auth-key-on-panorama

ephemeral "panos_vm_auth_key" "this" {
  lifetime = 1
}
