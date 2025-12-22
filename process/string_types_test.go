package process

import (
	"reflect"
	"testing"
)

func TestStringTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "string valid",
			fieldType: reflect.TypeOf(""),
			input:     "hello",
			want:      "hello",
		},
		{
			name:      "pattern validator pass",
			fieldType: reflect.TypeOf(""),
			tags:      `pattern:"^[a-z]+$"`,
			input:     "hello",
			want:      "hello",
		},
		{
			name:      "pattern validator fail",
			fieldType: reflect.TypeOf(""),
			tags:      `pattern:"^[a-z]+$"`,
			input:     "HELLO",
			wantErr:   true,
		},
		{
			name:      "lexical min validator pass",
			fieldType: reflect.TypeOf(""),
			tags:      `min:"b"`,
			input:     "cat",
			want:      "cat",
		},
		{
			name:      "lexical min validator fail",
			fieldType: reflect.TypeOf(""),
			tags:      `min:"b"`,
			input:     "apple",
			wantErr:   true,
		},
		{
			name:      "lexical max validator pass",
			fieldType: reflect.TypeOf(""),
			tags:      `max:"b"`,
			input:     "apple",
			want:      "apple",
		},
		{
			name:      "lexical max validator fail",
			fieldType: reflect.TypeOf(""),
			tags:      `max:"b"`,
			input:     "cat",
			wantErr:   true,
		},
		{
			name:      "lexical range validator pass",
			fieldType: reflect.TypeOf(""),
			tags:      `min:"b" max:"d"`,
			input:     "cat",
			want:      "cat",
		},
		{
			name:      "lexical range validator fail lower",
			fieldType: reflect.TypeOf(""),
			tags:      `min:"b" max:"c"`,
			input:     "apple",
			wantErr:   true,
		},
		{
			name:      "lexical range validator fail higher",
			fieldType: reflect.TypeOf(""),
			tags:      `min:"b" max:"c"`,
			input:     "cat", // ca is greater than c
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := New(tt.fieldType, tt.tags, nil, nil)
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
