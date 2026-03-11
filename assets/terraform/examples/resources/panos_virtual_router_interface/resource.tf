resource "panos_template" "example" {
    location = { panorama = {} }
    name     = "my-template"
}

resource "panos_virtual_router" "example" {
	location = {
		template = {
			name = panos_template.example.name
		}
	}
	name = "vr-1"
}

resource "panos_ethernet_interface" "example" {
	location = {
		template = {
			name = panos_template.example.name
		}
	}
	name = "ethernet1/1"
	layer3 = {
		ips = [{ name = "10.1.1.1/24" }]
	}
}

resource "panos_virtual_router_interface" "example" {
	location = {
		template = {
			name = panos_template.example.name
		}
	}
	virtual_router = panos_virtual_router.example.name
	interface      = panos_ethernet_interface.example.name
}