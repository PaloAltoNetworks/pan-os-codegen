---
page_title: 'Migration Guide'
---

v2.0.0 of the Terraform provider introduce major breaking changes to the schemas and there is no automatic state upgrades.

Most of the resources support being imported, even if there is no import section present in the docs. We are in the process of updating the documentation and all resources will include an import sectiton in future.

The import ID of a resource is the base64encoded version of the Terraform resource configuration.

## Importing an Address

```hcl
# An address can be imported by providing the following base64 encoded object as the ID
{
  location = {
    device_group = {
      name            = "example-device-group"
      panorama_device = "localhost.localdomain"
 }
 }

  name = "addr1"
}
```

```bash
terraform import panos_address.example $(echo '{"location":{"device_group":{"name":"example-device-group","panorama_device":"localhost.localdomain"}},"name":"addr1"}' | base64)
```

## Importing the entire security policy rule base

```hcl
# The entire policy can be imported by providing the following base64 encoded object as the ID
{
    location = {
        device_group = {
        name = "example-device-group"
        rulebase = "pre-rulebase"
        panorama_device = "localhost.localdomain"
 }
 }


    names = [
        "rule-1", <- the first rule in the policy
 ]
}
```

```bash
terraform import panos_security_policy.example $(echo '{"location":{"device_group":{"name":"example-device-group","panorama_device":"localhost.localdomain","rulebase":"pre-rulebase"}},"names":["rule-1"]}' | base64)
```
