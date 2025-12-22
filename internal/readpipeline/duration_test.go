package readpipeline

import (
	"reflect"
	"testing"
	"time"
)

func TestDurationTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "duration valid",
			fieldType: reflect.TypeOf(time.Duration(0)),
			input:     "30s",
			want:      30 * time.Second,
		},
		{
			name:      "invalid duration",
			fieldType: reflect.TypeOf(time.Duration(0)),
			input:     "foo",
			wantErr:   true,
		},
		{
			name:      "duration min pass",
			fieldType: reflect.TypeOf(time.Duration(0)),
			tags:      `min:"10s"`,
			input:     "15s",
			want:      15 * time.Second,
		},
		{
			name:      "duration min fail",
			fieldType: reflect.TypeOf(time.Duration(0)),
			tags:      `min:"10s"`,
			input:     "5s",
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
