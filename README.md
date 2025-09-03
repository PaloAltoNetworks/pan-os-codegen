# PAN-OS Code Generation Repository (pan-os-codegen)

Welcome to the PAN-OS Code Generation Repository! This repository provides tools for generating
the [pango SDK](https://github.com/PaloAltoNetworks/pango) and
the `panos` [Terraform provider](https://github.com/PaloAltoNetworks/terraform-provider-panos) for Palo Alto Networks
PAN-OS devices.

## Overview

PAN-OS is the operating system for Palo Alto Networks next-generation firewalls and Panorama, providing advanced
security features and capabilities. This repository aims to simplify the process of building and maintainging the Go SDK
and Terraform provider.

The repository contains:

- Spec files: Represent a normalised version of the PAN-OS XML schema.
- Code generator: Generates the Pango SDK and the PAN-OS Terraform provider based on the spec files.

## Roadmap

We are maintaining a [public roadmap](https://github.com/orgs/PaloAltoNetworks/projects/62) to help users understand
when we will release new features, bug fixes and enhancements.

## Getting Help

Open an [issue](https://github.com/PaloAltoNetworks/pan-os-codegen/issues) on Github.

## Usage

The code have run login in `cmd/codegen` directory, to run it with default option please use:

```bash
go run cmd/codegen/main.go
```
This command can be parametrizes using options:
- `-t/-type` - operation type, default is to create both Terraform
  - `mktp` - create only Terraform provider
  - `mksdk` - create only PAN-OS SDK
- `config` - specify path for the config file, default is `cmd/codegen/config.yaml`

You can control logging level with `CODEGEN_LOG_LEVEL' environment variable, setting to accepted
values: *error*, *warning*, *info*, *debug*.

## Generate SDK

In order to use generated SDK code, go to directory defined in `config.yaml` e.g. `../generated/pango` and execute
example code:

```
go run cmd/codegen/main.go -t mksdk
cd ../generated/pango
PANOS_HOSTNAME='***' PANOS_USERNAME='***' PANOS_PASSWORD='***' go run example/main.go
```

## Acceptance testing
The acceptance test suite creates real resources in a configured instance. We need to set the following environment variables in order to run an acceptance test against the PANOS provider.

```sh
# set up the Terraform plugin testing framework in acceptance testing mode
# explicitly to allow the creation of real resources using a provider
# See: https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests#requirements-and-recommendations
export TF_ACC=1

# Provider configurations
export PANOS_HOSTNAME=...

# if we have self-signed certificate for a testing instance
export PANOS_SKIP_VERIFY_CERTIFICATE=true

# API key
export PANOS_API_KEY=...
```

Consult this [documentation page](https://docs.paloaltonetworks.com/pan-os/11-0/pan-os-panorama-api/get-started-with-the-pan-os-xml-api/get-your-api-key) to obtain an API key. The same page mentions that:

> If you have an existing key and generate another key for the same user, all existing sessions will end for the user and previous API sessions will be deleted.

This is the reason why we don't use user name and password for a provider configuration with the environment variables `PANOS_USERNAME` and `PANOS_PASSWORD` in the context of acceptance tests.

We can run the acceptance test suite with the following command.
```sh
$ go test ./test/... -v -count 1 -parallel 20  -timeout 180m
```
