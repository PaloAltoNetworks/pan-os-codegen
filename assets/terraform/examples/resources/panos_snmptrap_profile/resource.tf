resource "panos_snmptrap_profile" "example" {
	location = {
		vsys = {
			name = "vsys1"
		}
	}
	name = "my-snmp-trap-profile"

	version = {
		v2c = {
			servers = [
				{
					name      = "snmp-server-1"
					manager   = "192.0.2.1"
					community = "public"
				}
			]
		}
	}
}
