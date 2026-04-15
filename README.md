# Terraform Provider for Oracle BICC

A Terraform provider for managing Oracle Business Intelligence Cloud Connector (BICC) extraction jobs.

## Features

- Create and manage BICC extraction jobs using REST API
- Support for incremental extraction
- Configurable chunking for large datasets
- Column-level extraction control
- Filter expressions for data extraction

## Background: Oracle BICC

[Business Intelligence Cloud Connector (BICC)](https://docs.oracle.com/en/cloud/saas/applications-common/25d/biacc/index.html) is an Oracle offering to extract data from Oracle Fusion Applications cloud data sources into an external object storage — either OCI (Oracle Cloud Infrastructure) Object Storage or UCM (Universal Content Management).

### Extract Schedules vs Job Schedules

BICC has two types of schedules:

**Extract schedule** — global configuration per object. Each object has a single last-run reference. On the next run, BICC checks when the last extract ran and pulls incremental data from that point. Because the last-run reference is global, you can only have one incremental cursor per object across the entire system.

**Job schedule** — scoped within the job itself. If the same object appears in two different jobs, each job maintains its own independent last-run reference. This can be used to extract data more frequently than BICC's built-in minimum interval (hourly), for example every 10 minutes — though this should be tested carefully as it may add load to the Fusion database and impact operational performance.

This provider manages **job-scoped** configuration (`bicc_job`), giving you fine-grained control over which objects are extracted together and their independent incremental cursors.

## Design: Scope of This Provider

### What this provider manages

This provider is scoped to BICC **job definitions** — what data to extract, which columns, and incremental extraction settings. It interacts with the BICC REST API to create, update, and delete jobs.

### Job execution and scheduling

Job execution is triggered by an external scheduler (e.g., GCP Cloud Scheduler, AWS EventBridge) via the Oracle ESS SOAP API's `submitRequest` operation, passing the OCI bucket as `EXTERNAL_STORAGE_LIST` at call time.

### Why not use Oracle's built-in BICC scheduling?

Oracle provides a [SOAP API](https://docs.oracle.com/en/cloud/saas/applications-common/25d/biacc/soap-api.html) that can schedule and trigger BICC jobs. However, during testing we found it unreliable and subject to a fundamental constraint:

**You cannot set both a recurring schedule frequency and an external OCI bucket destination at the same time.**

- If a job is configured with a **daily frequency** (scheduled type), the BICC API does not allow an external OCI Object Storage bucket to be set. The extract writes to internal UCM storage only.
- If a job is configured with an **external OCI bucket**, it becomes an **instant/on-demand** type — no recurring frequency can be set natively.

This means Oracle's built-in scheduling is not viable when you need extracts delivered to an external OCI bucket on a recurring schedule — which is the primary use case for data lake pipelines.

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for development)
- Oracle Fusion Applications instance with BICC enabled

## Prerequisites

For BICC concepts, configuration (external storage, data stores, jobs), and console navigation, refer to the [Oracle BICC documentation](https://docs.oracle.com/en/cloud/saas/applications-common/25d/biacc/index.html).

## Installation

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/tra-jkv/terraform-provider-bicc.git
cd terraform-provider-bicc
```

2. Build the provider:
```bash
go build -o terraform-provider-bicc
```

3. Install locally for development:
```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/tra-jkv/bicc/1.0.0/darwin_arm64
cp terraform-provider-bicc ~/.terraform.d/plugins/registry.terraform.io/tra-jkv/bicc/1.0.0/darwin_arm64/
```

Note: Adjust the path based on your OS and architecture (darwin_arm64, darwin_amd64, linux_amd64, etc.)

## Usage

### Provider Configuration

```hcl
terraform {
  required_providers {
    bicc = {
      source = "tra-jkv/bicc"
      version = "~> 1.0"
    }
  }
}

provider "bicc" {
  host     = "servername.fa.us2.oraclecloud.com"
  username = var.bicc_username
  password = var.bicc_password
  port     = 443  # Optional, defaults to 443
}
```

You can also use environment variables:
```bash
export BICC_HOST="servername.fa.us2.oraclecloud.com"
export BICC_USERNAME="your-username"
export BICC_PASSWORD="your-password"
```

### Creating a BICC Job

**Single Data Store Example:**

```hcl
resource "bicc_job" "crm_extract" {
  name        = "CRM_FULL_EXTRACT_JOB"
  description = "Full extract job for CRM Analytics data"

  data_stores {
    data_store_key              = "CrmAnalyticsAM.PartiesAnalyticsAM.Person"
    filters                     = "__DATASTORE__.CreationDate > '2024-01-01'"
    is_silent_error             = true
    is_effective_date_disabled  = false

    columns {
      name        = "PersonProfileId"
      is_populate = true
    }

    columns {
      name        = "PartyId"
      is_populate = true
    }
  }
}
```

**Multiple Data Stores in One Job (Recommended for Related Data):**

Group related data stores together for better organization. For example, billing orders with headers and lines:

```hcl
resource "bicc_job" "billing_orders" {
  name        = "BillingOrders"
  description = "Complete billing order lifecycle - headers, lines, fulfillment"

  # Sales Order Headers
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.HeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2024-01-01"
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Sales Order Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.LineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2024-01-01"
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Fulfillment Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.FulfillLineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2024-01-01"
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}
```

**Benefits of Multiple Data Stores per Job:**
- Related data extracts together (e.g., headers + lines)
- Single job execution extracts all related data
- Easier management in external schedulers
- Maintains data consistency across related entities

### Creating a Job with Incremental Extraction

```hcl
resource "bicc_job" "supplier_incremental" {
  name        = "SupplierIncrementalExtract"
  description = "Incremental extract for Supplier data"

  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    
    # Enable incremental extraction
    use_union_for_incremental = true
    initial_extract_date      = "2025-01-01"  # Optional: omit for all historical data
    
    # Optional: Enable chunking for large datasets
    chunk_type          = "DATE"
    chunk_date_seq_incr = 7  # Extract in 7-day chunks
  }
}
```

### Backfilling Data (Setting Last Extract Date)

When you need to backfill historical data for an incremental extraction job, use the `bicc_job_backfill` resource to temporarily set a `last_extract_date`. This is a separate resource because the BICC API accepts this field but doesn't persist or return it, which would otherwise cause Terraform drift.

**Key Features:**
- Uses `data_store_key` (not array index) for robust identification
- Supports backfilling multiple data stores in a single resource
- No drift detection - stable state management
- Easy to add/remove backfills without affecting the main job

**Workflow:**
1. Create the `bicc_job_backfill` resource with your desired backfill date(s)
2. Run the BICC job to extract data from that date forward
3. Destroy the `bicc_job_backfill` resource when backfilling is complete

**Example - Single Data Store:**

```hcl
# Main incremental job
resource "bicc_job" "supplier_incremental" {
  name        = "SupplierIncrementalExtract"
  description = "Incremental extract for Supplier data"

  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2025-01-01"
    
    auto_populate_all_columns = true
    
    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true  # Required for incremental
    }
  }
}

