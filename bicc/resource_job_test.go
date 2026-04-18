package bicc

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// -------------------------------------------------------------------------
// isListEmpty
// -------------------------------------------------------------------------

func TestIsListEmpty(t *testing.T) {
	elemType := types.ObjectType{AttrTypes: columnAttrTypes}

	t.Run("null list", func(t *testing.T) {
		if !isListEmpty(types.ListNull(elemType)) {
			t.Error("null list should be empty")
		}
	})

	t.Run("unknown list", func(t *testing.T) {
		if !isListEmpty(types.ListUnknown(elemType)) {
			t.Error("unknown list should be empty")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		l := types.ListValueMust(elemType, []attr.Value{})
		if !isListEmpty(l) {
			t.Error("zero-element list should be empty")
		}
	})

	t.Run("non-empty list", func(t *testing.T) {
		obj, _ := types.ObjectValue(columnAttrTypes, map[string]attr.Value{
			"name":                    types.StringValue("C1"),
			"is_populate":             types.BoolValue(true),
			"is_primary_key":          types.BoolValue(false),
			"is_last_update_date":     types.BoolValue(false),
			"is_creation_date":        types.BoolValue(false),
			"is_effective_start_date": types.BoolValue(false),
			"is_natural_key":          types.BoolValue(false),
		})
		l := types.ListValueMust(elemType, []attr.Value{obj})
		if isListEmpty(l) {
			t.Error("one-element list should not be empty")
		}
	})
}

// -------------------------------------------------------------------------
// normalizeDatePlanModifier
// -------------------------------------------------------------------------

func TestNormalizeDatePlanModifier(t *testing.T) {
	m := normalizeDatePlanModifier{}

	applyMod := func(stateVal, planVal string) string {
		req := planmodifier.StringRequest{
			StateValue: types.StringValue(stateVal),
			PlanValue:  types.StringValue(planVal),
		}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		m.PlanModifyString(context.Background(), req, resp)
		return resp.PlanValue.ValueString()
	}

	t.Run("same date different format — suppressed to state value", func(t *testing.T) {
		// state has ISO, plan has plain date
		got := applyMod("2025-01-15T00:00:00.000Z", "2025-01-15")
		if got != "2025-01-15T00:00:00.000Z" {
			t.Errorf("expected state value to win, got %q", got)
		}
	})

	t.Run("different dates — plan value kept", func(t *testing.T) {
		got := applyMod("2025-01-15T00:00:00.000Z", "2025-02-01")
		if got != "2025-02-01" {
			t.Errorf("expected plan value to stay, got %q", got)
		}
	})

	t.Run("both plain YYYY-MM-DD equal — suppressed", func(t *testing.T) {
		got := applyMod("2025-06-01", "2025-06-01")
		if got != "2025-06-01" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("null state — plan value kept unchanged", func(t *testing.T) {
		req := planmodifier.StringRequest{
			StateValue: types.StringNull(),
			PlanValue:  types.StringValue("2025-01-01"),
		}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		m.PlanModifyString(context.Background(), req, resp)
		if resp.PlanValue.ValueString() != "2025-01-01" {
			t.Errorf("expected plan value unchanged, got %q", resp.PlanValue.ValueString())
		}
	})

	t.Run("null plan — plan value kept unchanged", func(t *testing.T) {
		req := planmodifier.StringRequest{
			StateValue: types.StringValue("2025-01-01"),
			PlanValue:  types.StringNull(),
		}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		m.PlanModifyString(context.Background(), req, resp)
		if !resp.PlanValue.IsNull() {
			t.Errorf("expected null plan value, got %q", resp.PlanValue.ValueString())
		}
	})
}

// -------------------------------------------------------------------------
// buildOldDataStoreMap
// -------------------------------------------------------------------------

func makeDataStoreSet(t *testing.T, keys ...string) types.Set {
	t.Helper()
	objs := make([]attr.Value, len(keys))
	for i, k := range keys {
		obj, d := types.ObjectValue(dataStoreAttrTypes, map[string]attr.Value{
			"data_store_key":             types.StringValue(k),
			"filters":                    types.StringValue(""),
			"is_silent_error":            types.BoolValue(false),
			"is_effective_date_disabled": types.BoolValue(false),
			"use_union_for_incremental":  types.BoolValue(false),
			"initial_extract_date":       types.StringNull(),
			"chunk_type":                 types.StringNull(),
			"chunk_date_seq_incr":        types.Int64Value(0),
			"chunk_date_seq_min":         types.Int64Value(0),
			"chunk_pk_seq_incr":          types.Int64Value(0),
			"auto_populate_all_columns":  types.BoolValue(false),
			"column_overrides":           types.ListValueMust(types.ObjectType{AttrTypes: columnOverrideAttrTypes}, []attr.Value{}),
			"columns":                    types.ListValueMust(types.ObjectType{AttrTypes: columnAttrTypes}, []attr.Value{}),
		})
		if d.HasError() {
			t.Fatalf("building test object: %v", d)
		}
		objs[i] = obj
	}
	s, d := types.SetValue(types.ObjectType{AttrTypes: dataStoreAttrTypes}, objs)
	if d.HasError() {
		t.Fatalf("building test set: %v", d)
	}
	return s
}

func TestBuildOldDataStoreMap(t *testing.T) {
	ctx := context.Background()

	t.Run("null set returns empty map", func(t *testing.T) {
		m, diags := buildOldDataStoreMap(ctx, types.SetNull(types.ObjectType{AttrTypes: dataStoreAttrTypes}))
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(m) != 0 {
			t.Errorf("expected empty map, got %d entries", len(m))
		}
	})

	t.Run("populated set indexed by key", func(t *testing.T) {
		set := makeDataStoreSet(t, "DS.A", "DS.B", "DS.C")
		m, diags := buildOldDataStoreMap(ctx, set)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(m) != 3 {
			t.Errorf("expected 3 entries, got %d", len(m))
		}
		for _, key := range []string{"DS.A", "DS.B", "DS.C"} {
			if _, ok := m[key]; !ok {
				t.Errorf("key %q missing from map", key)
			}
		}
	})

	t.Run("map entry key matches model field", func(t *testing.T) {
		set := makeDataStoreSet(t, "MyDS.Key")
		m, _ := buildOldDataStoreMap(ctx, set)
		ds, ok := m["MyDS.Key"]
		if !ok {
			t.Fatal("key not found")
		}
		if ds.DataStoreKey.ValueString() != "MyDS.Key" {
			t.Errorf("DataStoreKey mismatch: %q", ds.DataStoreKey.ValueString())
		}
	})
}

// -------------------------------------------------------------------------
// normalizeDatePlanModifier — description methods
// -------------------------------------------------------------------------

func TestNormalizeDatePlanModifierDescription(t *testing.T) {
	m := normalizeDatePlanModifier{}
	ctx := context.Background()

	if m.Description(ctx) == "" {
		t.Error("Description should not be empty")
	}
	if m.MarkdownDescription(ctx) == "" {
		t.Error("MarkdownDescription should not be empty")
	}
}
