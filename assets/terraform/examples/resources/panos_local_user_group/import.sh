#!/bin/bash

# Import a local user group from template_vsys location
terraform import panos_local_user_group.example 'template_vsys:{"template":"example-template"}:example-group'

# Import a local user group from shared location
# terraform import panos_local_user_group.example 'shared::example-group'

# Import a local user group from vsys location
# terraform import panos_local_user_group.example 'vsys:{"vsys":"vsys1"}:example-group'
