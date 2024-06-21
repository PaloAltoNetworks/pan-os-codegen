### How to use it locally

Firstly you need to create `.terraformrc` file in your home directory:

```shell
touch ~/.terraformrc
```

The content of the file should look like that: 

```
provider_installation {

  dev_overrides {
      "github.com/PaloAltoNetworks/panos" = "/home/USERNAME/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Now when you generate the terraform-provider go to `terraform-provider-panos` folder and follow steps:

Install dependencies:
```shell
go mod tidy
```

Install provider:

```shell
go install .
```

Now you can use Terraform with locally built provider.

Example Terraform file: 

`provider.tf`
```terraform
# Traditional provider example.
provider "panos" {
  hostname = ""
  username = ""
  password = ""
  skip_verify_certificate = true
}

# Local inspection mode provider example.
# provider "panos" {
#   #config_file = file("/tmp/candidate-config.xml")
#
#   # This is only used if a "detail-version" attribute is not present in
#   # the exported XML schema. If it's there, this can be omitted.
#   panos_version = "10.2.0"
# }

terraform {
  required_providers {
    panos = {
      source  = "github.com/PaloAltoNetworks/panos"
      version = "2.0.0"
    }
  }
}
```

`main.tf`

```terraform
resource "panos_address_group" "test1" {
  location = {
    device_group = {
      name = "test123"
    }
  }

  name = "test123123"
}

resource "panos_address_object" "test1" {
  location = {
    device_group = {
      name = "test123"
    }
  }

  name = "test123123"
  fqdn = "10.0.0.0"
  ip_netmask = "8"
}
```

### Known issue:
In early stages of Terraform Provider development helps a lot manually removing the resources and the statefile, after that some errors not occurred again.
