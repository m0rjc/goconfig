package customtypes

import (
	"errors"
	"testing"
)

func TestNewValidatorWrapper(t *testing.T) {
	parser := func(rawValue string) (int, error) {
		if rawValue == "1" {
			return 1, nil
		}
		if rawValue == "2" {
			return 2, nil
		}
		return 0, errors.New("parse error")
	}

	t.Run("SingleValidator", func(t *testing.T) {
		validator := func(val int) error {
			if val == 1 {
				return nil
			}
			return errors.New("must be 1")
		}

		wrapper := NewValidatorWrapper(validator)
		pipeline, err := wrapper("", parser)
		if err != nil {
			t.Fatalf("wrapper failed: %v", err)
		}

		if _, err := pipeline("1"); err != nil {
			t.Errorf("expected success for 1, got %v", err)
		}
		if _, err := pipeline("2"); err == nil {
			t.Error("expected error for 2, got nil")
		}
	})

	t.Run("MultipleValidators", func(t *testing.T) {
		v1 := func(val int) error {
			if val > 0 {
				return nil
			}
			return errors.New("must be positive")
		}
		v2 := func(val int) error {
			if val < 2 {
				return nil
			}
			return errors.New("must be less than 2")
		}

		wrapper := NewValidatorWrapper(v1, v2)
		pipeline, err := wrapper("", parser)
		if err != nil {
			t.Fatalf("wrapper failed: %v", err)
		}

		if _, err := pipeline("1"); err != nil {
			t.Errorf("expected success for 1, got %v", err)
		}
		if _, err := pipeline("2"); err == nil {
			t.Error("expected error for 2, got nil")
		}
	})

	t.Run("No validators", func(t *testing.T) {
		wrapper := NewValidatorWrapper[int]()
		pipeline, err := wrapper("", parser)
		if err != nil {
			t.Fatalf("wrapper failed: %v", err)
		}

		if val, err := pipeline("1"); err != nil || val != 1 {
			t.Errorf("expected 1, got %v, %v", val, err)
		}
	})
}
