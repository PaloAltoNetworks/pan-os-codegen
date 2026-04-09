# Local XML Mode & Experimental Batching/Querying

## Local XML Mode

Local XML mode lets the PAN-OS Terraform provider read from and write to a local XML configuration file instead of connecting to a live PAN-OS device. This is useful for offline configuration management, testing, and CI/CD pipelines.

### Enabling Local XML Mode

Set the `xml_file_path` provider parameter to a PAN-OS configuration XML file:

```hcl
provider "panos" {
  xml_file_path = "/path/to/running-config.xml"
}
```

Or via environment variable:

```bash
export PANOS_XML_FILE_PATH="/path/to/running-config.xml"
```

**Mutual exclusivity:** `xml_file_path` cannot be used together with API mode parameters (`hostname`/`api_key`) or local inspection mode (`config_file`).

### How It Works

When `xml_file_path` is set, the provider creates a `LocalXmlClient` instead of an API client. This client:

1. Loads the XML file into an in-memory DOM tree
2. Auto-detects the PAN-OS version from the `detail-version` attribute
3. Auto-detects the device type (Firewall vs Panorama)
4. Performs all CRUD operations against the DOM tree
5. Writes changes back to the file (behavior depends on write mode)

All standard Terraform operations work: `plan`, `apply`, `destroy`, `import`.

### Resource Support

Most resources (~88%) support local XML mode automatically. Resources that do **not** support it include:

