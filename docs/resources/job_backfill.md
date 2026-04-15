---
page_title: "bicc_job_backfill Resource - BICC Provider"
description: |-
  Manages backfilling for Oracle BICC extraction jobs by temporarily setting a last extract date.
---

# bicc_job_backfill (Resource)

Manages backfilling for BICC extraction jobs by setting a `last_extract_date` on specific data stores within a job. This causes the next job execution to re-extract data from that date forward.

## Why a Separate Resource?

The BICC API accepts `last_extract_date` when updating a job but does not persist or return it on subsequent reads. Including it in `bicc_job` would cause constant Terraform drift. This separate resource manages it independently with no drift detection.

## Workflow

1. Create a `bicc_job_backfill` resource with the desired backfill date(s).
2. Run `terraform apply` to set the backfill date on the job.
3. Trigger the BICC job execution via the external scheduler.
4. Once the backfill run completes, destroy the `bicc_job_backfill` resource — the main job is unaffected and continues its normal incremental extraction.

## Example Usage

### Backfill a Single Data Store

```terraform
resource "bicc_job_backfill" "supplier_backfill" {
  job_id = bicc_job.suppliers.id

  backfills {
    data_store_key    = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
    last_extract_date = "2024-01-01"
  }
}
```

### Backfill Multiple Data Stores

```terraform
resource "bicc_job_backfill" "billing_orders_backfill" {
  job_id = bicc_job.billing_orders.id

  backfills {
    data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.HeaderExtractPVO"
    last_extract_date = "2023-06-01"
  }

  backfills {
    data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.LineExtractPVO"
    last_extract_date = "2023-06-01"
  }

  backfills {
    data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.FulfillLineExtractPVO"
    last_extract_date = "2023-06-01"
  }
}
```

### Destroy After Backfill Completes

```shell
terraform destroy -target=bicc_job_backfill.billing_orders_backfill
```

## Schema

### Required

- `job_id` (String, Forces Replacement) - The ID of the BICC job to backfill.
- `backfills` (Block Set, Min: 1) - Set of backfill configurations. See [backfills](#nested-schema-for-backfills) below.

### Read-Only

- `id` (String) - Compound ID based on `job_id` and a hash of the data store keys.

---

### Nested Schema for `backfills`

#### Required

- `data_store_key` (String) - The data store key to backfill. Must match a data store key defined in the referenced `bicc_job`.
- `last_extract_date` (String) - The date from which to re-extract data. Format: `YYYY-MM-DD`.

## Behavior

| Operation  | Effect |
|------------|--------|
| **Create** | Sets `last_extract_date` on all specified data stores in the job |
| **Update** | Sets `last_extract_date` for new or changed entries — removing an entry from the set does not clear the date on the main job |
| **Delete** | No-op — the main job is not modified. `last_extract_date` is left as-is and continues to be managed by BICC after the job runs |
| **Read**   | Returns configured values (does not query API — no drift) |

## Import

Backfill resources are not importable as the API does not expose `last_extract_date`.
