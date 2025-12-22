package readpipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
		customValidator := func(v int64) error {
			if v%2 != 0 {
				return errors.New("must be even")
			}
			return nil
		}

		// Tag validators: min=10. Note: custom parser doubles this, so effective min is 20
		tags := reflect.StructTag(`key:"PORT" min:"10"`)
		fieldType := reflect.TypeOf(int64(0))

		registry := NewDefaultTypeRegistry()
		registry.RegisterType(fieldType, typeHandlerImpl[int64]{
			Parser: func(s string) (int64, error) {
				v, err := customParser(s)
				if err != nil {
					return 0, err
				}
				return v.(int64), nil
			},
			ValidationWrapper: NewCompositeWrapper(
				func(tags reflect.StructTag, inputProcess FieldProcessor[int64]) (FieldProcessor[int64], error) {
					return Pipe(inputProcess, customValidator), nil
				},
				WrapProcessUsingRangeTags[int64],
			),
		})

		p, err := New(fieldType, tags, registry)
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
		registry := NewDefaultTypeRegistry()
		registry.RegisterType(fieldType, NewCustomHandler(func(s string) (Point, error) {
			v, err := customParser(s)
			if err != nil {
				return Point{}, err
			}
			return v.(Point), nil
		}, func(v Point) error {
			return customValidator(v)
		}))
		p, err := New(fieldType, "", registry)
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
		customValidator := func(value int64) error {
			if value == 42 {
				return errors.New("42 is forbidden")
			}
			return nil
		}

		fieldType := reflect.TypeOf(int64(0))
		registry := NewDefaultTypeRegistry()
		// Since we want to use the default parser but add a custom validator, we can prepend it
		baseHandler := NewTypedIntHandler(64)
		handler, err := PrependValidators(baseHandler, customValidator)
		if err != nil {
			t.Fatalf("Failed to prepend validator: %v", err)
		}
		registry.RegisterType(fieldType, handler)
		p, err := New(fieldType, "", registry)
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

	t.Run("Custom parser for non-built-in type", func(t *testing.T) {
		// Build a parser for a complex number and run it through the pipeline
		customParser := func(rawValue string) (any, error) {
			return complex(1, 2), nil
		}
		fieldType := reflect.TypeOf(complex(0, 0))
		registry := NewDefaultTypeRegistry()
		registry.RegisterType(fieldType, NewCustomHandler(func(s string) (complex128, error) {
			v, err := customParser(s)
			return v.(complex128), err
		}))
		p, err := New(fieldType, "", registry)
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		value, err := p("This value is ignored by the mock parser")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if value != complex(1, 2) {
			t.Errorf("Expected complex(1, 2), got %v", value)
		}
	})

	t.Run("ReplaceParser and PrependValidators", func(t *testing.T) {
		baseHandler := NewTypedIntHandler(64)

		t.Run("Parser override via ReplaceParser", func(t *testing.T) {
			// Replace the parser with one that always returns 42
			decorated, err := ReplaceParser(baseHandler, func(s string) (int64, error) {
				return 42, nil
			})
			if err != nil {
				t.Fatalf("ReplaceParser failed: %v", err)
			}

			p, err := decorated.Build("")
			if err != nil {
				t.Fatalf("Build failed: %v", err)
			}
			val, err := p("any value")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if val.(int64) != 42 {
				t.Errorf("Expected 42, got %v", val)
			}
		})

		t.Run("Validator prepending via PrependValidators", func(t *testing.T) {
			// Base has range validation (via NewTypedIntHandler)
			// Prepend a check for even numbers
			decorated, err := PrependValidators(baseHandler, func(v int64) error {
				if v%2 != 0 {
					return errors.New("must be even")
				}
				return nil
			})
			if err != nil {
				t.Fatalf("PrependValidators failed: %v", err)
			}

			// tags with min=10
			tags := reflect.StructTag(`min:"10"`)
			p, err := decorated.Build(tags)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			tests := []struct {
				input   string
				wantErr string
			}{
				{"12", ""},                // Pass: >= 10 and even
				{"11", "must be even"},    // Fail: >= 10 but odd (prepended validator fails)
				{"8", "below minimum 10"}, // Fail: < 10 (base validator fails)
			}

			for _, tt := range tests {
				_, err := p(tt.input)
				if tt.wantErr == "" {
					if err != nil {
						t.Errorf("input %s: unexpected error %v", tt.input, err)
					}
				} else {
					if err == nil {
						t.Errorf("input %s: expected error %q, got nil", tt.input, tt.wantErr)
					} else if !strings.Contains(err.Error(), tt.wantErr) {
						t.Errorf("input %s: expected error to contain %q, got %q", tt.input, tt.wantErr, err.Error())
					}
				}
			}
		})

		t.Run("Multiple prepended validators", func(t *testing.T) {
			// Prepend "must be even"
			handler1, _ := PrependValidators(baseHandler, func(v int64) error {
				if v%2 != 0 {
					return errors.New("must be even")
				}
				return nil
			})
			// Prepend "must be positive"
			handler2, _ := PrependValidators(handler1, func(v int64) error {
				if v <= 0 {
					return errors.New("must be positive")
				}
				return nil
			})

			p, _ := handler2.Build("")
			if _, err := p("-2"); err == nil || !strings.Contains(err.Error(), "must be positive") {
				t.Errorf("expected positive error, got %v", err)
			}
			if _, err := p("3"); err == nil || !strings.Contains(err.Error(), "must be even") {
				t.Errorf("expected even error, got %v", err)
			}
			if v, err := p("4"); err != nil || v.(int64) != 4 {
				t.Errorf("expected 4, got %v (err: %v)", v, err)
			}
		})
	})
}
