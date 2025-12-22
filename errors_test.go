package goconfig

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestConfigErrors(t *testing.T) {
	t.Run("Basic operations", func(t *testing.T) {
		ce := &ConfigErrors{}
		if ce.HasErrors() {
			t.Error("expected HasErrors to be false")
		}
		if ce.Len() != 0 {
			t.Errorf("expected Len 0, got %d", ce.Len())
		}

		err1 := errors.New("error 1")
		ce.Add("KEY1", err1)

		if !ce.HasErrors() {
			t.Error("expected HasErrors to be true")
		}
		if ce.Len() != 1 {
			t.Errorf("expected Len 1, got %d", ce.Len())
		}

		err2 := errors.New("error 2")
		ce.Add("KEY2", err2)
		if ce.Len() != 2 {
			t.Errorf("expected Len 2, got %d", ce.Len())
		}

		unwrapped := ce.Unwrap()
		if len(unwrapped) != 2 {
			t.Errorf("expected 2 unwrapped errors, got %d", len(unwrapped))
		}
		if unwrapped[0] != err1 || unwrapped[1] != err2 {
			t.Error("unwrapped errors mismatch")
		}
	})

	t.Run("Error formatting and prefix stripping", func(t *testing.T) {
		ce := &ConfigErrors{}
		ce.Add("PORT", errors.New("invalid port"))
		ce.Add("HOST", fmt.Errorf("invalid value for HOST: empty not allowed"))

		got := ce.Error()
		// Expect prefix stripping for HOST
		// "PORT: invalid port\nHOST: empty not allowed"
		expected := "PORT: invalid port\nHOST: empty not allowed"
		if got != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, got)
		}
	})

	t.Run("Empty ConfigErrors formatting", func(t *testing.T) {
		ce := &ConfigErrors{}
		if ce.Error() != "" {
			t.Errorf("expected empty string for no errors, got %q", ce.Error())
		}
	})
}

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		// Remove time to make it deterministic
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	t.Run("LogAll", func(t *testing.T) {
		buf.Reset()
		ce := &ConfigErrors{}
		ce.Add("KEY1", errors.New("err1"))
		ce.LogAll(logger)

		output := buf.String()
		if !strings.Contains(output, "level=ERROR") {
			t.Error("expected ERROR level")
		}
		if !strings.Contains(output, "msg=\"configuration error\"") {
			t.Error("expected default message")
		}
		if !strings.Contains(output, "key=KEY1") || !strings.Contains(output, "error=err1") {
			t.Error("missing key or error in log")
		}
	})

	t.Run("LogAll with custom message", func(t *testing.T) {
		buf.Reset()
		ce := &ConfigErrors{}
		ce.Add("KEY1", errors.New("err1"))
		ce.LogAll(logger, WithLogMessage("custom_msg"))

		output := buf.String()
		if !strings.Contains(output, "msg=custom_msg") {
			t.Error("expected custom message")
		}
	})

	t.Run("LogError with ConfigErrors", func(t *testing.T) {
		buf.Reset()
		ce := &ConfigErrors{}
		ce.Add("KEY1", errors.New("err1"))
		LogError(logger, ce)

		output := buf.String()
		if !strings.Contains(output, "key=KEY1") {
			t.Error("expected individual error logging via LogError")
		}
	})

	t.Run("LogError with regular error", func(t *testing.T) {
		buf.Reset()
		err := errors.New("regular error")
		LogError(logger, err, WithLogMessage("fail"))

		output := buf.String()
		if !strings.Contains(output, "msg=fail") {
			t.Error("expected custom message for regular error")
		}
		if !strings.Contains(output, "error=\"regular error\"") {
			t.Error("missing error in log")
		}
	})
}
