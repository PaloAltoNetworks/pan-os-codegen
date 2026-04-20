resource "panos_email_server_profile" "example" {
	location = {
		vsys = {
			name = "vsys1"
		}
	}
	name = "my-email-server-profile"

	servers = [
		{
			name    = "email-server-1"
			from    = "panos@example.com"
			to      = "alerts@example.com"
			gateway = "smtp.example.com"
			port    = 25
		}
	]
}
