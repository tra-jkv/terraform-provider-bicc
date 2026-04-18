---
page_title: "bicc_job Resource - BICC Provider"
description: |-
  Manages an Oracle BICC extraction job.
---

# bicc_job (Resource)

Manages an Oracle BICC extraction job, including its data stores, column configuration, and incremental extraction settings.

## Example Usage

### Basic Job with Auto-Populate Columns (Recommended)

```terraform
resource "bicc_job" "suppliers" {
  name        = "Suppliers"
  description = "Supplier master data - suppliers and sites"

  data_stores = [
    {
      data_store_key             = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
      is_silent_error            = true
      is_effective_date_disabled = false
      use_union_for_incremental  = true
      auto_populate_all_columns  = true
      filters                    = null
      initial_extract_date       = null
      chunk_type                 = null
      chunk_date_seq_incr        = 0
      chunk_date_seq_min         = 0
      chunk_pk_seq_incr          = 0
      column_overrides = [{
        name                    = "LastUpdateDate"
        is_last_update_date     = true
        is_populate             = false
        is_primary_key          = false
        is_creation_date        = false
        is_effective_start_date = false
        is_natural_key          = false
      }]
      columns = []
    },
    {
      data_store_key             = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierSiteExtractPVO"
      is_silent_error            = true
      is_effective_date_disabled = false
      use_union_for_incremental  = true
      auto_populate_all_columns  = true
      filters                    = null
      initial_extract_date       = null
      chunk_type                 = null
      chunk_date_seq_incr        = 0
      chunk_date_seq_min         = 0
      chunk_pk_seq_incr          = 0
      column_overrides = [{
        name                    = "LastUpdateDate"
        is_last_update_date     = true
        is_populate             = false
        is_primary_key          = false
        is_creation_date        = false
        is_effective_start_date = false
        is_natural_key          = false
      }]
      columns = []
    },
  ]
}
```

### Job with Manual Column Selection

```terraform
resource "bicc_job" "crm_extract" {
  name        = "CRMExtract"
  description = "CRM person data with selected columns"

  data_stores = [
    {
      data_store_key             = "CrmAnalyticsAM.PartiesAnalyticsAM.Person"
      is_silent_error            = true
      is_effective_date_disabled = false
      use_union_for_incremental  = false
      auto_populate_all_columns  = false
      filters                    = null
      initial_extract_date       = null
      chunk_type                 = null
      chunk_date_seq_incr        = 0
      chunk_date_seq_min         = 0
      chunk_pk_seq_incr          = 0
      column_overrides           = []
      columns = [
        {
          name                    = "PersonProfileId"
          is_populate             = true
          is_primary_key          = false
          is_last_update_date     = false
          is_creation_date        = false
          is_effective_start_date = false
          is_natural_key          = false
        },
        {
          name                    = "LastUpdateDate"
          is_populate             = true
          is_primary_key          = false
          is_last_update_date     = true
          is_creation_date        = false
          is_effective_start_date = false
          is_natural_key          = false
        },
      ]
    },
  ]
}
```

### Job with Incremental Extraction and Date Chunking

```terraform
resource "bicc_job" "inventory" {
  name        = "InventoryValuations"
  description = "Inventory valuations with date-based chunking"

  data_stores = [
    {
      data_store_key             = "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstInventoryValuationExtractPVO"
      is_silent_error            = true
      is_effective_date_disabled = false
      use_union_for_incremental  = true
      auto_populate_all_columns  = true
      filters                    = null
      initial_extract_date       = "2024-01-01"
      chunk_type                 = "DATE"
      chunk_date_seq_incr        = 7
      chunk_date_seq_min         = 0
      chunk_pk_seq_incr          = 0
      column_overrides = [{
        name                    = "LastUpdateDate"
        is_last_update_date     = true
        is_populate             = false
        is_primary_key          = false
        is_creation_date        = false
        is_effective_start_date = false
        is_natural_key          = false
      }]
      columns = []
    },
  ]
}
```

## Upgrading from v1.x

In v2.0, `data_stores`, `columns`, and `column_overrides` changed from block syntax to assignment syntax.

**Before (v1.x):**
```terraform
data_stores {
  data_store_key = "..."
  columns {
    name = "LastUpdateDate"
  }
}
```

