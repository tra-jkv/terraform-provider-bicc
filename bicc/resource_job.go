package bicc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tra-jkv/terraform-provider-bicc/bicc/client"
)

// dataStoreHash hashes a data_stores block on its data_store_key so that
// TypeSet treats the set as unordered — reordering data_stores in config or
// API response will no longer produce a spurious diff.
func dataStoreHash(v interface{}) int {
	m := v.(map[string]interface{})
	return schema.HashString(m["data_store_key"].(string))
}

func resourceBICCJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBICCJobCreate,
		ReadContext:   resourceBICCJobRead,
		UpdateContext: resourceBICCJobUpdate,
		DeleteContext: resourceBICCJobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the BICC job",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the BICC job",
			},
			"data_stores": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Set of data stores to extract (order-independent, keyed by data_store_key)",
				Set:         dataStoreHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_store_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique key for the data store (e.g., CrmAnalyticsAM.PartiesAnalyticsAM.Person)",
						},
						"filters": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter expression for data extraction",
						},
						"is_silent_error": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Continue extraction even if this data store fails",
						},
						"is_effective_date_disabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Disable effective date filtering",
						},
						"use_union_for_incremental": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable incremental extraction using UNION approach",
						},
						"initial_extract_date": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Normalize both dates to compare (handle format differences)
								// API returns ISO 8601, user provides YYYY-MM-DD
								oldNorm := strings.Split(old, "T")[0] // "2025-01-01T00:00:00.000" -> "2025-01-01"
								newNorm := strings.Split(new, "T")[0] // Already "2025-01-01"
								return oldNorm == newNorm
							},
							Description: "Initial extract date for incremental extraction (format: YYYY-MM-DD)",
						},
						"chunk_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Chunking type for large extractions (e.g., DATE, SEQUENCE)",
						},
						"chunk_date_seq_incr": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Date sequence increment for chunking",
						},
						"chunk_date_seq_min": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Minimum date sequence for chunking",
						},
						"chunk_pk_seq_incr": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Primary key sequence increment for chunking",
						},
						"auto_populate_all_columns": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Automatically fetch and include all available columns from the data store. When true, you only need to define column overrides (e.g., marking LastUpdateDate as incremental tracking column)",
						},
						"column_overrides": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Column configuration overrides (use with auto_populate_all_columns to mark specific columns like LastUpdateDate as incremental tracking)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Column name",
									},
									"is_populate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Override: Include this column in extraction",
									},
									"is_primary_key": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Override: Mark as primary key column",
									},
									"is_last_update_date": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Override: Mark as last update date column (required for incremental)",
									},
									"is_creation_date": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Override: Mark as creation date column",
									},
									"is_effective_start_date": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Override: Mark as effective start date column",
									},
									"is_natural_key": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Override: Mark as natural key column",
									},
								},
							},
						},
						"columns": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Column configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Column name",
									},
									"is_populate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "Include this column in extraction",
									},
									"is_primary_key": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Mark as primary key column",
									},
									"is_last_update_date": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Mark as last update date column (required for incremental)",
									},
									"is_creation_date": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Mark as creation date column",
									},
									"is_effective_start_date": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Mark as effective start date column",
									},
									"is_natural_key": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Mark as natural key column",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceBICCJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	var diags diag.Diagnostics

	job := buildJobFromResourceData(d, c)

	resp, err := c.CreateOrUpdateJob(job)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating job: %v", err))
	}

	d.SetId(strconv.FormatInt(resp.ID, 10))

	return diags
}

func resourceBICCJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	var diags diag.Diagnostics

	jobID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid job ID: %v", err))
	}

	job, err := c.GetJob(jobID)
	if err != nil {
		// If job not found, remove from state
		d.SetId("")
		return diags
	}

	d.Set("name", job.Name)
	d.Set("description", job.Description)

	// Build a lookup map from the existing state set so we can retrieve
	// config-only fields (auto_populate_all_columns, column_overrides) by
	// data_store_key rather than by index.
	oldSet := d.Get("data_stores").(*schema.Set)
	oldByKey := make(map[string]map[string]interface{}, oldSet.Len())
	for _, raw := range oldSet.List() {
		if oldDS, ok := raw.(map[string]interface{}); ok {
			if key, ok := oldDS["data_store_key"].(string); ok {
				oldByKey[key] = oldDS
			}
		}
	}

	// Build the new set of data store maps from the API response.
	dataStores := make([]interface{}, len(job.DataStores))

	for i, ds := range job.DataStores {
		dataStore := make(map[string]interface{})
		key := ds.DataStoreMeta.DataStoreKey
		dataStore["data_store_key"] = key
		dataStore["filters"] = ds.DataStoreMeta.Filters
		dataStore["is_silent_error"] = ds.DataStoreMeta.IsSilentError
		dataStore["is_effective_date_disabled"] = ds.DataStoreMeta.IsEffectiveDateDisabled
		dataStore["use_union_for_incremental"] = ds.DataStoreMeta.UseUnionForIncremental

		// Preserve config-only fields by looking up the old entry by key.
		autoPopulateEnabled := false
		if oldDS, exists := oldByKey[key]; exists {
			if autoPopulate, ok := oldDS["auto_populate_all_columns"]; ok {
				dataStore["auto_populate_all_columns"] = autoPopulate
				if autoPop, ok := autoPopulate.(bool); ok {
					autoPopulateEnabled = autoPop
				}
			}
			if columnOverrides, ok := oldDS["column_overrides"]; ok {
				dataStore["column_overrides"] = columnOverrides
			}
		}

		// Handle initial extract date
		if ds.DataStoreMeta.InitialExtractDate != nil {
			if dateStr, ok := ds.DataStoreMeta.InitialExtractDate.(string); ok {
				dataStore["initial_extract_date"] = dateStr
			}
		}

		// Handle chunk type
		if ds.DataStoreMeta.ChunkType != nil {
			if chunkStr, ok := ds.DataStoreMeta.ChunkType.(string); ok {
				dataStore["chunk_type"] = chunkStr
			}
		}

		dataStore["chunk_date_seq_incr"] = ds.DataStoreMeta.ChunkDateSeqIncr
		dataStore["chunk_date_seq_min"] = ds.DataStoreMeta.ChunkDateSeqMin
		dataStore["chunk_pk_seq_incr"] = ds.DataStoreMeta.ChunkPkSeqIncr

		// When auto-populate is enabled, don't store columns in state
		// to avoid drift from API-returned column lists.
		if autoPopulateEnabled {
			dataStore["columns"] = []interface{}{}
			dataStores[i] = dataStore
			continue
		}

		// Set columns from API response only when NOT using auto_populate_all_columns.
		columns := make([]interface{}, 0)
		for _, col := range ds.DataStoreMeta.Columns {
			column := make(map[string]interface{})
			column["name"] = col.Name
			column["is_populate"] = col.IsPopulate
			column["is_primary_key"] = col.IsPrimaryKey
			column["is_last_update_date"] = col.IsLastUpdateDate
			column["is_creation_date"] = col.IsCreationDate
			column["is_effective_start_date"] = col.IsEffectiveStartDate
			column["is_natural_key"] = col.IsNaturalKey
			columns = append(columns, column)
		}
		dataStore["columns"] = columns

		dataStores[i] = dataStore
	}
	d.Set("data_stores", dataStores)

	return diags
}

func resourceBICCJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	if d.HasChanges("name", "description", "data_stores") {
		job := buildJobFromResourceData(d, c)

		_, err := c.CreateOrUpdateJob(job)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating job: %v", err))
		}
	}

	return resourceBICCJobRead(ctx, d, meta)
}

func resourceBICCJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	var diags diag.Diagnostics

	jobID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid job ID: %v", err))
	}

	err = c.DeleteJob(jobID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting job: %v", err))
	}

	d.SetId("")

	return diags
}

