package customtypes

import (
	"testing"
)

type MyString string

func TestNewStringEnum(t *testing.T) {
	handler := NewStringEnum[MyString]("A", "B", "C")
	pipeline, err := handler.BuildPipeline("")
	if err != nil {
		t.Fatalf("BuildPipeline failed: %v", err)
	}

	tests := []struct {
		input    string
		expected MyString
		wantErr  bool
	}{
		{"A", "A", false},
		{"B", "B", false},
		{"C", "C", false},
		{"D", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := pipeline(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("pipeline(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if val != tt.expected {
				t.Errorf("pipeline(%q) = %v, want %v", tt.input, val, tt.expected)
			}
		})
	}
}
