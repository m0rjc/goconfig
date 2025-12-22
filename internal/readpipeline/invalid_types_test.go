package readpipeline

import (
	"reflect"
	"testing"
)

func TestInvalidTypes(t *testing.T) {
	// This test will need to be removed or replaced if we ever support complex numbers
	registry := NewDefaultTypeRegistry()
	t.Run("Complex128", func(t *testing.T) {
		fieldType := reflect.TypeOf(complex128(0))
		tags := reflect.StructTag("")

		_, err := New(fieldType, tags, registry)
		if err == nil {
			t.Fatal("Expected error for complex128 type, but got nil")
		}

		expectedErr := "no handler for type complex128"
		if err.Error() != expectedErr {
			t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
		}
	})

	t.Run("Interface", func(t *testing.T) {
		var i any
		fieldType := reflect.TypeOf(&i).Elem()
		tags := reflect.StructTag("")

		_, err := New(fieldType, tags, registry)
		if err == nil {
			t.Fatal("Expected error for interface type, but got nil")
		}

		expectedErr := "no handler for type interface {}"
		if err.Error() != expectedErr {
			t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
		}
	})
}
