# Look up predefined DLP file type properties for use with custom data patterns.
#
# Each predefined file type (pdf, rtf, docx, etc.) has file properties with
# internal names (e.g. "panav-rsp-rtf-dlp-keywords") and human-readable labels
# (e.g. "Keywords/Tags"). When configuring panos_custom_data_object resources
# with file-properties pattern types, the file_property field requires the
# internal name — use this data source to resolve it from the label.

# Define custom data pattern input — users typically pass this as a variable.
# The file_property label is the human-readable name shown in the PAN-OS UI.
variable "custom_data_pattern" {
  default = {
    name = "example-pattern"
    file_property = [
      {
        name           = "keyword-check"
        file_type      = "rtf"
        file_property  = "Keywords/Tags"
        property_value = "confidential"
      },
      {
        name           = "author-check"
        file_type      = "pdf"
        file_property  = "Author"
        property_value = "admin"
      },
    ]
  }
}

# Look up each file type referenced in the pattern.
data "panos_predefined_dlp_file_type" "rtf" {
  location = { predefined = {} }
  name     = "rtf"
}

data "panos_predefined_dlp_file_type" "pdf" {
  location = { predefined = {} }
  name     = "pdf"
}

locals {
  # Map file_type -> properties list for easy lookup
  dlp_file_types = {
    "rtf" = data.panos_predefined_dlp_file_type.rtf.file_property
    "pdf" = data.panos_predefined_dlp_file_type.pdf.file_property
  }

  # Resolve each label to its internal property name
  resolved = [
    for fp in var.custom_data_pattern.file_property : {
      name           = fp.name
      file_type      = fp.file_type
      file_property  = one([
        for p in local.dlp_file_types[fp.file_type]
        : p.name if p.label == fp.file_property
      ])
      property_value = fp.property_value
    }
  ]
}

# Use the resolved properties in a custom data object.
resource "panos_custom_data_object" "example" {
  location = { shared = {} }
  name     = var.custom_data_pattern.name

  pattern_type = {
    file_properties = {
      pattern = [
        for fp in local.resolved : {
          name           = fp.name
          file_type      = fp.file_type
          file_property  = fp.file_property
          property_value = fp.property_value
        }
      ]
    }
  }
}
