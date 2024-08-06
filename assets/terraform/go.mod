module github.com/PaloAltoNetworks/terraform-provider-panos

go 1.22

require (
	github.com/PaloAltoNetworks/pango v0.10.3-0.20240408115758-216d8509e7cf
	github.com/hashicorp/terraform-plugin-framework v1.8.0
)

replace github.com/PaloAltoNetworks/pango v0.10.3-0.20240408115758-216d8509e7cf => ../pango
