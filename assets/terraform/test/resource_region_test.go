package provider_test

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccRegion(t *testing.T) {
	resourceName := "test_region"
	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"shared": config.ObjectVariable(map[string]config.Variable{}),
	})

	regionName := fmt.Sprintf("%s-region1", prefix)
	addresses := []config.Variable{
		config.StringVariable("10.220.34.0/24"),
		config.StringVariable("10.220.35.0/24"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: makeRegionConfig(resourceName),
				ConfigVariables: map[string]config.Variable{
					"location":    location,
					"region_name": config.StringVariable(regionName),
					"address":     config.ListVariable(addresses...),
					"latitude":    config.FloatVariable(1.283333),
					"longitude":   config.FloatVariable(103.833333),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_region.%s", resourceName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(regionName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_region.%s", resourceName),
						tfjsonpath.New("address"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("10.220.34.0/24"),
							knownvalue.StringExact("10.220.35.0/24"),
						}),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_region.%s", resourceName),
						tfjsonpath.New("latitude"),
						knownvalue.Float64Exact(1.283333),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_region.%s", resourceName),
						tfjsonpath.New("longitude"),
						knownvalue.Float64Exact(103.833333),
					),
				},
			},
			// Update: move the region and drop one address.
			{
				Config: makeRegionConfig(resourceName),
				ConfigVariables: map[string]config.Variable{
					"location":    location,
					"region_name": config.StringVariable(regionName),
					"address":     config.ListVariable(addresses[0]),
					"latitude":    config.FloatVariable(37.774929),
					"longitude":   config.FloatVariable(-122.419418),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_region.%s", resourceName),
						tfjsonpath.New("address"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("10.220.34.0/24"),
						}),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_region.%s", resourceName),
						tfjsonpath.New("latitude"),
						knownvalue.Float64Exact(37.774929),
					),
				},
			},
		},
	})
}

const regionResourceTmpl = `
variable "location" { type = map }
variable "region_name" { type = string }
variable "address" { type = list(string) }
variable "latitude" { type = number }
variable "longitude" { type = number }

resource "panos_region" "{{ .ResourceName }}" {
  location = var.location

  name      = var.region_name
  address   = var.address
  latitude  = var.latitude
  longitude = var.longitude
}
`

func makeRegionConfig(resourceName string) string {
	var buf bytes.Buffer
	tmpl := template.Must(template.New("").Parse(regionResourceTmpl))

	context := struct {
		ResourceName string
	}{
		ResourceName: resourceName,
	}

	err := tmpl.Execute(&buf, context)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
