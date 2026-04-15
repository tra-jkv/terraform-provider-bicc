package bicc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tra-jkv/terraform-provider-bicc/bicc/client"
)

func resourceBICCJobBackfill() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBICCJobBackfillCreate,
		ReadContext:   resourceBICCJobBackfillRead,
		UpdateContext: resourceBICCJobBackfillUpdate,
		DeleteContext: resourceBICCJobBackfillDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"job_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the BICC job to backfill",
			},
			"backfills": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Set of backfill configurations for data stores",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_store_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The data store key to backfill (e.g., 'FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO')",
						},
						"last_extract_date": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The last extract date for backfilling (format: YYYY-MM-DD)",
						},
					},
				},
				Set: func(v interface{}) int {
					// Hash based on both data_store_key and last_extract_date
					m := v.(map[string]interface{})
					key := m["data_store_key"].(string)
					date := m["last_extract_date"].(string)
					return schema.HashString(key + "|" + date)
				},
			},
		},
	}
}

func resourceBICCJobBackfillCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	var diags diag.Diagnostics

	jobID, err := strconv.ParseInt(d.Get("job_id").(string), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid job_id: %v", err))
	}

	backfills := d.Get("backfills").(*schema.Set).List()

	// Get the existing job
	job, err := c.GetJob(jobID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get job: %v", err))
	}

	// Apply each backfill
	for _, bf := range backfills {
		backfillMap := bf.(map[string]interface{})
		dataStoreKey := backfillMap["data_store_key"].(string)
		lastExtractDate := backfillMap["last_extract_date"].(string)

		// Find the data store index by key
		found := false
		for i, ds := range job.DataStores {
			if ds.DataStoreMeta.DataStoreKey == dataStoreKey {
				// Convert date format if needed (YYYY-MM-DD -> ISO 8601)
				isoDate := convertToISO8601(lastExtractDate)
				job.DataStores[i].LastExtractDate = isoDate
				found = true
				break
			}
		}

		if !found {
			return diag.FromErr(fmt.Errorf("data_store_key '%s' not found in job %d", dataStoreKey, jobID))
		}
	}

	// Update the job with all backfills
	_, err = c.CreateOrUpdateJob(job)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update job with backfill dates: %v", err))
	}

	// Generate a stable ID based on job_id and data_store_keys
	d.SetId(generateBackfillID(jobID, backfills))

	return diags
}

func resourceBICCJobBackfillRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Since the API doesn't return last_extract_date, we simply preserve what's in config
	// The resource ID confirms this backfill resource exists

	return diags
}

func resourceBICCJobBackfillUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	var diags diag.Diagnostics

	if d.HasChange("backfills") {
		jobID, err := strconv.ParseInt(d.Get("job_id").(string), 10, 64)
		if err != nil {
			return diag.FromErr(fmt.Errorf("invalid job_id: %v", err))
		}

		// Get the existing job
		job, err := c.GetJob(jobID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to get job: %v", err))
		}

		_, new := d.GetChange("backfills")
		newBackfills := new.(*schema.Set)

		// Only apply new/changed backfills - never clear last_extract_date on removed entries
		// as that would interfere with the main job's incremental cursor managed by BICC
		for _, bf := range newBackfills.List() {
			backfillMap := bf.(map[string]interface{})
			dataStoreKey := backfillMap["data_store_key"].(string)
			lastExtractDate := backfillMap["last_extract_date"].(string)

			found := false
			for i, ds := range job.DataStores {
				if ds.DataStoreMeta.DataStoreKey == dataStoreKey {
					isoDate := convertToISO8601(lastExtractDate)
					job.DataStores[i].LastExtractDate = isoDate
					found = true
					break
				}
			}

			if !found {
				return diag.FromErr(fmt.Errorf("data_store_key '%s' not found in job %d", dataStoreKey, jobID))
			}
		}

		// Update the job
		_, err = c.CreateOrUpdateJob(job)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update job with new backfill dates: %v", err))
		}

		// Update resource ID
		d.SetId(generateBackfillID(jobID, newBackfills.List()))
	}

	return diags
}

func resourceBICCJobBackfillDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// No-op: deleting a backfill resource does not modify the main job.
	// The backfill is an ad-hoc operation — once the job has run, last_extract_date
	// is managed by BICC itself and should not be reset on destroy.
	return nil
}

// Helper function to convert YYYY-MM-DD to ISO 8601 format
func convertToISO8601(dateStr string) string {
	// If already in ISO 8601 format, return as-is
	if strings.Contains(dateStr, "T") {
		return dateStr
	}

	// Parse YYYY-MM-DD format
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// If parsing fails, return as-is and let API handle it
		return dateStr
	}

	// Convert to ISO 8601 with UTC timezone
	return t.Format("2006-01-02T15:04:05.000Z")
}

// Generate a stable resource ID based on job_id and data_store_keys
func generateBackfillID(jobID int64, backfills []interface{}) string {
	// Collect and sort data store keys for consistent ID generation
	var keys []string
	for _, bf := range backfills {
		backfillMap := bf.(map[string]interface{})
		keys = append(keys, backfillMap["data_store_key"].(string))
	}
	sort.Strings(keys)

	// Create a hash of the sorted keys
	h := sha256.New()
	h.Write([]byte(strings.Join(keys, "|")))
	hash := fmt.Sprintf("%x", h.Sum(nil))[:16] // Use first 16 chars of hash

	return fmt.Sprintf("%d:%s", jobID, hash)
}
