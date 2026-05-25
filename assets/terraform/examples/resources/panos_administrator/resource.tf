resource "panos_administrator" "example" {
	location = {
		panorama = {}
	}
	name                   = "admin-user"
	authentication_profile = "my-auth-profile"

	permissions = {
		role_based = {
			superuser = true
		}
	}
}
