# DGP-583 Codegen Changes

This document describes the four changes made to pan-os-codegen to fix perpetual plan diffs and
a crash in the generated Terraform provider when managing Panorama-based firewall resources.

---

## 1. `assets/terraform/internal/manager/uuid.go` — Fix crash/API error in `moveExhaustive`

### What changed
Added `"sort"` to imports. In `moveExhaustive`, replaced the naive index-based slice population
with a filter-then-sort approach:

**Before:**
```go
entries := make([]E, len(entriesByName))
for _, elt := range entriesByName {
    entries[elt.StateIdx] = elt.Entry
}
```

**After:**
```go
var active []idxEntry
for _, elt := range entriesByName {
    if elt.State != entryDeleted {
        active = append(active, idxEntry{elt.StateIdx, elt.Entry})
    }
}
sort.Slice(active, func(i, j int) bool { return active[i].idx < active[j].idx })
entries := make([]E, len(active))
for i, ae := range active { entries[i] = ae.entry }
```

### Why
When Terraform deletes a security rule (or any UUID-managed entry), the provider first calls
`Delete` on Panorama to remove the entry, then calls `moveExhaustive` to reorder the remaining
rules. At the point `moveExhaustive` runs, the deleted entries no longer exist on Panorama, but
they are still present in the `entriesByName` map with `State = entryDeleted`.

The old code allocated a slice of `len(entriesByName)` (which includes deleted entries) and then
used each entry's `StateIdx` as a direct index into that slice. Because deleted entries have
already been removed from the map's index space but their `StateIdx` values can collide with or
leave gaps among the surviving entries, the result is either nil slots in the slice (causing a
nil-pointer panic inside `MoveGroup`) or an API error when Panorama receives a move request for
a rule that no longer exists.

### What it fixes
- Eliminates panics/crashes when `terraform apply` removes one or more security rules in the
  same plan that also reorders remaining rules.
- Eliminates Panorama API errors (`Object not found`) returned from `MoveGroup` when deleted
  rule names are included in the move request.

---

## 2. `specs/network/virtual-router.yaml` — Add `vsys` default and `template-stack` vsys path

### What changed

**a) `imports.default_value: vsys1`**
```yaml
imports:
  target: virtual-router
  variants:
  - '*'
  default_value: vsys1     # added
```

**b) `vsys` added to `template-stack` location xpath and vars**
```yaml
- name: template-stack
  xpath:
    path:
    - ...
    - $ngfw_device
    - vsys          # added
    - $vsys         # added
    vars:
    - ...
    - name: vsys
      description: The vsys.
      required: false
      default: vsys1
      validators:
      - type: not-values
        spec:
          values:
          - value: shared
            error: The vsys cannot be "shared". Use the "shared" location instead.
      type: entry   # added
```

### Why

**Part a — imports default_value:**
Because `virtual-router` has `imports.variants: ['*']`, the code generator
(`pkg/translate/terraform_provider/location.go`) automatically injects a `vsys` input field into
the `template` location schema. The default value for that injected field is taken from
`imports.default_value`. Without a `default_value`, the generated `vsys` attribute has no
`Default`, so when a user imports a virtual router that lives in `vsys1` (the PAN-OS default),
Terraform records `vsys = "vsys1"` in state. On the next plan, because there is no schema
default to fall back to, Terraform plans to change `vsys` from `"vsys1"` to null and then
re-creates the resource (ForceNew), even though nothing actually changed.

**Part b — template-stack vsys path:**
The `template-stack` location was missing `vsys` entirely from its xpath, so resources managed
under a template stack were being addressed at the device level rather than the vsys level.
This caused import to succeed but subsequent reads to return the wrong (or empty) object,
leading to perpetual diffs or unexpected replacements.

### What it fixes
- Stops `panos_virtual_router` from planning a destroy/recreate purely because the `vsys`
  attribute has no default after import.
- Corrects the Panorama xpath used when the resource is managed via a template stack, preventing
  mismatched reads after import.

---

## 3. `specs/network/ike-gateway.yaml` — Mark `protocol-common` as Computed

### What changed
```yaml
- name: protocol-common
  type: object
  ...
  codegen_overrides:
    terraform:
      computed: true    # added
  spec:
```

### Why
`protocol-common` contains settings (NAT-traversal, IKE fragmentation, passive mode) that
PAN-OS always returns in a GET response with its own defaults, even when the user never
explicitly configured them. Before this change, `protocol-common` was `Optional` only in the
generated Terraform schema. The sequence that causes perpetual diffs:

1. User creates an IKE gateway without specifying `protocol_common`.
2. Terraform writes the resource; PAN-OS stores its own defaults for `protocol_common`.
3. On import (or a subsequent read), the provider reads back those PAN-OS defaults and stores
   them in state.
4. On the next plan, Terraform compares plan (no `protocol_common` block) against state
   (`protocol_common` populated with PAN-OS defaults) and plans to remove the block.
5. Apply removes the block; PAN-OS silently re-applies its defaults on the next read → loop.

Marking the field `computed: true` tells Terraform: "the provider may set this value even if
the user didn't". Terraform then keeps whatever the provider returns instead of planning to
remove it.

### What it fixes
Eliminates the perpetual plan diff where `protocol_common` shows as `- protocol_common { ... }`
on every plan, even though the user never set it and the firewall config did not change.

---

## 4. `specs/network/tunnels/ipsec.yaml` — Mark `tunnel-monitor` as Computed

### What changed
```yaml
- name: tunnel-monitor
  type: object
  ...
  codegen_overrides:
    terraform:
      computed: true    # added
  spec:
```

### Why
Same root cause as `protocol-common` in IKE gateway. `tunnel-monitor` contains monitoring
settings that PAN-OS populates with defaults even when the user omits the block. The generated
schema had `tunnel_monitor` as `Optional` only, causing Terraform to perpetually plan removal of
the block after import, because the provider's read response includes PAN-OS-default values but
the plan has no `tunnel_monitor` block.

### What it fixes
Eliminates the perpetual plan diff where `tunnel_monitor` shows as `- tunnel_monitor { ... }`
on every plan after importing an IPsec tunnel that was created without an explicit
`tunnel_monitor` block.
