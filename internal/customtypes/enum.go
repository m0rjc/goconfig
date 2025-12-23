package customtypes

import (
	"fmt"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

func NewStringEnum[T ~string](validValues ...T) readpipeline.TypedHandler[T] {
	return NewParser[T](func(rawValue string) (T, error) {
		for _, validValue := range validValues {
			if rawValue == string(validValue) {
				return validValue, nil
			}
		}
		return "", fmt.Errorf("invalid value: %s", rawValue)
	})
}
