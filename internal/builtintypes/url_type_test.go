package builtintypes

import (
	"net/url"
	"reflect"
	"testing"
)

func TestUrlTypedHandler(t *testing.T) {
	handler := NewUrlTypedHandler()
	if handler == nil {
		t.Fatal("NewUrlTypedHandler returned nil")
	}

	t.Run("ValidURL", func(t *testing.T) {
		pipeline, err := handler.BuildPipeline("")
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		var u *url.URL
		u, err = pipeline("http://example.com/path?q=1")
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}

		if u.String() != "http://example.com/path?q=1" {
			t.Errorf("expected http://example.com/path?q=1, got %s", u.String())
		}
	})

	t.Run("InvalidURL", func(t *testing.T) {
		pipeline, err := handler.BuildPipeline("")
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		_, err = pipeline("not a url")
		if err == nil {
			t.Error("expected error for invalid URL, got nil")
		}
	})

	t.Run("PatternValidation", func(t *testing.T) {
		tags := reflect.StructTag(`pattern:"^https://.*"`)
		pipeline, err := handler.BuildPipeline(tags)
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		_, err = pipeline("https://example.com")
		if err != nil {
			t.Errorf("expected success for https://example.com, got %v", err)
		}

		_, err = pipeline("http://example.com")
		if err == nil {
			t.Error("expected error for http://example.com (not matching pattern), got nil")
		}
	})

	t.Run("InvalidPattern", func(t *testing.T) {
		tags := reflect.StructTag(`pattern:"["`)
		_, err := handler.BuildPipeline(tags)
		if err == nil {
			t.Error("expected error for invalid pattern, got nil")
		}
	})

	t.Run("SchemeValidation", func(t *testing.T) {
		tags := reflect.StructTag(`scheme:"https,mailto"`)
		pipeline, err := handler.BuildPipeline(tags)
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		_, err = pipeline("https://example.com")
		if err != nil {
			t.Errorf("expected success for https, got %v", err)
		}

		_, err = pipeline("mailto:user@example.com")
		if err != nil {
			t.Errorf("expected success for mailto, got %v", err)
		}

		_, err = pipeline("http://example.com")
		if err == nil {
			t.Error("expected error for http, got nil")
		}
	})

	t.Run("CombinedValidation", func(t *testing.T) {
		tags := reflect.StructTag(`scheme:"https" pattern:".*example\\.com.*"`)
		pipeline, err := handler.BuildPipeline(tags)
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		_, err = pipeline("https://example.com/test")
		if err != nil {
			t.Errorf("expected success, got %v", err)
		}

		_, err = pipeline("https://other.com")
		if err == nil {
			t.Error("expected pattern error, got nil")
		}

		_, err = pipeline("http://example.com")
		if err == nil {
			t.Error("expected scheme error, got nil")
		}
	})
}
