#!/bin/bash

set -e

# This script validates all terraform examples.
# It should be run from the root of the repository.

cd assets/terraform/examples

# Create provider configuration template
cat << EOF > provider.tf.template
terraform {
  required_providers {
    panos = {
      source = "PaloAltoNetworks/panos"
      version = "1.0.0"
    }
  }
}

provider "panos" {
  alias = "ci"
}
EOF

# Find all directories containing .tf files
DIRS=$(find . -type f -name "*.tf" -exec dirname {} \; | grep -v "./provider" | sort -u)

# Loop through each directory and validate
for dir in $DIRS; do

  # Copy provider configuration for validation
  cp provider.tf.template "$dir/provider.tf"
  
  echo "Validating configurations in: $dir"
  (
    cd "$dir"

    # Initialize and validate
    [[ $CI == "true" ]] && terraform init -no-color
    OUTPUT=$(terraform validate -no-color)
    if [ $? -ne 0 ]; then
        echo "Vailed to validate examples: $dir"
        echo "${OUTPUT}"
        exit 1
    fi
    
    # Clean up provider configuration
    rm provider.tf
  )
done

# Clean up template
rm provider.tf.template

echo "All examples validated successfully."
