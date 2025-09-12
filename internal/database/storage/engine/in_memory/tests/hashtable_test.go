package tests

import (
	"testing"

	inmemory "github.com/WithSoull/in-memory-database/internal/database/storage/engine/in_memory"
	"github.com/stretchr/testify/require"
)

func TestHashtable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(ht inmemory.Hashtable)
		key      string
		want     string
		wantOk   bool
		testType string // "get", "del", or "setup_only"
	}{
		// Basic Get tests
		{
			name:     "get non-existent key",
			setup:    func(ht inmemory.Hashtable) {},
			key:      "nonexistent",
			want:     "",
			wantOk:   false,
			testType: "get",
		},
		{
			name: "get existing key",
			setup: func(ht inmemory.Hashtable) {
				ht.Set("test", "value")
			},
			key:      "test",
			want:     "value",
			wantOk:   true,
			testType: "get",
		},
		{
			name: "get after overwrite",
			setup: func(ht inmemory.Hashtable) {
				ht.Set("key", "old")
				ht.Set("key", "new")
			},
			key:      "key",
			want:     "new",
			wantOk:   true,
			testType: "get",
		},

		// Edge cases
		{
			name: "get with empty key",
			setup: func(ht inmemory.Hashtable) {
				ht.Set("", "empty_key_value")
			},
			key:      "",
			want:     "empty_key_value",
			wantOk:   true,
			testType: "get",
		},
		{
			name: "get with empty value",
			setup: func(ht inmemory.Hashtable) {
				ht.Set("empty_val", "")
			},
			key:      "empty_val",
			want:     "",
			wantOk:   true,
			testType: "get",
		},

		// Delete tests
		{
			name: "delete existing key",
			setup: func(ht inmemory.Hashtable) {
				ht.Set("to_delete", "value")
			},
			key:      "to_delete",
			testType: "del",
		},
		{
			name:     "delete non-existent key",
			setup:    func(ht inmemory.Hashtable) {},
			key:      "nonexistent",
			testType: "del",
		},

		// Setup only tests (for multiple operations)
		{
			name: "multiple operations",
			setup: func(ht inmemory.Hashtable) {
				ht.Set("key1", "val1")
				ht.Set("key2", "val2")
				ht.Set("key3", "val3")
				ht.Del("key2")
				ht.Set("key1", "updated")
			},
			testType: "setup_only",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ht := inmemory.NewHashtable()

			// Setup the initial state
			tt.setup(ht)

			switch tt.testType {
			case "get":
				got, ok := ht.Get(tt.key)
				require.Equal(t, tt.wantOk, ok, "Get() ok result mismatch")
				require.Equal(t, tt.want, got, "Get() value mismatch")

			case "del":
				// Verify key exists before deletion if it should
				if tt.setup != nil {
					_, initialOk := ht.Get(tt.key)
					ht.Del(tt.key)

					// Verify key is gone after deletion
					_, finalOk := ht.Get(tt.key)
					if initialOk {
						require.False(t, finalOk, "Key should be deleted but still exists")
					} else {
						require.False(t, finalOk, "Non-existent key should remain non-existent after deletion")
					}
				} else {
					ht.Del(tt.key) // Should not panic
				}

			case "setup_only":
				// For complex setup, verify final state
				val1, ok1 := ht.Get("key1")
				require.True(t, ok1, "key1 should exist")
				require.Equal(t, "updated", val1, "key1 should have updated value")

				_, ok2 := ht.Get("key2")
				require.False(t, ok2, "key2 should be deleted")

				val3, ok3 := ht.Get("key3")
				require.True(t, ok3, "key3 should exist")
				require.Equal(t, "val3", val3, "key3 should have original value")
			}
		})
	}
}

func TestNewHashtable(t *testing.T) {
	ht := inmemory.NewHashtable()
	require.NotNil(t, ht, "NewHashtable() should not return nil")

	// Verify it's empty
	val, ok := ht.Get("any")
	require.False(t, ok, "New hashtable should be empty")
	require.Empty(t, val, "Value should be empty for non-existent key")
}
