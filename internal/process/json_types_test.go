package process

import (
	"reflect"
	"testing"
)

func TestJsonTypes(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name      string
		fieldType reflect.Type
		tags      reflect.StructTag
		input     string
		want      any
		wantErr   bool
	}{
		{
			name:      "struct valid",
			fieldType: reflect.TypeOf(Config{}),
			input:     `{"name":"Alice","age":30}`,
			want:      Config{Name: "Alice", Age: 30},
		},
		{
			name:      "map valid",
			fieldType: reflect.TypeOf(map[string]string{}),
			input:     `{"foo":"bar"}`,
			want:      map[string]string{"foo": "bar"},
		},
		{
			name:      "invalid json",
			fieldType: reflect.TypeOf(Config{}),
			input:     `{"name":`,
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
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() got = %v, want %v", got, tt.want)
			}
		})
	}
}
