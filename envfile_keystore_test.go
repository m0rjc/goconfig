package goconfig

import (
	"context"
	"os"
	"testing"
)

func TestNewEnvFileKeyStore(t *testing.T) {
	// Create a temporary .env file
	content := `
PORT=9000
DB_HOST=localhost
# This is a comment
EMPTY=
QUOTED="quoted value"
SINGLE_QUOTED='single quoted'
`
	err := os.WriteFile(".env", []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer os.Remove(".env")

	t.Run("Default .env", func(t *testing.T) {
		store := NewEnvFileKeyStore()
		ctx := context.Background()

		tests := []struct {
			key     string
			wantVal string
			wantOk  bool
		}{
			{"PORT", "9000", true},
			{"DB_HOST", "localhost", true},
			{"EMPTY", "", true},
			{"QUOTED", "quoted value", true},
			{"SINGLE_QUOTED", "single quoted", true},
			{"MISSING", "", false},
		}

		for _, tt := range tests {
			val, ok, err := store(ctx, tt.key)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.key, err)
			}
			if ok != tt.wantOk {
				t.Errorf("%s: got ok %v, want %v", tt.key, ok, tt.wantOk)
			}
			if val != tt.wantVal {
				t.Errorf("%s: got val %q, want %q", tt.key, val, tt.wantVal)
			}
		}
	})

	t.Run("Specific files", func(t *testing.T) {
		f1 := "test1.env"
		f2 := "test2.env"
		os.WriteFile(f1, []byte("KEY1=VAL1\nKEY2=VAL2_F1"), 0644)
		os.WriteFile(f2, []byte("KEY2=VAL2_F2\nKEY3=VAL3"), 0644)
		defer os.Remove(f1)
		defer os.Remove(f2)

		store := NewEnvFileKeyStore(f1, f2)
		ctx := context.Background()

		tests := []struct {
			key     string
			wantVal string
			wantOk  bool
		}{
			{"KEY1", "VAL1", true},
			{"KEY2", "VAL2_F1", true}, // First one wins in my implementation
			{"KEY3", "VAL3", true},
		}

		for _, tt := range tests {
			val, ok, _ := store(ctx, tt.key)
			if ok != tt.wantOk || val != tt.wantVal {
				t.Errorf("%s: got (%q, %v), want (%q, %v)", tt.key, val, ok, tt.wantVal, tt.wantOk)
			}
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		store := NewEnvFileKeyStore("nonexistent.env")
		ctx := context.Background()
		_, ok, _ := store(ctx, "ANY")
		if ok {
			t.Error("Expected ok=false for nonexistent file")
		}
	})
}
