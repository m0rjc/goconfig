package goconfig

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

// mockHandler captures log records for testing
type mockHandler struct {
	records []logRecord
}

type logRecord struct {
	level   slog.Level
	message string
	attrs   map[string]any
}

func (h *mockHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *mockHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.records = append(h.records, logRecord{
		level:   r.Level,
		message: r.Message,
		attrs:   attrs,
	})
	return nil
}

func (h *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *mockHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestLogError_WithNonConfigErrors(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	err := errors.New("database connection failed")
	LogError(logger, err)

	if len(handler.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(handler.records))
	}

	record := handler.records[0]
	if record.level != slog.LevelError {
		t.Errorf("expected level Error, got %v", record.level)
	}

	if record.message != "configuration error" {
		t.Errorf("expected message 'configuration error', got %q", record.message)
	}

	if errVal, ok := record.attrs["error"]; !ok || errVal.(error).Error() != "database connection failed" {
		t.Errorf("expected error attribute with message 'database connection failed', got %v", errVal)
	}
}

func TestLogError_WithNonConfigErrors_CustomMessage(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	err := errors.New("database connection failed")
	LogError(logger, err, WithLogMessage("custom error message"))

	if len(handler.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(handler.records))
	}

	record := handler.records[0]
	if record.message != "custom error message" {
		t.Errorf("expected message 'custom error message', got %q", record.message)
	}
}

func TestLogError_WithConfigErrors(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	configErrs := &ConfigErrors{}
	configErrs.Add("DB_HOST", errors.New("invalid host"))
	configErrs.Add("DB_PORT", errors.New("port must be between 1 and 65535"))
	configErrs.Add("API_KEY", ErrMissingValue)

	LogError(logger, configErrs)

	if len(handler.records) != 3 {
		t.Fatalf("expected 3 log records, got %d", len(handler.records))
	}

	// Verify first error
	record := handler.records[0]
	if record.level != slog.LevelError {
		t.Errorf("expected level Error, got %v", record.level)
	}
	if record.message != "configuration error" {
		t.Errorf("expected message 'configuration error', got %q", record.message)
	}
	if key, ok := record.attrs["key"]; !ok || key != "DB_HOST" {
		t.Errorf("expected key 'DB_HOST', got %v", key)
	}
	if errVal, ok := record.attrs["error"]; !ok || errVal.(error).Error() != "invalid host" {
		t.Errorf("expected error 'invalid host', got %v", errVal)
	}

	// Verify second error
	record = handler.records[1]
	if key, ok := record.attrs["key"]; !ok || key != "DB_PORT" {
		t.Errorf("expected key 'DB_PORT', got %v", key)
	}
	if errVal, ok := record.attrs["error"]; !ok || errVal.(error).Error() != "port must be between 1 and 65535" {
		t.Errorf("expected error 'port must be between 1 and 65535', got %v", errVal)
	}

	// Verify third error
	record = handler.records[2]
	if key, ok := record.attrs["key"]; !ok || key != "API_KEY" {
		t.Errorf("expected key 'API_KEY', got %v", key)
	}
	if errVal, ok := record.attrs["error"]; !ok || errVal.(error) != ErrMissingValue {
		t.Errorf("expected error ErrMissingValue, got %v", errVal)
	}
}

func TestLogError_WithConfigErrors_CustomMessage(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	configErrs := &ConfigErrors{}
	configErrs.Add("DB_HOST", errors.New("invalid host"))
	configErrs.Add("DB_PORT", errors.New("port out of range"))

	LogError(logger, configErrs, WithLogMessage("validation failed"))

	if len(handler.records) != 2 {
		t.Fatalf("expected 2 log records, got %d", len(handler.records))
	}

	for i, record := range handler.records {
		if record.message != "validation failed" {
			t.Errorf("record %d: expected message 'validation failed', got %q", i, record.message)
		}
	}
}

func TestConfigErrors_LogAll(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	configErrs := &ConfigErrors{}
	configErrs.Add("FIELD1", errors.New("error 1"))
	configErrs.Add("FIELD2", errors.New("error 2"))

	configErrs.LogAll(logger)

	if len(handler.records) != 2 {
		t.Fatalf("expected 2 log records, got %d", len(handler.records))
	}

	// Verify the log records
	expectedKeys := []string{"FIELD1", "FIELD2"}
	expectedErrors := []string{"error 1", "error 2"}

	for i, record := range handler.records {
		if record.level != slog.LevelError {
			t.Errorf("record %d: expected level Error, got %v", i, record.level)
		}

		if record.message != "configuration error" {
			t.Errorf("record %d: expected message 'configuration error', got %q", i, record.message)
		}

		if key, ok := record.attrs["key"]; !ok || key != expectedKeys[i] {
			t.Errorf("record %d: expected key %q, got %v", i, expectedKeys[i], key)
		}

		if errVal, ok := record.attrs["error"]; !ok || errVal.(error).Error() != expectedErrors[i] {
			t.Errorf("record %d: expected error %q, got %v", i, expectedErrors[i], errVal)
		}
	}
}

func TestConfigErrors_LogAll_WithCustomMessage(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	configErrs := &ConfigErrors{}
	configErrs.Add("FIELD1", errors.New("error 1"))

	configErrs.LogAll(logger, WithLogMessage("custom validation error"))

	if len(handler.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(handler.records))
	}

	record := handler.records[0]
	if record.message != "custom validation error" {
		t.Errorf("expected message 'custom validation error', got %q", record.message)
	}
}

func TestConfigErrors_LogAll_EmptyErrors(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)

	configErrs := &ConfigErrors{}
	configErrs.LogAll(logger)

	if len(handler.records) != 0 {
		t.Errorf("expected 0 log records for empty ConfigErrors, got %d", len(handler.records))
	}
}

func TestWithLogMessage(t *testing.T) {
	settings := getLogSettings(WithLogMessage("test message"))

	if settings.message != "test message" {
		t.Errorf("expected message 'test message', got %q", settings.message)
	}
}

func TestGetLogSettings_DefaultMessage(t *testing.T) {
	settings := getLogSettings()

	if settings.message != "configuration error" {
		t.Errorf("expected default message 'configuration error', got %q", settings.message)
	}
}