- **Custom-type resources** (API key management, certificate import, etc.)
- **Resources with hashed fields** (passwords that can't round-trip through XML)

If you use an unsupported resource in local XML mode, you'll get a clear error:

```
Error: Resource not supported in local XML mode
The resource 'panos_xxx' does not support local XML mode.
Please use API mode or refer to documentation.
```

### XML Write Modes

Control how changes are persisted to disk with `xml_write_mode`:

| Mode | Description | Best For |
|------|-------------|----------|
| `safe` (default) | Writes to disk after every operation | General use, data safety |
| `deferred` | Async writes with dirty flag checking | Bulk operations |
| `periodic` | Timer-based batched writes | Continuous updates |

```hcl
provider "panos" {
  xml_file_path = "/path/to/config.xml"

  # Write mode selection
  xml_write_mode = "safe"  # "safe" | "deferred" | "periodic"

  # Deferred mode tuning (only when xml_write_mode = "deferred")
  xml_write_check_interval_ms = 10  # range: 5-20

  # Periodic mode tuning (only when xml_write_mode = "periodic")
  xml_write_flush_interval_sec = 30  # range: 1-3600
}
```

Environment variables: none for write mode options (Terraform config only).

### Comparison: Local XML vs Local Inspection vs API

| Feature | API Mode | Local XML Mode | Local Inspection Mode |
|---------|----------|----------------|----------------------|
| Connection | Live PAN-OS device | Local XML file | Local XML file |
| Read | Yes | Yes | Yes |
| Write | Yes | Yes | **No** (read-only) |
| Provider param | `hostname` + `api_key` | `xml_file_path` | `config_file` |
| Version detection | Automatic | From XML `detail-version` | Manual (`panos_version`) |

---

## Experimental Batching & Querying

These provider-level settings tune how the provider reads and writes data. All are optional and have sensible defaults. They are marked **experimental** and may change in future releases.

### Provider Configuration

```hcl
provider "panos" {
  hostname = "192.168.1.1"
  api_key  = "your-api-key"

  # Write batching (not experimental, but included for completeness)
  multi_config_batch_size = 500  # default: 500, range: 1-10000

  # Experimental read/list tuning
  experimental_list_strategy     = "eager"    # "eager" (default) | "lazy"
  experimental_read_batch_size   = 50         # default: 50, range: 1-1000
  experimental_sharding_strategy = "disabled" # "disabled" (default) | "enabled"
  experimental_cache_strategy    = "disabled" # "disabled" (default) | "enabled"
}
```

All experimental options have environment variable equivalents:

| Parameter | Environment Variable | Default |
|-----------|---------------------|---------|
| `multi_config_batch_size` | `PANOS_MULTI_CONFIG_BATCH_SIZE` | `500` |
| `experimental_list_strategy` | `PANOS_LIST_STRATEGY` | `eager` |
| `experimental_read_batch_size` | `PANOS_READ_BATCH_SIZE` | `50` |
| `experimental_sharding_strategy` | `PANOS_SHARDING_STRATEGY` | `disabled` |
| `experimental_cache_strategy` | `PANOS_CACHE_STRATEGY` | `disabled` |

### List Strategies

Controls how the provider fetches collections of resources (e.g., all address objects).

#### Eager (default)

Single query fetches all entries with full details. Simple, works well for small-to-medium collections.

```
API call: GET all entries with full data → done
```

#### Lazy

Two-phase approach: first lists entry names, then fetches entries in batches. Better for very large collections (1000+ entries) where a single query might time out or consume excessive memory.

```
Phase 1: GET entry names only (lightweight)
Phase 2: GET entries in batches of `read_batch_size`
```

Configure batch size:

```hcl
experimental_list_strategy   = "lazy"
experimental_read_batch_size = 100  # fetch 100 entries per batch
```

### Sharding Strategy

Only applies when using the **lazy** list strategy. Controls how the initial name listing query is executed.

#### Disabled (default)

Single query to list all entry names.

#### Enabled

Splits the name listing into 7 parallel queries based on the first character of entry names:

| Shard | Characters |
|-------|-----------|
| 1 | `0-4` |
| 2 | `5-9` |
| 3 | `a-f` (case-insensitive) |
| 4 | `g-l` (case-insensitive) |
| 5 | `m-r` (case-insensitive) |
| 6 | `s-z` (case-insensitive) |
| 7 | Special characters (`-`, `_`, `.`, `@`, `#`, etc.) |

Useful when you have tens of thousands of entries and the name listing itself is slow.

```hcl
experimental_list_strategy     = "lazy"
experimental_sharding_strategy = "enabled"
```

### Cache Strategy

When enabled, caches resource data in memory to reduce redundant API calls within a single Terraform operation (`plan` or `apply`).

```hcl
experimental_cache_strategy = "enabled"
```

**Important:** Caching is a two-level opt-in:

1. **Provider level:** `experimental_cache_strategy = "enabled"` turns on the caching infrastructure
2. **Resource level:** Each resource spec must have `experimental_cache_enabled: true` to participate

Resources without `experimental_cache_enabled: true` in their spec won't be cached even when the provider-level strategy is enabled. Currently, resources like `address` have caching enabled in their specs.

Cache behavior:
- Uses deep copies to prevent mutation issues
- Thread-safe via read-write locks
- Scoped to a single Terraform operation (not persisted across runs)

### Write Batching (MultiConfig)

`multi_config_batch_size` controls how many write operations are sent in a single `MultiConfig` API call. This is used during `terraform apply` when creating, updating, or deleting multiple resources.

```hcl
multi_config_batch_size = 500  # default
```

- Operations are chunked into batches of this size
- Each batch is an atomic transaction (all succeed or all roll back)
- Increase for very large applies; decrease if hitting API limits

---

## Recommended Configurations

### Small environment (< 100 resources)

```hcl
provider "panos" {
  hostname = "firewall.example.com"
  api_key  = "..."
  # All defaults are fine
}
```

### Large environment (1000+ resources)

```hcl
provider "panos" {
  hostname = "firewall.example.com"
  api_key  = "..."

  experimental_list_strategy     = "lazy"
  experimental_read_batch_size   = 100
  experimental_sharding_strategy = "enabled"
  experimental_cache_strategy    = "enabled"
  multi_config_batch_size        = 500
}
```

### Offline/CI configuration generation

```hcl
provider "panos" {
  xml_file_path  = "/path/to/config.xml"
  xml_write_mode = "deferred"  # faster for bulk operations
}
```
