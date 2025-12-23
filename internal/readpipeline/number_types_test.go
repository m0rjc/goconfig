package readpipeline

import (
	"reflect"
	"testing"
)

func TestIntTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "int valid",
			fieldType: reflect.TypeOf(int(0)),
			input:     "42",
			want:      int64(42),
		},
		{
			name:      "int8 valid",
			fieldType: reflect.TypeOf(int8(0)),
			input:     "127",
			want:      int64(127),
		},
		{
			name:      "int8 overflow",
			fieldType: reflect.TypeOf(int8(0)),
			input:     "128",
			wantErr:   true,
		},
		{
			name:      "hex input",
			fieldType: reflect.TypeOf(int(0)),
			input:     "0xFF",
			want:      int64(255),
		},
		{
			name:      "invalid int",
			fieldType: reflect.TypeOf(int(0)),
			input:     "foo",
			wantErr:   true,
		},
		{
			name:      "int min validator pass",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"10"`,
			input:     "10",
			want:      int64(10),
		},
		{
			name:      "int min validator fail",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"10"`,
			input:     "9",
			wantErr:   true,
		},
		{
			name:      "int max validator pass",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `max:"100"`,
			input:     "100",
			want:      int64(100),
		},
		{
			name:      "int max validator fail",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `max:"100"`,
			input:     "101",
			wantErr:   true,
		},
		{
			name:      "int range pass",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"10" max:"20"`,
			input:     "15",
			want:      int64(15),
		},
		{
			name:      "int range fail lower",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"10" max:"20"`,
			input:     "9",
			wantErr:   true,
		},
		{
			name:      "int range fail higher",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"10" max:"20"`,
			input:     "21",
			wantErr:   true,
		},
		{
			name:      "int min hex pass",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"0x10"`, // 16 in hex
			input:     "16",
			want:      int64(16),
		},
		{
			name:      "int min hex fail",
			fieldType: reflect.TypeOf(int(0)),
			tags:      `min:"0x10"`, // 16 in hex
			input:     "0x0F",
			wantErr:   true,
		},
	}

	registry := NewTypeRegistry()
	t.Run("invalid min tag", func(t *testing.T) {
		_, err := New(reflect.TypeOf(int(0)), `min:"foo"`, registry)
		if err == nil {
			t.Error("expected error for invalid min tag, got nil")
		}
	})

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

func TestUintTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "uint valid",
			fieldType: reflect.TypeOf(uint(0)),
			input:     "42",
			want:      uint64(42),
		},
		{
			name:      "uint8 overflow",
			fieldType: reflect.TypeOf(uint8(0)),
			input:     "256",
			wantErr:   true,
		},
		{
			name:      "invalid uint",
			fieldType: reflect.TypeOf(uint(0)),
			input:     "-1",
			wantErr:   true,
		},
		{
			name:      "uint min validator pass",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `min:"10"`,
			input:     "10",
			want:      uint64(10),
		},
		{
			name:      "uint min validator fail lower",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `min:"10"`,
			input:     "9",
			wantErr:   true,
		},
		{
			name:      "uint max validator pass",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `max:"10"`,
			input:     "10",
			want:      uint64(10),
		},
		{
			name:      "uint max validator fail",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `max:"10"`,
			input:     "11",
			wantErr:   true,
		},
		{
			name:      "uint range validator pass",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `min:"5" max:"10"`,
			input:     "6",
			want:      uint64(6),
		},
		{
			name:      "uint range validator fail lower",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `min:"5" max:"10"`,
			input:     "4",
			wantErr:   true,
		},
		{
			name:      "uint range validator fail higher",
			fieldType: reflect.TypeOf(uint(0)),
			tags:      `min:"5" max:"10"`,
			input:     "11",
			wantErr:   true,
		},
	}

	registry := NewTypeRegistry()
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

func TestFloatTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "float64 valid",
			fieldType: reflect.TypeOf(float64(0)),
			input:     "3.14",
			want:      3.14,
		},
		{
			name:      "invalid float",
			fieldType: reflect.TypeOf(float64(0)),
			input:     "foo",
			wantErr:   true,
		},
		{
			name:      "float range validator",
			fieldType: reflect.TypeOf(float64(0)),
			tags:      `min:"0.0" max:"1.0"`,
			input:     "1.1",
			wantErr:   true,
		},
	}

	registry := NewTypeRegistry()
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
