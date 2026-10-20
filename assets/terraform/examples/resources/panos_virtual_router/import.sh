#!/bin/bash
# Import a virtual router from a Panorama template.
# The complete import ID, including the location and resource name, must be base64-encoded.
import_id='{"location":{"template":{"name":"example-template","panorama_device":"localhost.localdomain","ngfw_device":"localhost.localdomain"}},"name":"production-vr"}'
encoded_import_id=$(printf '%s' "$import_id" | base64 | tr -d '\n')
terraform import "panos_virtual_router.example" "$encoded_import_id"
