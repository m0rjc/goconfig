package process

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestCustomParserAndValidators(t *testing.T) {
	t.Run("Custom parser for built-in int with custom and tag validators", func(t *testing.T) {
		// Custom parser that doubles the input value
		customParser := func(rawValue string) (any, error) {
			val, err := strconv.Atoi(rawValue)
			if err != nil {
				return nil, err
			}
			return int64(val * 2), nil
		}

		// Custom validator that checks if even
		customValidator := func(value any) error {
			v := value.(int64)
			if v%2 != 0 {
				return errors.New("must be even")
			}
			return nil
		}

		// Tag validators: min=10. Note: custom parser doubles this, so effective min is 20
		tags := reflect.StructTag(`key:"PORT" min:"10"`)
		fieldType := reflect.TypeOf(int64(0))

		p, err := New(fieldType, tags, customParser, []Validator[any]{customValidator})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		tests := []struct {
			input   string
			want    int64
			wantErr bool
		}{
			{"11", 22, false}, // 11*2 = 22, which is >= 20 and even
			{"6", 12, true},   // 6*2 = 12, which is < 20 (tag validator fails)
			{"3", 6, true},    // 3*2 = 6, which is < 20 (tag validator fails)
			{"foo", 0, true},  // Invalid input (parser fails)
		}

		for _, tt := range tests {
			got, err := p(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("p(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				continue
			}
			if !tt.wantErr && got.(int64) != tt.want {
				t.Errorf("p(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})

	t.Run("Custom parser for custom struct type", func(t *testing.T) {
		type Point struct {
			X, Y int
		}

		// Custom parser "X,Y"
		customParser := func(rawValue string) (any, error) {
			var x, y int
			_, err := fmt.Sscanf(rawValue, "%d,%d", &x, &y)
			if err != nil {
				return nil, errors.New("invalid point format")
			}
			return Point{X: x, Y: y}, nil
		}

		// Custom validator: X must be positive
		customValidator := func(value any) error {
			p := value.(Point)
			if p.X < 0 {
				return errors.New("X must be positive")
			}
			return nil
		}

		fieldType := reflect.TypeOf(Point{})
		p, err := New(fieldType, "", customParser, []Validator[any]{customValidator})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		tests := []struct {
			input   string
			want    Point
			wantErr bool
		}{
			{"1,2", Point{1, 2}, false},
			{"-1,2", Point{-1, 2}, true}, // Validator fails
			{"bad", Point{}, true},       // Parser fails
		}

		for _, tt := range tests {
			got, err := p(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("p(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				continue
			}
			if !tt.wantErr && got.(Point) != tt.want {
				t.Errorf("p(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})

	t.Run("Custom validator against built-in int", func(t *testing.T) {
		// No custom parser, use default int64 parser
		customValidator := func(value any) error {
			v := value.(int64)
			if v == 42 {
				return errors.New("42 is forbidden")
			}
			return nil
		}

		fieldType := reflect.TypeOf(int64(0))
		p, err := New(fieldType, "", nil, []Validator[any]{customValidator})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		if _, err := p("42"); err == nil {
			t.Error("Expected error for 42, got nil")
		}
		if got, err := p("10"); err != nil || got.(int64) != 10 {
			t.Errorf("p(10) = %v, %v, want 10, nil", got, err)
		}
	})
}
