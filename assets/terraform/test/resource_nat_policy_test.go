package provider_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"text/template"

	sdkerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/policies/rules/nat"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type deviceType int

const (
	devicePanorama deviceType = iota
	deviceFirewall deviceType = iota
)

type expectServerRulesOrder struct {
	Location  nat.Location
	Prefix    string
	RuleNames []string
}

func ExpectServerRulesOrder(prefix string, location nat.Location, ruleNames []string) *expectServerRulesOrder {
	return &expectServerRulesOrder{
		Location:  location,
		Prefix:    prefix,
		RuleNames: ruleNames,
	}
}

func (o *expectServerRulesOrder) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	service := nat.NewService(sdkClient)

	objects, err := service.List(ctx, o.Location, "get", "", "")
	if err != nil {
		resp.Error = fmt.Errorf("failed to query server for rules: %w", err)
		return
	}

	type ruleWithState struct {
		Idx   int
		State int
	}

	rulesWithIdx := make(map[string]ruleWithState)
	for idx, elt := range o.RuleNames {
		rulesWithIdx[fmt.Sprintf("%s-%s", o.Prefix, elt)] = ruleWithState{
			Idx:   idx,
			State: 0,
		}
	}

	var prevActualIdx = -1
	for actualIdx, elt := range objects {
		if state, ok := rulesWithIdx[elt.Name]; !ok {
			continue
		} else {
			state.State = 1
			rulesWithIdx[elt.Name] = state

			if state.Idx == 0 {
				prevActualIdx = actualIdx
				continue
			} else if prevActualIdx == -1 {
				resp.Error = fmt.Errorf("rules missing from the server")
				return
			} else if actualIdx-prevActualIdx > 1 {
				resp.Error = fmt.Errorf("invalid rules order on the server")
				return
			}
			prevActualIdx = actualIdx
		}
	}

	var missing []string
	for name, elt := range rulesWithIdx {
		if elt.State != 1 {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		resp.Error = fmt.Errorf("not all rules are present on the server: %s", strings.Join(missing, ", "))
		return
	}
}

func TestAccPanosNatPolicy(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	rulesInitial := []string{"rule-1", "rule-2", "rule-3"}
	rulesReordered := []string{"rule-2", "rule-1", "rule-3"}

	prefixed := func(name string) string {
		return fmt.Sprintf("%s-%s", prefix, name)
	}

	withPrefix := func(rules []string) []config.Variable {
		var result []config.Variable
		for _, elt := range rules {
			result = append(result, config.StringVariable(prefixed(elt)))
		}

		return result
	}

	device := devicePanorama

	location := locationByDeviceType(device)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             postTestCheck(prefix, device),
		Steps: []resource.TestStep{
			{
				Config: makeConfig(prefix, device),
				ConfigVariables: map[string]config.Variable{
					"rule_names": config.ListVariable(withPrefix(rulesInitial)...),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_nat_policy.%s", prefix),
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact(prefixed("rule-1")),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_nat_policy.%s", prefix),
						tfjsonpath.New("rules").AtSliceIndex(1).AtMapKey("name"),
						knownvalue.StringExact(prefixed("rule-2")),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_nat_policy.%s", prefix),
						tfjsonpath.New("rules").AtSliceIndex(2).AtMapKey("name"),
						knownvalue.StringExact(prefixed("rule-3")),
					),
					ExpectServerRulesOrder(prefix, *location, rulesInitial),
				},
			},
			{
				Config: makeConfig(prefix, device),
				ConfigVariables: map[string]config.Variable{
					"rule_names": config.ListVariable(withPrefix(rulesInitial)...),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: makeConfig(prefix, device),
				ConfigVariables: map[string]config.Variable{
					"rule_names": config.ListVariable(withPrefix(rulesReordered)...),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							fmt.Sprintf("panos_nat_policy.%s", prefix),
							tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("name"),
							knownvalue.StringExact(prefixed("rule-2")),
						),
						plancheck.ExpectKnownValue(
							fmt.Sprintf("panos_nat_policy.%s", prefix),
							tfjsonpath.New("rules").AtSliceIndex(1).AtMapKey("name"),
							knownvalue.StringExact(prefixed("rule-1")),
						),
						plancheck.ExpectKnownValue(
							fmt.Sprintf("panos_nat_policy.%s", prefix),
							tfjsonpath.New("rules").AtSliceIndex(2).AtMapKey("name"),
							knownvalue.StringExact(prefixed("rule-3")),
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_nat_policy.%s", prefix),
						tfjsonpath.New("rules").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact(prefixed("rule-2")),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_nat_policy.%s", prefix),
						tfjsonpath.New("rules").AtSliceIndex(1).AtMapKey("name"),
						knownvalue.StringExact(prefixed("rule-1")),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("panos_nat_policy.%s", prefix),
						tfjsonpath.New("rules").AtSliceIndex(2).AtMapKey("name"),
						knownvalue.StringExact(prefixed("rule-3")),
					),
					ExpectServerRulesOrder(prefix, *location, rulesReordered),
				},
			},
		},
	})
}