**After (v2.x):**
```terraform
data_stores = [
  {
    data_store_key = "..."
    columns = [
      {
        name = "LastUpdateDate"
        # all boolean fields must be explicitly set
        is_populate             = true
        is_primary_key          = false
        is_last_update_date     = true
        is_creation_date        = false
        is_effective_start_date = false
        is_natural_key          = false
      },
    ]
  },
]
```

All boolean fields within `columns` and `column_overrides` objects must be explicitly specified — there are no omittable defaults within nested list objects.

## Importing Existing Jobs

```shell
terraform import bicc_job.example 123456789
```

After importing, run `terraform plan`. If your config uses `auto_populate_all_columns = true`, the provider will suppress the one-time drift from the import automatically — no apply is needed.

## Schema

### Required

- `name` (String) - Name of the BICC job.
- `data_stores` (Set of Objects, Min: 1) - Set of data stores to extract (order-independent). See [data_stores](#nested-schema-for-data_stores) below.

### Optional

- `description` (String) - Description of the BICC job.

### Read-Only

- `id` (String) - The BICC job ID.

---

### Nested Schema for `data_stores`

#### Required

- `data_store_key` (String) - The unique key for the data store (e.g., `FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO`).

#### Optional

- `filters` (String) - Filter expression for data extraction (e.g., `__DATASTORE__.CreationDate > '2024-01-01'`). Set to `null` if unused.
- `is_silent_error` (Boolean) - If `true`, continue extraction even if this data store fails. Default: `false`.
- `is_effective_date_disabled` (Boolean) - If `true`, disable effective date filtering. Default: `false`.
- `use_union_for_incremental` (Boolean) - If `true`, enable incremental extraction using the UNION approach. Default: `false`.
- `initial_extract_date` (String) - Initial extract date for incremental extraction. Format: `YYYY-MM-DD`. Set to `null` to extract all historical data on the first run.
- `chunk_type` (String) - Chunking type for large extractions. Valid values: `DATE`, `SEQUENCE`. Set to `null` if unused.
- `chunk_date_seq_incr` (Number) - Date sequence increment for chunking (e.g., `7` for weekly chunks). Default: `0`.
- `chunk_date_seq_min` (Number) - Minimum date sequence for chunking. Default: `0`.
- `chunk_pk_seq_incr` (Number) - Primary key sequence increment for chunking. Default: `0`.
- `auto_populate_all_columns` (Boolean) - If `true`, automatically fetch and include all available columns from the data store. Use `column_overrides` to mark specific columns (e.g., `LastUpdateDate`) for incremental tracking. Default: `false`.
- `column_overrides` (List of Objects) - Column overrides applied when `auto_populate_all_columns = true`. Set to `[]` if unused. See [column_overrides](#nested-schema-for-data_storescolumn_overrides) below.
- `columns` (List of Objects) - Manual column configuration. Use when `auto_populate_all_columns = false`. Set to `[]` if unused. See [columns](#nested-schema-for-data_storescolumns) below.

---

### Nested Schema for `data_stores.column_overrides`

Used with `auto_populate_all_columns = true` to override specific column properties.

All fields must be explicitly specified within each object.

#### Required

- `name` (String) - Column name to override.
- `is_populate` (Boolean) - Override whether this column is included in extraction.
- `is_primary_key` (Boolean) - Mark as primary key column.
- `is_last_update_date` (Boolean) - Mark as last update date column. Set to `true` on `LastUpdateDate` for incremental extraction to work correctly.
- `is_creation_date` (Boolean) - Mark as creation date column.
- `is_effective_start_date` (Boolean) - Mark as effective start date column.
- `is_natural_key` (Boolean) - Mark as natural key column.

---

### Nested Schema for `data_stores.columns`

Used when `auto_populate_all_columns = false` for explicit column selection.

All fields must be explicitly specified within each object.

#### Required

- `name` (String) - Column name.
- `is_populate` (Boolean) - Include this column in extraction.
- `is_primary_key` (Boolean) - Mark as primary key column.
- `is_last_update_date` (Boolean) - Mark as last update date column.
- `is_creation_date` (Boolean) - Mark as creation date column.
- `is_effective_start_date` (Boolean) - Mark as effective start date column.
- `is_natural_key` (Boolean) - Mark as natural key column.