# Backfill resource - create when needed, destroy when done
resource "bicc_job_backfill" "supplier_backfill" {
  job_id = bicc_job.supplier_incremental.id
  
  backfills {
    data_store_key    = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
    last_extract_date = "2024-06-01"  # Backfill from this date
  }
}
```

**Example - Multiple Data Stores (Billing Orders):**

```hcl
# Job with multiple related data stores
resource "bicc_job" "billing_orders" {
  name        = "BillingOrders"
  description = "Complete billing order lifecycle"
  
  # Sales Order Headers
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.HeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2024-01-01"
    auto_populate_all_columns  = true
    
    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
  
  # Sales Order Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.LineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    initial_extract_date       = "2024-01-01"
    auto_populate_all_columns  = true
    
    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
  
  # Add more related data stores as needed...
}

# Backfill multiple data stores in the job
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
}
```

**To apply backfill:**
```bash
terraform apply
# Run your BICC job via scheduler or ESS API
```

**To update a backfill date:**
```bash
# Just change the date in your config and apply
terraform apply
```

**To remove backfill (after job completes):**
```bash
terraform destroy -target=bicc_job_backfill.supplier_backfill
# OR comment out the resource and run: terraform apply
```

The backfill resource won't cause drift - you can run `terraform plan` multiple times and it will show no changes.

### Job Scheduling

As described in the [Design](#design-why-terraform-for-jobs-but-external-scheduler-for-execution) section, job execution is triggered externally via the ESS SOAP API. Any scheduler that can make an HTTP request can be used:

- GCP Cloud Scheduler
- AWS EventBridge
- Azure Logic Apps
- Cron jobs

## Resource Reference

For full resource documentation, see the [Terraform Registry docs](https://registry.terraform.io/providers/tra-jkv/bicc/latest/docs) once published, or the `docs/` folder in this repository.


## Development

### Building

```bash
go build -o terraform-provider-bicc
```

### Testing

```bash
go test ./...
```

### Running Examples

```bash
cd examples
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init
terraform plan
terraform apply
```

## API Documentation

This provider uses:
- **REST API** for job management: [Oracle BICC REST API](https://docs.oracle.com/en/cloud/saas/applications-common/25d/biacc/manage-meta-data-rest-api1-vo-attributes-for-sdm-vo-and-non-sdm.html)

For job execution (handled separately):
- **SOAP API** for job execution: [Oracle BICC SOAP API](https://docs.oracle.com/en/cloud/saas/applications-common/25d/biacc/soap-api.html)

## Limitations

- Job scheduling must be handled externally (e.g., GCP Cloud Scheduler, AWS EventBridge, cron, etc.)

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License

## Author

Tra Nguyen

## Acknowledgments

- Oracle BICC Documentation
- HashiCorp Terraform Plugin SDK