const configTmpl = `
variable "rule_names" { type = list(string) }

resource "panos_nat_policy" "{{ .ResourceName }}" {
{{- if .IsPanorama }}
  location = {
    shared = {
      rulebase = "pre-rulebase"
    }
  }
{{- else }}
  location = {
    vsys = {
      name = "vsys1"
    }
  }
{{- end }}

  rules = [
    for index, name in var.rule_names: {
      name = name
      source_zones = ["any"]
      source_addresses = ["any"]
      destination_zone = ["external"]
      destination_addresses = ["any"]

      destination_translation = {
        translated_address = format("172.16.0.%s", index)
      }
    }
  ]
}
`

func makeConfig(prefix string, deviceType deviceType) string {
	var buf bytes.Buffer
	tmpl := template.Must(template.New("").Parse(configTmpl))

	context := struct {
		IsPanorama   bool
		ResourceName string
		DeviceGroup  string
	}{
		IsPanorama:   deviceType == devicePanorama,
		ResourceName: prefix,
		DeviceGroup:  fmt.Sprintf("%s-dg", prefix),
	}

	err := tmpl.Execute(&buf, context)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

var DanglingObjectsError = errors.New("some objects were not deleted by the provider")

func locationByDeviceType(typ deviceType) *nat.Location {
	var location nat.Location
	switch typ {
	case devicePanorama:
		location = nat.Location{
			Shared: &nat.SharedLocation{
				Rulebase: "pre-rulebase",
			},
		}
	case deviceFirewall:
		location = nat.Location{
			Vsys: &nat.VsysLocation{
				NgfwDevice: "localhost.localdomain",
				Vsys:       "vsys1",
			},
		}
	}

	return &location
}

func postTestCheck(prefix string, deviceType deviceType) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		service := nat.NewService(sdkClient)
		ctx := context.TODO()

		location := locationByDeviceType(deviceType)

		rules, err := service.List(ctx, *location, "get", "", "")
		if err != nil && !sdkerrors.IsObjectNotFound(err) {
			panic(err)
		}

		var foundDanglingRules bool
		for _, elt := range rules {
			if strings.HasPrefix(elt.Name, prefix) {
				foundDanglingRules = true
			}
		}

		if foundDanglingRules {
			return DanglingObjectsError
		}

		return nil
	}
}

func init() {
	resource.AddTestSweepers("pango_nat_policy", &resource.Sweeper{
		Name: "pango_nat_policy",
		F: func(typ string) error {
			service := nat.NewService(sdkClient)

			var deviceTyp deviceType
			switch typ {
			case "panorama":
				deviceTyp = devicePanorama
			case "firewall":
				deviceTyp = deviceFirewall
			default:
				panic("invalid device type")
			}

			location := locationByDeviceType(deviceTyp)
			ctx := context.TODO()
			objects, err := service.List(ctx, *location, "get", "", "")
			if err != nil && !sdkerrors.IsObjectNotFound(err) {
				return fmt.Errorf("Failed to list NAT rules during sweep: %w", err)
			}

			var names []string
			for _, elt := range objects {
				if strings.HasPrefix(elt.Name, "test-acc") {
					names = append(names, elt.Name)
				}
			}

			if len(names) > 0 {
				err = service.Delete(ctx, *location, names...)
				if err != nil {
					return fmt.Errorf("Failed to delete NAT rules during sweep: %w", err)
				}
			}

			return nil
		},
	})
}
