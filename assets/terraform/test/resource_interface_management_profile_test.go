package provider_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	sdkErrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/network/profiles/interface_management"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccInterfaceManagementProfile(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	templateName := fmt.Sprintf("%s-tmpl", nameSuffix)
	location := config.ObjectVariable(map[string]config.Variable{
		"template": config.ObjectVariable(map[string]config.Variable{
			"name": config.StringVariable(templateName),
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             testAccInterfaceManagementProfileDestroy(prefix, templateName),
		Steps: []resource.TestStep{
			{
				Config: interfaceManagementProfileResourceTmpl,
				ConfigVariables: map[string]config.Variable{
					"location":      location,
					"template_name": config.StringVariable(templateName),
					"name":          config.StringVariable(prefix),
					"permitted_ips": config.ListVariable(
						config.ObjectVariable(map[string]config.Variable{
							"name": config.StringVariable("172.16.0.1"),
						}),
						config.ObjectVariable(map[string]config.Variable{
							"name": config.StringVariable("172.16.0.2"),
						}),
					),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("http"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("https"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("ping"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("response_pages"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("userid_service"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("userid_syslog_listener_ssl"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("userid_syslog_listener_udp"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("ssh"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("telnet"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("snmp"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("http_ocsp"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"panos_interface_management_profile.profile",
						tfjsonpath.New("permitted_ips"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("172.16.0.1"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("172.16.0.2"),
							}),
						}),
					),
				},
			},
		},
	})
}

const interfaceManagementProfileResourceTmpl = `
variable "location" { type = map }
variable "template_name" { type = string }
variable "name" { type = string }
variable "permitted_ips" {
  type = list(map(string))
  default = []
}

resource "panos_template" "template" {
  location = { panorama = {} }

  name = var.template_name
}

resource "panos_interface_management_profile" "profile" {
  depends_on = [ resource.panos_template.template ]
  location = var.location

  name = var.name

  http  = true
  https = true
  ping  = true

  response_pages = true

  userid_service             = true
  userid_syslog_listener_ssl = true
  userid_syslog_listener_udp = true

  ssh    = true
  telnet = true
  snmp   = true

  http_ocsp = true

  permitted_ips = var.permitted_ips
}
`

func testAccInterfaceManagementProfileDestroy(prefix string, template string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		api := interface_management.NewService(sdkClient)
		ctx := context.TODO()

		location := interface_management.NewTemplateLocation()
		location.Template = &interface_management.TemplateLocation{
			Template:       template,
			NgfwDevice:     "localhost.localdomain",
			PanoramaDevice: "localhost.localdomain",
		}

		entries, err := api.List(ctx, *location, "get", "", "")
		if err != nil && !sdkErrors.IsObjectNotFound(err) {
			return fmt.Errorf("listing interface management entries via sdk: %v", err)
		}

		var leftEntries []string
		for _, elt := range entries {
			if strings.HasPrefix(elt.Name, prefix) {
				leftEntries = append(leftEntries, elt.Name)
			}
		}

		if len(leftEntries) > 0 {
			err := fmt.Errorf("terraform failed to remove entries from the server")
			delErr := api.Delete(ctx, *location, leftEntries...)
			if delErr != nil {
				return errors.Join(err, delErr)
			}
		}

		return nil
	}
}