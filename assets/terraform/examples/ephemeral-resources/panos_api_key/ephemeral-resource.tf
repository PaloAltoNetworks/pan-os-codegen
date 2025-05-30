# Allows you to generate an API key based on the provided username and password for the `panos_api_key` ephemeral resource.
# Note: The current implementation still requires you to provide a username/password when configuring the provider.
# If you use the same admin account username/password within the provider configuration and the ephemeral resource,
# you will not be able to use the same provider with other resources. The reason is that when we generate an API key
# with the `panos_api_key` ephemeral resource, the old tokens are invalidated automatically (i.e., the token for the 
# provider itself). To avoid conflicts, consider using different credentials for the provider configuration and the 
# ephemeral resource.

# Use cases: 
# - Store short lived API keys in a Cloud Key Management Service which will also support ephemeral resources
# - Dynamically configure diffferent instances of panos provider instances bound to different admin accounts

provider "panos" {
  hostname = "<hostname>"
  username = "<username>"
  password = "<password>"
}

ephemeral "panos_api_key" "example" {
  username = "<user-1>"
  password = "<password>"
}

# Use case 1: Configure a new provider with the new API key
provider "panos" {
  hostname = "<hostname>"
  api_key  = ephemeral.panos_api_key.example.api_key

  alias = "user1"
}
