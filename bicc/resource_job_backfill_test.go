package bicc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// -------------------------------------------------------------------------
// convertToISO8601
// -------------------------------------------------------------------------

func TestConvertToISO8601(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "plain YYYY-MM-DD",
			input: "2025-01-15",
			want:  "2025-01-15T00:00:00.000Z",
		},
		{
			name:  "already ISO 8601 — date part extracted and re-formatted",
			input: "2025-01-15T00:00:00.000Z",
			want:  "2025-01-15T00:00:00.000Z",
		},
		{
			name:  "ISO with non-midnight time — normalised to midnight",
			input: "2025-06-01T12:34:56.000Z",
			want:  "2025-06-01T00:00:00.000Z",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "MM/DD/YYYY not accepted",
			input:   "01/15/2025",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertToISO8601(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// -------------------------------------------------------------------------
// generateBackfillID
// -------------------------------------------------------------------------

func TestGenerateBackfillID(t *testing.T) {
	entry := func(key string) backfillEntryModel {
		return backfillEntryModel{
			DataStoreKey:    types.StringValue(key),
			LastExtractDate: types.StringValue("2025-01-01"),
		}
	}

	t.Run("deterministic — same input same ID", func(t *testing.T) {
		entries := []backfillEntryModel{entry("DS.A"), entry("DS.B")}
		id1 := generateBackfillID(42, entries)
		id2 := generateBackfillID(42, entries)
		if id1 != id2 {
			t.Errorf("non-deterministic: %q vs %q", id1, id2)
		}
	})

	t.Run("order-independent", func(t *testing.T) {
		// DS.A, DS.B vs DS.B, DS.A should produce the same ID.
		ab := generateBackfillID(1, []backfillEntryModel{entry("DS.A"), entry("DS.B")})
		ba := generateBackfillID(1, []backfillEntryModel{entry("DS.B"), entry("DS.A")})
		if ab != ba {
			t.Errorf("order should not matter: %q vs %q", ab, ba)
		}
	})

	t.Run("different job IDs produce different IDs", func(t *testing.T) {
		entries := []backfillEntryModel{entry("DS.A")}
		id1 := generateBackfillID(1, entries)
		id2 := generateBackfillID(2, entries)
		if id1 == id2 {
			t.Error("different job IDs should produce different resource IDs")
		}
	})

	t.Run("different data_store_keys produce different IDs", func(t *testing.T) {
		id1 := generateBackfillID(1, []backfillEntryModel{entry("DS.A")})
		id2 := generateBackfillID(1, []backfillEntryModel{entry("DS.B")})
		if id1 == id2 {
			t.Error("different keys should produce different resource IDs")
		}
	})

	t.Run("ID contains job ID prefix", func(t *testing.T) {
		id := generateBackfillID(77, []backfillEntryModel{entry("DS.A")})
		if len(id) < 3 || id[:3] != "77:" {
			t.Errorf("expected ID to start with '77:', got %q", id)
		}
	})
}
