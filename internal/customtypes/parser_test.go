package customtypes

import (
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	parserFunc := func(rawValue string) (int, error) {
		if rawValue == "42" {
			return 42, nil
		}
		return 0, nil
	}

	handler := NewParser[int](parserFunc)
	if handler == nil {
		t.Fatal("NewParser returned nil")
	}

	pipeline, err := handler.BuildPipeline("")
	if err != nil {
		t.Fatalf("BuildPipeline failed: %v", err)
	}

	val, err := pipeline("42")
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}
}

func TestNewParser_BuildPipelineWithTags(t *testing.T) {
	parserFunc := func(rawValue string) (string, error) {
		return rawValue, nil
	}

	handler := NewParser[string](parserFunc)
	tags := reflect.StructTag(`key:"TEST_KEY"`)

	pipeline, err := handler.BuildPipeline(tags)
	if err != nil {
		t.Fatalf("BuildPipeline failed: %v", err)
	}

	val, err := pipeline("hello")
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected hello, got %v", val)
	}
}