func buildJobFromResourceData(d *schema.ResourceData, c *client.Client) *client.Job {
	job := &client.Job{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Schedules:   nil,
	}

	dataStoresSet := d.Get("data_stores").(*schema.Set)
	dataStoresData := dataStoresSet.List()
	dataStores := make([]client.DataStore, len(dataStoresData))

	for i, dsData := range dataStoresData {
		dsMap := dsData.(map[string]interface{})

		var columns []client.Column
		autoPopulate := dsMap["auto_populate_all_columns"].(bool)

		if autoPopulate {
			// Fetch all columns from the data store
			dataStoreKey := dsMap["data_store_key"].(string)
			allColumns, err := c.GetDataStoreColumns(dataStoreKey)
			if err != nil {
				// If we can't fetch columns, fall back to manual definition
				columns = make([]client.Column, 0)
			} else {
				// Set all columns to be populated by default
				// Clean columns by removing metadata fields (label, dataType, etc.)
				columns = make([]client.Column, len(allColumns))
				for i, col := range allColumns {
					columns[i] = col.ToJobColumn() // Strip metadata fields
					columns[i].IsPopulate = true   // Ensure all columns are included by default
				}

				// Apply column overrides
				if overridesData, ok := dsMap["column_overrides"].([]interface{}); ok && len(overridesData) > 0 {
					overridesMap := make(map[string]map[string]interface{})
					for _, overrideData := range overridesData {
						overrideMap := overrideData.(map[string]interface{})
						name := overrideMap["name"].(string)
						overridesMap[name] = overrideMap
					}

					// Apply overrides to matching columns
					for j, col := range columns {
						if override, exists := overridesMap[col.Name]; exists {
							if val, ok := override["is_populate"].(bool); ok {
								columns[j].IsPopulate = val
							}
							if val, ok := override["is_primary_key"].(bool); ok {
								columns[j].IsPrimaryKey = val
							}
							if val, ok := override["is_last_update_date"].(bool); ok {
								columns[j].IsLastUpdateDate = val
							}
							if val, ok := override["is_creation_date"].(bool); ok {
								columns[j].IsCreationDate = val
							}
							if val, ok := override["is_effective_start_date"].(bool); ok {
								columns[j].IsEffectiveStartDate = val
							}
							if val, ok := override["is_natural_key"].(bool); ok {
								columns[j].IsNaturalKey = val
							}
						}
					}
				}
			}
		} else {
			// Use manually defined columns
			columns = make([]client.Column, 0)
			if columnsData, ok := dsMap["columns"].([]interface{}); ok {
				for _, colData := range columnsData {
					colMap := colData.(map[string]interface{})
					columns = append(columns, client.Column{
						Name:                 colMap["name"].(string),
						IsPopulate:           colMap["is_populate"].(bool),
						IsPrimaryKey:         colMap["is_primary_key"].(bool),
						IsLastUpdateDate:     colMap["is_last_update_date"].(bool),
						IsCreationDate:       colMap["is_creation_date"].(bool),
						IsEffectiveStartDate: colMap["is_effective_start_date"].(bool),
						IsNaturalKey:         colMap["is_natural_key"].(bool),
						ColConversion:        nil,
					})
				}
			}
		}

		// Convert initial_extract_date from YYYY-MM-DD to Oracle format
		initialExtractDate := getInterfaceValue(dsMap, "initial_extract_date")
		if dateStr, ok := initialExtractDate.(string); ok && dateStr != "" {
			// Parse YYYY-MM-DD format and convert to ISO 8601 with timezone
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				// Convert to ISO 8601 format with UTC timezone: 2025-01-01T00:00:00.000Z
				initialExtractDate = t.UTC().Format("2006-01-02T15:04:05.000Z")
			} else {
				// If parsing fails, set to nil
				initialExtractDate = nil
			}
		} else {
			// If not set or empty, use nil instead of empty string
			initialExtractDate = nil
		}

		// Get chunk type, use nil if empty
		chunkType := getInterfaceValue(dsMap, "chunk_type")
		if chunkTypeStr, ok := chunkType.(string); ok && chunkTypeStr == "" {
			chunkType = nil
		}

		// Get filters, use empty string if not set (filters can be empty)
		filters := ""
		if filterVal, ok := dsMap["filters"].(string); ok {
			filters = filterVal
		}

		dataStores[i] = client.DataStore{
			DataStoreMeta: client.DataStoreMeta{
				DataStoreKey:            dsMap["data_store_key"].(string),
				Filters:                 filters,
				IsSilentError:           dsMap["is_silent_error"].(bool),
				IsEffectiveDateDisabled: dsMap["is_effective_date_disabled"].(bool),
				UseUnionForIncremental:  dsMap["use_union_for_incremental"].(bool),
				InitialExtractDate:      initialExtractDate,
				ChunkType:               chunkType,
				ChunkDateSeqIncr:        dsMap["chunk_date_seq_incr"].(int),
				ChunkDateSeqMin:         dsMap["chunk_date_seq_min"].(int),
				ChunkPkSeqIncr:          dsMap["chunk_pk_seq_incr"].(int),
				Columns:                 columns,
			},
			GroupNumber:       0,
			GroupItemPriority: 0,
		}
	}

	job.DataStores = dataStores

	return job
}

// getInterfaceValue returns the value from map if it exists and is not empty, otherwise nil
func getInterfaceValue(m map[string]interface{}, key string) interface{} {
	if val, ok := m[key]; ok {
		if strVal, isString := val.(string); isString && strVal != "" {
			return strVal
		} else if val != nil {
			return val
		}
	}
	return nil
}
