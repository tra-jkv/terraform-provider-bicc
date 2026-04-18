package bicc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tra-jkv/terraform-provider-bicc/bicc/client"
)

var _ resource.Resource = &biccJobResource{}
var _ resource.ResourceWithImportState = &biccJobResource{}
var _ resource.ResourceWithModifyPlan = &biccJobResource{}

type biccJobModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	DataStores  types.Set    `tfsdk:"data_stores"`
}

type dataStoreModel struct {
	DataStoreKey            types.String `tfsdk:"data_store_key"`
	Filters                 types.String `tfsdk:"filters"`
	IsSilentError           types.Bool   `tfsdk:"is_silent_error"`
	IsEffectiveDateDisabled types.Bool   `tfsdk:"is_effective_date_disabled"`
	UseUnionForIncremental  types.Bool   `tfsdk:"use_union_for_incremental"`
	InitialExtractDate      types.String `tfsdk:"initial_extract_date"`
	ChunkType               types.String `tfsdk:"chunk_type"`
	ChunkDateSeqIncr        types.Int64  `tfsdk:"chunk_date_seq_incr"`
	ChunkDateSeqMin         types.Int64  `tfsdk:"chunk_date_seq_min"`
	ChunkPkSeqIncr          types.Int64  `tfsdk:"chunk_pk_seq_incr"`
	AutoPopulateAllColumns  types.Bool   `tfsdk:"auto_populate_all_columns"`
	ColumnOverrides         types.List   `tfsdk:"column_overrides"`
	Columns                 types.List   `tfsdk:"columns"`
}

type columnOverrideModel struct {
	Name                 types.String `tfsdk:"name"`
	IsPopulate           types.Bool   `tfsdk:"is_populate"`
	IsPrimaryKey         types.Bool   `tfsdk:"is_primary_key"`
	IsLastUpdateDate     types.Bool   `tfsdk:"is_last_update_date"`
	IsCreationDate       types.Bool   `tfsdk:"is_creation_date"`
	IsEffectiveStartDate types.Bool   `tfsdk:"is_effective_start_date"`
	IsNaturalKey         types.Bool   `tfsdk:"is_natural_key"`
}

type columnModel struct {
	Name                 types.String `tfsdk:"name"`
	IsPopulate           types.Bool   `tfsdk:"is_populate"`
	IsPrimaryKey         types.Bool   `tfsdk:"is_primary_key"`
	IsLastUpdateDate     types.Bool   `tfsdk:"is_last_update_date"`
	IsCreationDate       types.Bool   `tfsdk:"is_creation_date"`
	IsEffectiveStartDate types.Bool   `tfsdk:"is_effective_start_date"`
	IsNaturalKey         types.Bool   `tfsdk:"is_natural_key"`
}

var columnOverrideAttrTypes = map[string]attr.Type{
	"name":                    types.StringType,
	"is_populate":             types.BoolType,
	"is_primary_key":          types.BoolType,
	"is_last_update_date":     types.BoolType,
	"is_creation_date":        types.BoolType,
	"is_effective_start_date": types.BoolType,
	"is_natural_key":          types.BoolType,
}

var columnAttrTypes = map[string]attr.Type{
	"name":                    types.StringType,
	"is_populate":             types.BoolType,
	"is_primary_key":          types.BoolType,
	"is_last_update_date":     types.BoolType,
	"is_creation_date":        types.BoolType,
	"is_effective_start_date": types.BoolType,
	"is_natural_key":          types.BoolType,
}

var dataStoreAttrTypes = map[string]attr.Type{
	"data_store_key":             types.StringType,
	"filters":                    types.StringType,
	"is_silent_error":            types.BoolType,
	"is_effective_date_disabled": types.BoolType,
	"use_union_for_incremental":  types.BoolType,
	"initial_extract_date":       types.StringType,
	"chunk_type":                 types.StringType,
	"chunk_date_seq_incr":        types.Int64Type,
	"chunk_date_seq_min":         types.Int64Type,
	"chunk_pk_seq_incr":          types.Int64Type,
	"auto_populate_all_columns":  types.BoolType,
	"column_overrides":           types.ListType{ElemType: types.ObjectType{AttrTypes: columnOverrideAttrTypes}},
	"columns":                    types.ListType{ElemType: types.ObjectType{AttrTypes: columnAttrTypes}},
}

