
resource "panos_device_group" "example" {
	location = { panorama = {} }
	name = "my-device-group"
}

resource "panos_custom_data_object" "example" {
	location = {
		device_group = panos_device_group.example.name
	}
	name = "my-custom-data-object"
	description = "test custom data object"
	pattern_type = {
		regex = {
			pattern = [
				{
					name = "pattern1"
					regex = "test-regex"
				}
			]
		}
	}
}

resource "panos_data_filtering_profile" "example" {
	location = {
		device_group = panos_device_group.example.name
	}
	name = "my-data-filtering-profile"
	data_capture = true
	description = "test description"
	disable_override = "yes"
	rules = [
		{
			name = "rule1"
			data_object = panos_custom_data_object.example.name
			direction = "both"
			alert_threshold = 10
			block_threshold = 20
			log_severity = "high"
			application = ["any"]
			file_type = ["any"]
		}
	]
}
