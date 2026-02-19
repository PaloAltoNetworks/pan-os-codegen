resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = "example-dg"
}

resource "panos_custom_data_object" "basic" {
    location = { device_group = { name = panos_device_group.example.name } }
    name = "basic"
    description = "test-description"
}

resource "panos_custom_data_object" "file_properties" {
    location = { device_group = { name = panos_device_group.example.name } }
    name = "file_properties"
	pattern_type = {
		file_properties = {
			pattern = [
				{
					name = "test-pattern"
					file_type = "pdf"
					file_property = "panav-rsp-pdf-dlp-author"
					property_value = "author"
				}
			]
		}
	}
}

resource "panos_custom_data_object" "predefined" {
    location = { device_group = { name = panos_device_group.example.name } }
    name = "predefined"
	pattern_type = {
		predefined = {
			pattern = [
				{
					name = "ABA-Routing-Number"
					file_type = ["xlsx"]
				},
				{
					name = "credit-card-numbers",
					file_type = ["text/html"]
				},
			]
		}
	}
}

resource "panos_custom_data_object" "regex" {
    location = { device_group = { name = panos_device_group.example.name } }
    name = "regex"
	pattern_type = {
		regex = {
			pattern = [
				{
					name = "test-pattern"
					regex = "test-regex"
				}
			]
		}
	}
}