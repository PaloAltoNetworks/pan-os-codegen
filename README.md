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

```bash
go run cmd/mktp/main.go cmd/mktp/config.yaml
```