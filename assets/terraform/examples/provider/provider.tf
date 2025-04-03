# Username and password based authentication
provider "panos" {
  hostname = "hostname"
  username = "username"
  password = "password"
}

# API key based authentication
provider "panos" {
  hostname = "hostname"
  api_key  = "api_key"
}