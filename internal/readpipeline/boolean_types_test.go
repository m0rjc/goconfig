package readpipeline

import (
	"reflect"
	"testing"
)

func TestBoolTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "bool true",
			fieldType: reflect.TypeOf(true),
			input:     "true",
			want:      true,
		},
		{
			name:      "bool 1",
			fieldType: reflect.TypeOf(true),
			input:     "1",
			want:      true,
		},
		{
			name:      "bool false",
			fieldType: reflect.TypeOf(true),
			input:     "false",
			want:      false,
		},
		{
			name:      "bool 0",
			fieldType: reflect.TypeOf(true),
			input:     "0",
			want:      false,
		},
		{
			name:      "invalid bool",
			fieldType: reflect.TypeOf(true),
			input:     "notabool",
			wantErr:   true,
		},
	}

	registry := NewDefaultTypeRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := New(tt.fieldType, tt.tags, registry)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			got, err := proc(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Process() got = %v, want %v", got, tt.want)
			}
		})
	}
}
