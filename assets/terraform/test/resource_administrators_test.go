// Add missing attributes to panos_administrator.example from panosAdministratorsTmpl1 based on administrators.yaml definition.
package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	//"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccAdministrators(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	location := config.ObjectVariable(map[string]config.Variable{
		"panorama": config.ObjectVariable(map[string]config.Variable{}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: panosAdministratorsTmpl1,
				ConfigVariables: map[string]config.Variable{
					"prefix":         config.StringVariable(prefix),
					"location":       location,
					"ssh_public_key": config.StringVariable(sshPublicKey),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_administrator.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(prefix),
					),
				},
			},
		},
	})
}

const panosAdministratorsTmpl1 = `
variable "location" { type = any }
variable "prefix" { type = string }
variable "ssh_public_key" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = format("%s-tmpl", var.prefix)
}

resource "panos_authentication_profile" "example" {
  location = var.location

  name = var.prefix

  allow_list = ["all"]

  lockout = {
    failed_attempts = 5
    lockout_time    = 30
  }

  method = {
    none = {}
  }
}

resource "panos_administrator" "example" {
  location = { panorama = {} }

  name = var.prefix

  authentication_profile = panos_authentication_profile.example.name

  client_certificate_only = false
  #password_profile = "default"

  permissions = {
    role_based = {
      superuser = "yes"
    }
  }

  phash = "examplehash"

  preferences = {
    disable_dns = false
    saved_log_query = {
      traffic = [{
        name = "example_traffic_query"
        query = "receive_time >= '2023-01-01' and receive_time <= '2023-12-31'"
      }]
    }
  }

  public_key = var.ssh_public_key
}
`

const sshPublicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDJSkxCN4fQHapPmOWQOHllXCc2amr6CrLPnUUV6zPu8XclJiklQi4qIfYrxbm0UFpdrXcJT/g8rtqXFY7jnDBiGZbG47gykAt6befSj1pV5LBzSmDatyiwi8uCvovt8902kHOaNodkOt9pwvMVtzJ2vS6Y/qjL0mdev5d5azJRytSW7h33yG4PZv6/GkV9MuUfGWFXU6HHsHA48b+oPc0YF0zWfG1yCWjmgq+3tm2tO3mjkKdBP2BBZbHfZnFxRsDBpvddfjEPcZXJuzzt7T2H3vlyigf7WQT39s4eZMjMI7JF1bs34iSIo2kGMUkHVLxu1nEi+WlXJByIpo0YxXa2V2A9YIpoJzZ1x+Ma8JDPmcrT1PbBROxK3ZNytidliH+j8LSuNRgIygfmFLSrqLI6UJNuHtWf8lpNo50JZuzedeieffa2NzHHazx7ZYC0qBu8eTVzz/rZHzto1+nfB/CS+XsUZ1VV/mVIv68xa4spKlmItTSuKiJIkOFaGYHJgjwCId3b9NSz5gOkacHvYBec1vWQyJ8a6Vr1uU416KKTbMMvnBFe5vpRIicg75lHWalBZspu2XZBiSdKhMUO0qvDBooPzeS0f3XecIVZ1+u7dsl9hA9N3BRseYwxyVhK7ocfGmhVA5azouzo1p42AYvA5I9RANiHO5IqaxIV2LDH2Q== jdoe@example.com`
