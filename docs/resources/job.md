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

  data_stores {
    data_store_key            = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
    is_silent_error           = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  data_stores {
    data_store_key            = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierSiteExtractPVO"
    is_silent_error           = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}
```

### Job with Manual Column Selection

```terraform
resource "bicc_job" "crm_extract" {
  name        = "CRMExtract"
  description = "CRM person data with selected columns"

  data_stores {
    data_store_key             = "CrmAnalyticsAM.PartiesAnalyticsAM.Person"
    is_silent_error            = true
    is_effective_date_disabled = false

    columns {
      name        = "PersonProfileId"
      is_populate = true
    }

    columns {
      name                = "LastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }
  }
}
```

### Job with Incremental Extraction and Date Chunking

```terraform
resource "bicc_job" "inventory" {
  name        = "InventoryValuations"
  description = "Inventory valuations with date-based chunking"

  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstInventoryValuationExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2024-01-01"
    chunk_type                 = "DATE"
    chunk_date_seq_incr        = 7
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}
```

## Schema

### Required

- `name` (String) - Name of the BICC job.
- `data_stores` (Block Set, Min: 1) - Set of data stores to extract (order-independent). See [data_stores](#nested-schema-for-data_stores) below.

### Optional

- `description` (String) - Description of the BICC job.

### Read-Only

- `id` (String) - The BICC job ID.

---

### Nested Schema for `data_stores`

#### Required

- `data_store_key` (String) - The unique key for the data store (e.g., `FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO`).

#### Optional

- `filters` (String) - Filter expression for data extraction (e.g., `__DATASTORE__.CreationDate > '2024-01-01'`).
- `is_silent_error` (Boolean) - If `true`, continue extraction even if this data store fails. Default: `false`.
- `is_effective_date_disabled` (Boolean) - If `true`, disable effective date filtering. Default: `false`.
- `use_union_for_incremental` (Boolean) - If `true`, enable incremental extraction using the UNION approach. Default: `false`.
- `initial_extract_date` (String) - Initial extract date for incremental extraction. Format: `YYYY-MM-DD`. Omit to extract all historical data on the first run.
- `chunk_type` (String) - Chunking type for large extractions. Valid values: `DATE`, `SEQUENCE`.
- `chunk_date_seq_incr` (Number) - Date sequence increment for chunking (e.g., `7` for weekly chunks). Default: `0`.
- `chunk_date_seq_min` (Number) - Minimum date sequence for chunking. Default: `0`.
- `chunk_pk_seq_incr` (Number) - Primary key sequence increment for chunking. Default: `0`.
- `auto_populate_all_columns` (Boolean) - If `true`, automatically fetch and include all available columns from the data store. Use `column_overrides` to mark specific columns (e.g., `LastUpdateDate`) for incremental tracking. Default: `false`.
- `column_overrides` (Block List) - Column overrides applied when `auto_populate_all_columns = true`. See [column_overrides](#nested-schema-for-data_storescolumn_overrides) below.
- `columns` (Block List) - Manual column configuration. Use when `auto_populate_all_columns = false`. See [columns](#nested-schema-for-data_storescolumns) below.

---

### Nested Schema for `data_stores.column_overrides`

Used with `auto_populate_all_columns = true` to override specific column properties.

#### Required

- `name` (String) - Column name to override.

#### Optional

- `is_populate` (Boolean) - Override whether this column is included in extraction.
- `is_primary_key` (Boolean) - Mark as primary key column.
- `is_last_update_date` (Boolean) - Mark as last update date column. Required for incremental extraction to work correctly.
- `is_creation_date` (Boolean) - Mark as creation date column.
- `is_effective_start_date` (Boolean) - Mark as effective start date column.
- `is_natural_key` (Boolean) - Mark as natural key column.

---

### Nested Schema for `data_stores.columns`

Used when `auto_populate_all_columns = false` for explicit column selection.

#### Required

- `name` (String) - Column name.

#### Optional

- `is_populate` (Boolean) - Include this column in extraction. Default: `true`.
- `is_primary_key` (Boolean) - Mark as primary key column. Default: `false`.
- `is_last_update_date` (Boolean) - Mark as last update date column. Default: `false`.
- `is_creation_date` (Boolean) - Mark as creation date column. Default: `false`.
- `is_effective_start_date` (Boolean) - Mark as effective start date column. Default: `false`.
- `is_natural_key` (Boolean) - Mark as natural key column. Default: `false`.

## Import

BICC jobs can be imported using the job ID:

```shell
terraform import bicc_job.example 123456789
```