type biccJobResource struct {
	client *client.Client
}

func NewBICCJobResource() resource.Resource {
	return &biccJobResource{}
}

func (r *biccJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (r *biccJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	columnAttrs := map[string]schema.Attribute{
		"name":                    schema.StringAttribute{Required: true, Description: "Column name."},
		"is_populate":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Include this column in extraction."},
		"is_primary_key":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Mark as primary key column."},
		"is_last_update_date":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Mark as last update date column (required for incremental)."},
		"is_creation_date":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Mark as creation date column."},
		"is_effective_start_date": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Mark as effective start date column."},
		"is_natural_key":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Mark as natural key column."},
	}

	columnOverrideAttrs := map[string]schema.Attribute{
		"name":                    schema.StringAttribute{Required: true, Description: "Column name to override."},
		"is_populate":             schema.BoolAttribute{Optional: true, Description: "Override: include this column in extraction."},
		"is_primary_key":          schema.BoolAttribute{Optional: true, Description: "Override: mark as primary key column."},
		"is_last_update_date":     schema.BoolAttribute{Optional: true, Description: "Override: mark as last update date column (required for incremental)."},
		"is_creation_date":        schema.BoolAttribute{Optional: true, Description: "Override: mark as creation date column."},
		"is_effective_start_date": schema.BoolAttribute{Optional: true, Description: "Override: mark as effective start date column."},
		"is_natural_key":          schema.BoolAttribute{Optional: true, Description: "Override: mark as natural key column."},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a BICC extraction job.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The numeric ID of the BICC job.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the BICC job.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the BICC job.",
			},
			"data_stores": schema.SetNestedAttribute{
				Required:    true,
				Description: "Set of data stores to extract (order-independent, keyed by data_store_key).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"data_store_key": schema.StringAttribute{
							Required:    true,
							Description: "Unique key for the data store (e.g., CrmAnalyticsAM.PartiesAnalyticsAM.Person).",
						},
						"filters": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Filter expression for data extraction.",
						},
						"is_silent_error": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Continue extraction even if this data store fails.",
						},
						"is_effective_date_disabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Disable effective date filtering.",
						},
						"use_union_for_incremental": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Enable incremental extraction using UNION approach.",
						},
						"initial_extract_date": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Initial extract date for incremental extraction (format: YYYY-MM-DD).",
							PlanModifiers: []planmodifier.String{
								normalizeDatePlanModifier{},
							},
						},
						"chunk_type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Chunking type for large extractions (e.g., DATE, SEQUENCE).",
						},
						"chunk_date_seq_incr": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Description: "Date sequence increment for chunking.",
						},
						"chunk_date_seq_min": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Description: "Minimum date sequence for chunking.",
						},
						"chunk_pk_seq_incr": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Description: "Primary key sequence increment for chunking.",
						},
						"auto_populate_all_columns": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Automatically fetch and include all available columns from the data store. When true, define only column_overrides.",
						},
						"column_overrides": schema.ListNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Column configuration overrides (use with auto_populate_all_columns to mark specific columns like LastUpdateDate as incremental tracking).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: columnOverrideAttrs,
							},
						},
						"columns": schema.ListNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Explicit column configuration. Not needed when auto_populate_all_columns is true.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: columnAttrs,
							},
						},
					},
				},
			},
		},
	}
}

