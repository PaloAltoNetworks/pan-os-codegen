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

## Generate SDK

In order to use generated SDK code, go to directory defined in `config.yaml` e.g. `../generated/pango` and execute
example code:

```
go run cmd/codegen/main.go -t mksdk
cd ../generated/pango
PANOS_HOSTNAME='***' PANOS_USERNAME='***' PANOS_PASSWORD='***' go run example/main.go
```