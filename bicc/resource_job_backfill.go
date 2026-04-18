package bicc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tra-jkv/terraform-provider-bicc/bicc/client"
)

var _ resource.Resource = &biccJobBackfillResource{}

type biccJobBackfillModel struct {
	ID        types.String `tfsdk:"id"`
	JobID     types.String `tfsdk:"job_id"`
	Backfills types.Set    `tfsdk:"backfills"`
}

type backfillEntryModel struct {
	DataStoreKey    types.String `tfsdk:"data_store_key"`
	LastExtractDate types.String `tfsdk:"last_extract_date"`
}

var backfillEntryAttrTypes = map[string]attr.Type{
	"data_store_key":    types.StringType,
	"last_extract_date": types.StringType,
}

type biccJobBackfillResource struct {
	client *client.Client
}

func NewBICCJobBackfillResource() resource.Resource {
	return &biccJobBackfillResource{}
}

func (r *biccJobBackfillResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_backfill"
}

func (r *biccJobBackfillResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets a last_extract_date on one or more data stores within a BICC job to trigger a backfill. " +
			"The BICC API accepts this field but does not return it, so this resource uses a no-op Read. " +
			"Destroy this resource once the backfill job run completes — BICC then manages the cursor itself.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Stable resource ID derived from job_id and data_store_keys.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the BICC job to backfill.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"backfills": schema.SetNestedAttribute{
				Required:    true,
				Description: "Set of data store backfill configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"data_store_key": schema.StringAttribute{
							Required:    true,
							Description: "The data store key to backfill.",
						},
						"last_extract_date": schema.StringAttribute{
							Required:    true,
							Description: "The last extract date for backfilling (format: YYYY-MM-DD).",
						},
					},
				},
			},
		},
	}
}

func (r *biccJobBackfillResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *biccJobBackfillResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan biccJobBackfillModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := strconv.ParseInt(plan.JobID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid job_id", err.Error())
		return
	}

	var entries []backfillEntryModel
	resp.Diagnostics.Append(plan.Backfills.ElementsAs(ctx, &entries, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyBackfills(jobID, entries); err != nil {
		resp.Diagnostics.AddError("Error applying backfills", err.Error())
		return
	}

	plan.ID = types.StringValue(generateBackfillID(jobID, entries))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read is a no-op: the BICC API never returns last_extract_date.
func (r *biccJobBackfillResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *biccJobBackfillResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan biccJobBackfillModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state biccJobBackfillModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	jobID, err := strconv.ParseInt(plan.JobID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid job_id", err.Error())
		return
	}

	var entries []backfillEntryModel
	resp.Diagnostics.Append(plan.Backfills.ElementsAs(ctx, &entries, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyBackfills(jobID, entries); err != nil {
		resp.Diagnostics.AddError("Error applying backfills", err.Error())
		return
	}

	plan.ID = types.StringValue(generateBackfillID(jobID, entries))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete is a no-op: destroying this resource does not reset the job's extract cursor.
// Once the backfill run completes, BICC manages the cursor itself.
func (r *biccJobBackfillResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *biccJobBackfillResource) applyBackfills(jobID int64, entries []backfillEntryModel) error {
	job, err := r.client.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	for _, entry := range entries {
		dataStoreKey := entry.DataStoreKey.ValueString()
		isoDate, err := convertToISO8601(entry.LastExtractDate.ValueString())
		if err != nil {
			return fmt.Errorf("invalid last_extract_date %q for data_store_key %q: %w", entry.LastExtractDate.ValueString(), dataStoreKey, err)
		}

		found := false
		for i, ds := range job.DataStores {
			if ds.DataStoreMeta.DataStoreKey == dataStoreKey {
				job.DataStores[i].LastExtractDate = isoDate
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("data_store_key %q not found in job %d", dataStoreKey, jobID)
		}
	}

	_, err = r.client.CreateOrUpdateJob(job)
	return err
}

// convertToISO8601 parses a YYYY-MM-DD date (or an existing ISO 8601 string) and
// returns it in "2006-01-02T15:04:05.000Z" format. Returns an error on invalid input.
func convertToISO8601(dateStr string) (string, error) {
	datePart := strings.Split(dateStr, "T")[0]
	t, err := time.Parse("2006-01-02", datePart)
	if err != nil {
		return "", fmt.Errorf("expected YYYY-MM-DD format, got %q: %w", dateStr, err)
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z"), nil
}

func generateBackfillID(jobID int64, entries []backfillEntryModel) string {
	keys := make([]string, len(entries))
	for i, e := range entries {
		keys[i] = e.DataStoreKey.ValueString()
	}
	sort.Strings(keys)

	h := sha256.New()
	h.Write([]byte(strings.Join(keys, "|")))
	return fmt.Sprintf("%d:%x", jobID, h.Sum(nil)[:8])
}