func (r *biccJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *biccJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan biccJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, diags := r.buildJobFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobResp, err := r.client.CreateOrUpdateJob(ctx, job)
	if err != nil {
		resp.Diagnostics.AddError("Error creating job", err.Error())
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(jobResp.ID, 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *biccJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state biccJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid job ID", err.Error())
		return
	}

	job, err := r.client.GetJob(ctx, jobID)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve config-only fields (auto_populate_all_columns, column_overrides) from state
	// by keying on data_store_key, since the API does not return them.
	oldByKey, diags := buildOldDataStoreMap(ctx, state.DataStores)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newDataStores, diags := r.buildDataStoresFromAPI(ctx, job.DataStores, oldByKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Name = types.StringValue(job.Name)
	state.Description = types.StringValue(job.Description)
	state.DataStores = newDataStores

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *biccJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan biccJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state biccJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	job, diags := r.buildJobFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CreateOrUpdateJob(ctx, job)
	if err != nil {
		resp.Diagnostics.AddError("Error updating job", err.Error())
		return
	}

	// Re-read from the API so state reflects server-normalized values.
	jobID, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid job ID after update", err.Error())
		return
	}

	refreshed, err := r.client.GetJob(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading job after update", err.Error())
		return
	}

	oldByKey, diags := buildOldDataStoreMap(ctx, plan.DataStores)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newDataStores, diags := r.buildDataStoresFromAPI(ctx, refreshed.DataStores, oldByKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Name = types.StringValue(refreshed.Name)
	plan.Description = types.StringValue(refreshed.Description)
	plan.DataStores = newDataStores

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *biccJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state biccJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid job ID", err.Error())
		return
	}

	if err := r.client.DeleteJob(ctx, jobID); err != nil {
		resp.Diagnostics.AddError("Error deleting job", err.Error())
	}
}

func (r *biccJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state biccJobModel
	state.ID = types.StringValue(req.ID)

	jobID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric job ID.")
		return
	}

	job, err := r.client.GetJob(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading job during import", err.Error())
		return
	}

	state.Name = types.StringValue(job.Name)
	state.Description = types.StringValue(job.Description)

	newDataStores, diags := r.buildDataStoresFromAPI(ctx, job.DataStores, map[string]dataStoreModel{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DataStores = newDataStores

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ModifyPlan suppresses a spurious post-import drift. After import the provider writes
// auto_populate_all_columns=false + full column list into state (no config context at
// import time). On the next plan, config says auto_populate_all_columns=true + columns=[],
// producing a diff with no real change. We detect this pattern and rewrite the plan element
// to match state, making plan == state and suppressing the diff. The correct state is written
// on the next apply via Read.
func (r *biccJobResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state biccJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan biccJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateDS []dataStoreModel
	resp.Diagnostics.Append(state.DataStores.ElementsAs(ctx, &stateDS, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stateByKey := make(map[string]dataStoreModel, len(stateDS))
	for _, ds := range stateDS {
		stateByKey[ds.DataStoreKey.ValueString()] = ds
	}

	var planDS []dataStoreModel
	resp.Diagnostics.Append(plan.DataStores.ElementsAs(ctx, &planDS, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	modified := false
	newElements := make([]attr.Value, len(planDS))
	for i, planElem := range planDS {
		key := planElem.DataStoreKey.ValueString()
		old, hasState := stateByKey[key]

		// Drift pattern: state has import-written explicit columns, plan wants auto-populate.
		needsFix := hasState &&
			planElem.AutoPopulateAllColumns.ValueBool() &&
			!old.AutoPopulateAllColumns.ValueBool() &&
			isListEmpty(planElem.Columns) &&
			!isListEmpty(old.Columns)

		var obj attr.Value
		var d diag.Diagnostics
		if needsFix {
			obj, d = types.ObjectValue(dataStoreAttrTypes, map[string]attr.Value{
				"data_store_key":             old.DataStoreKey,
				"filters":                    old.Filters,
				"is_silent_error":            old.IsSilentError,
				"is_effective_date_disabled": old.IsEffectiveDateDisabled,
				"use_union_for_incremental":  old.UseUnionForIncremental,
				"initial_extract_date":       old.InitialExtractDate,
				"chunk_type":                 old.ChunkType,
				"chunk_date_seq_incr":        old.ChunkDateSeqIncr,
				"chunk_date_seq_min":         old.ChunkDateSeqMin,
				"chunk_pk_seq_incr":          old.ChunkPkSeqIncr,
				"auto_populate_all_columns":  old.AutoPopulateAllColumns,
				"column_overrides":           old.ColumnOverrides,
				"columns":                    old.Columns,
			})
			modified = true
		} else if hasState {
			obj, d = types.ObjectValue(dataStoreAttrTypes, map[string]attr.Value{
				"data_store_key":             old.DataStoreKey,
				"filters":                    old.Filters,
				"is_silent_error":            old.IsSilentError,
				"is_effective_date_disabled": old.IsEffectiveDateDisabled,
				"use_union_for_incremental":  old.UseUnionForIncremental,
				"initial_extract_date":       old.InitialExtractDate,
				"chunk_type":                 old.ChunkType,
				"chunk_date_seq_incr":        old.ChunkDateSeqIncr,
				"chunk_date_seq_min":         old.ChunkDateSeqMin,
				"chunk_pk_seq_incr":          old.ChunkPkSeqIncr,
				"auto_populate_all_columns":  planElem.AutoPopulateAllColumns,
				"column_overrides":           planElem.ColumnOverrides,
				"columns":                    old.Columns,
			})
		} else {
			obj, d = types.ObjectValue(dataStoreAttrTypes, map[string]attr.Value{
				"data_store_key":             planElem.DataStoreKey,
				"filters":                    planElem.Filters,
				"is_silent_error":            planElem.IsSilentError,
				"is_effective_date_disabled": planElem.IsEffectiveDateDisabled,
				"use_union_for_incremental":  planElem.UseUnionForIncremental,
				"initial_extract_date":       planElem.InitialExtractDate,
				"chunk_type":                 planElem.ChunkType,
				"chunk_date_seq_incr":        planElem.ChunkDateSeqIncr,
				"chunk_date_seq_min":         planElem.ChunkDateSeqMin,
				"chunk_pk_seq_incr":          planElem.ChunkPkSeqIncr,
				"auto_populate_all_columns":  planElem.AutoPopulateAllColumns,
				"column_overrides":           planElem.ColumnOverrides,
				"columns":                    planElem.Columns,
			})
		}
		resp.Diagnostics.Append(d...)
		newElements[i] = obj
	}

	if resp.Diagnostics.HasError() || !modified {
		return
	}

	newSet, d := types.SetValue(types.ObjectType{AttrTypes: dataStoreAttrTypes}, newElements)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.DataStores = newSet
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func isListEmpty(l types.List) bool {
	return l.IsNull() || l.IsUnknown() || len(l.Elements()) == 0
}

func (r *biccJobResource) buildJobFromModel(ctx context.Context, model biccJobModel) (*client.Job, diag.Diagnostics) {
	var diags diag.Diagnostics

	job := &client.Job{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Schedules:   nil,
	}

	var dsModels []dataStoreModel
	diags.Append(model.DataStores.ElementsAs(ctx, &dsModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	dataStores := make([]client.DataStore, len(dsModels))
	for i, ds := range dsModels {
		var columns []client.Column

		if ds.AutoPopulateAllColumns.ValueBool() {
			allCols, err := r.client.GetDataStoreColumns(ctx, ds.DataStoreKey.ValueString())
			if err != nil {
				columns = []client.Column{}
			} else {
				columns = make([]client.Column, len(allCols))
				for j, col := range allCols {
					columns[j] = col.ToJobColumn()
					columns[j].IsPopulate = true
				}

				var overrides []columnOverrideModel
				diags.Append(ds.ColumnOverrides.ElementsAs(ctx, &overrides, false)...)
				if diags.HasError() {
					return nil, diags
				}
				overrideMap := make(map[string]columnOverrideModel, len(overrides))
				for _, o := range overrides {
					overrideMap[o.Name.ValueString()] = o
				}
				for j, col := range columns {
					if o, exists := overrideMap[col.Name]; exists {
						if !o.IsPopulate.IsNull() && !o.IsPopulate.IsUnknown() {
							columns[j].IsPopulate = o.IsPopulate.ValueBool()
						}
						if !o.IsPrimaryKey.IsNull() && !o.IsPrimaryKey.IsUnknown() {
							columns[j].IsPrimaryKey = o.IsPrimaryKey.ValueBool()
						}
						if !o.IsLastUpdateDate.IsNull() && !o.IsLastUpdateDate.IsUnknown() {
							columns[j].IsLastUpdateDate = o.IsLastUpdateDate.ValueBool()
						}
						if !o.IsCreationDate.IsNull() && !o.IsCreationDate.IsUnknown() {
							columns[j].IsCreationDate = o.IsCreationDate.ValueBool()
						}
						if !o.IsEffectiveStartDate.IsNull() && !o.IsEffectiveStartDate.IsUnknown() {
							columns[j].IsEffectiveStartDate = o.IsEffectiveStartDate.ValueBool()
						}
						if !o.IsNaturalKey.IsNull() && !o.IsNaturalKey.IsUnknown() {
							columns[j].IsNaturalKey = o.IsNaturalKey.ValueBool()
						}
					}
				}
			}
		} else {
			var colModels []columnModel
			diags.Append(ds.Columns.ElementsAs(ctx, &colModels, false)...)
			if diags.HasError() {
				return nil, diags
			}
			columns = make([]client.Column, len(colModels))
			for j, col := range colModels {
				columns[j] = client.Column{
					Name:                 col.Name.ValueString(),
					IsPopulate:           col.IsPopulate.ValueBool(),
					IsPrimaryKey:         col.IsPrimaryKey.ValueBool(),
					IsLastUpdateDate:     col.IsLastUpdateDate.ValueBool(),
					IsCreationDate:       col.IsCreationDate.ValueBool(),
					IsEffectiveStartDate: col.IsEffectiveStartDate.ValueBool(),
					IsNaturalKey:         col.IsNaturalKey.ValueBool(),
					ColConversion:        nil,
				}
			}
		}

		var initialExtractDate interface{}
		if !ds.InitialExtractDate.IsNull() && !ds.InitialExtractDate.IsUnknown() {
			dateStr := ds.InitialExtractDate.ValueString()
			if dateStr != "" {
				t, err := time.Parse("2006-01-02", strings.Split(dateStr, "T")[0])
				if err == nil {
					initialExtractDate = t.UTC().Format("2006-01-02T15:04:05.000Z")
				}
			}
		}

		// Send nil when chunk_type is unset so the API omits the field.
		var chunkType interface{}
		if !ds.ChunkType.IsNull() && !ds.ChunkType.IsUnknown() && ds.ChunkType.ValueString() != "" {
			chunkType = ds.ChunkType.ValueString()
		}

		dataStores[i] = client.DataStore{
			DataStoreMeta: client.DataStoreMeta{
				DataStoreKey:            ds.DataStoreKey.ValueString(),
				Filters:                 ds.Filters.ValueString(),
				IsSilentError:           ds.IsSilentError.ValueBool(),
				IsEffectiveDateDisabled: ds.IsEffectiveDateDisabled.ValueBool(),
				UseUnionForIncremental:  ds.UseUnionForIncremental.ValueBool(),
				InitialExtractDate:      initialExtractDate,
				ChunkType:               chunkType,
				ChunkDateSeqIncr:        int(ds.ChunkDateSeqIncr.ValueInt64()),
				ChunkDateSeqMin:         int(ds.ChunkDateSeqMin.ValueInt64()),
				ChunkPkSeqIncr:          int(ds.ChunkPkSeqIncr.ValueInt64()),
				Columns:                 columns,
			},
			GroupNumber:       0,
			GroupItemPriority: 0,
		}
	}

	job.DataStores = dataStores
	return job, diags
}

// buildOldDataStoreMap indexes existing state data stores by data_store_key.
func buildOldDataStoreMap(ctx context.Context, set types.Set) (map[string]dataStoreModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]dataStoreModel)

	if set.IsNull() || set.IsUnknown() {
		return result, diags
	}

	var dsModels []dataStoreModel
	diags.Append(set.ElementsAs(ctx, &dsModels, false)...)
	if diags.HasError() {
		return result, diags
	}

	for _, ds := range dsModels {
		result[ds.DataStoreKey.ValueString()] = ds
	}
	return result, diags
}

// buildDataStoresFromAPI converts API data stores into a types.Set.
// Config-only fields (auto_populate_all_columns, column_overrides) are carried
// forward from oldByKey since the API does not return them.
func (r *biccJobResource) buildDataStoresFromAPI(ctx context.Context, apiDS []client.DataStore, oldByKey map[string]dataStoreModel) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	dsObjects := make([]attr.Value, len(apiDS))

	for i, ds := range apiDS {
		key := ds.DataStoreMeta.DataStoreKey

		autoPopulate := false
		columnOverridesVal := types.ListValueMust(types.ObjectType{AttrTypes: columnOverrideAttrTypes}, []attr.Value{})
		if old, exists := oldByKey[key]; exists {
			autoPopulate = old.AutoPopulateAllColumns.ValueBool()
			columnOverridesVal = old.ColumnOverrides
		}

		initialExtractDate := types.StringNull()
		if ds.DataStoreMeta.InitialExtractDate != nil {
			if dateStr, ok := ds.DataStoreMeta.InitialExtractDate.(string); ok && dateStr != "" {
				initialExtractDate = types.StringValue(strings.Split(dateStr, "T")[0])
			}
		}

		chunkType := types.StringNull()
		if ds.DataStoreMeta.ChunkType != nil {
			if ct, ok := ds.DataStoreMeta.ChunkType.(string); ok && ct != "" {
				chunkType = types.StringValue(ct)
			}
		}

		// Omit columns from state when auto_populate is on to avoid drift from
		// API-returned column lists that the config doesn't manage explicitly.
		columnsVal := types.ListValueMust(types.ObjectType{AttrTypes: columnAttrTypes}, []attr.Value{})
		if !autoPopulate {
			colObjects := make([]attr.Value, len(ds.DataStoreMeta.Columns))
			for j, col := range ds.DataStoreMeta.Columns {
				colObj, d := types.ObjectValue(columnAttrTypes, map[string]attr.Value{
					"name":                    types.StringValue(col.Name),
					"is_populate":             types.BoolValue(col.IsPopulate),
					"is_primary_key":          types.BoolValue(col.IsPrimaryKey),
					"is_last_update_date":     types.BoolValue(col.IsLastUpdateDate),
					"is_creation_date":        types.BoolValue(col.IsCreationDate),
					"is_effective_start_date": types.BoolValue(col.IsEffectiveStartDate),
					"is_natural_key":          types.BoolValue(col.IsNaturalKey),
				})
				diags.Append(d...)
				colObjects[j] = colObj
			}
			var d diag.Diagnostics
			columnsVal, d = types.ListValue(types.ObjectType{AttrTypes: columnAttrTypes}, colObjects)
			diags.Append(d...)
		}

		dsObj, d := types.ObjectValue(dataStoreAttrTypes, map[string]attr.Value{
			"data_store_key":             types.StringValue(key),
			"filters":                    types.StringValue(ds.DataStoreMeta.Filters),
			"is_silent_error":            types.BoolValue(ds.DataStoreMeta.IsSilentError),
			"is_effective_date_disabled": types.BoolValue(ds.DataStoreMeta.IsEffectiveDateDisabled),
			"use_union_for_incremental":  types.BoolValue(ds.DataStoreMeta.UseUnionForIncremental),
			"initial_extract_date":       initialExtractDate,
			"chunk_type":                 chunkType,
			"chunk_date_seq_incr":        types.Int64Value(int64(ds.DataStoreMeta.ChunkDateSeqIncr)),
			"chunk_date_seq_min":         types.Int64Value(int64(ds.DataStoreMeta.ChunkDateSeqMin)),
			"chunk_pk_seq_incr":          types.Int64Value(int64(ds.DataStoreMeta.ChunkPkSeqIncr)),
			"auto_populate_all_columns":  types.BoolValue(autoPopulate),
			"column_overrides":           columnOverridesVal,
			"columns":                    columnsVal,
		})
		diags.Append(d...)
		dsObjects[i] = dsObj
	}

	result, d := types.SetValue(types.ObjectType{AttrTypes: dataStoreAttrTypes}, dsObjects)
	diags.Append(d...)
	return result, diags
}

// normalizeDatePlanModifier suppresses diffs caused by date format differences
// between config (YYYY-MM-DD) and the API response (ISO 8601).
type normalizeDatePlanModifier struct{}

func (m normalizeDatePlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs caused by date format differences between config (YYYY-MM-DD) and API response (ISO 8601)."
}

func (m normalizeDatePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeDatePlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}
	stateNorm := strings.Split(req.StateValue.ValueString(), "T")[0]
	planNorm := strings.Split(req.PlanValue.ValueString(), "T")[0]
	if stateNorm == planNorm {
		resp.PlanValue = req.StateValue
	}
}
