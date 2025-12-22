package readpipeline

import (
	"reflect"
	"testing"
)

func TestPointerTypes(t *testing.T) {
	registry := NewDefaultTypeRegistry()
	t.Run("PointerToInt", func(t *testing.T) {
		var i *int
		fieldType := reflect.TypeOf(i)
		tags := reflect.StructTag(`min:"10"`)

		processor, err := New(fieldType, tags, registry)
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		// Valid value
		val, err := processor("15")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != int64(15) {
			t.Errorf("Expected int64(15), got %v (%T)", val, val)
		}

		// Below minimum
		_, err = processor("5")
		if err == nil {
			t.Error("Expected error for value below minimum, got nil")
		}
	})

	t.Run("PointerToString", func(t *testing.T) {
		var s *string
		fieldType := reflect.TypeOf(s)
		tags := reflect.StructTag(`pattern:"^abc.*$"`)

		processor, err := New(fieldType, tags, registry)
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		// Valid value
		val, err := processor("abcdef")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != "abcdef" {
			t.Errorf("Expected 'abcdef', got %v", val)
		}

		// Invalid pattern
		_, err = processor("ghijk")
		if err == nil {
			t.Error("Expected error for pattern mismatch, got nil")
		}
	})

	t.Run("PointerToStruct", func(t *testing.T) {
		type MyStruct struct {
			Name string `json:"name"`
		}
		var ms *MyStruct
		fieldType := reflect.TypeOf(ms)

		processor, err := New(fieldType, "", registry)
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		// Valid JSON
		val, err := processor(`{"name": "test"}`)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := MyStruct{Name: "test"}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("Expected %v, got %v", expected, val)
		}
	})

	t.Run("PointerToCustomTypeWithCustomParser", func(t *testing.T) {
		type Point struct {
			X, Y int
		}
		var p *Point
		fieldType := reflect.TypeOf(p)

		customParser := func(rawValue string) (Point, error) {
			// Dummy parser for "1,2"
			return Point{X: 1, Y: 2}, nil
		}

		registry := NewDefaultTypeRegistry()
		registry.RegisterType(reflect.TypeOf(Point{}), WrapTypedHandler(NewCustomHandler(customParser)))

		processor, err := New(fieldType, "", registry)
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		val, err := processor("anything")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := Point{X: 1, Y: 2}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("Expected %v, got %v", expected, val)
		}
	})
}
