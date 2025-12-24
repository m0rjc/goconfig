package builtintypes

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

// WrapProcessUsingPatternTag applies the pattern tag validation if present.
func WrapProcessUsingPatternTag(tags reflect.StructTag, processor readpipeline.FieldProcessor[string]) (readpipeline.FieldProcessor[string], error) {
	patternTag, hasPattern := tags.Lookup("pattern")

	if hasPattern {
		pattern, err := regexp.Compile(patternTag)
		if err != nil {
			return nil, err
		}
		return readpipeline.Pipe(processor, func(value string) error {
			if !pattern.MatchString(value) {
				return fmt.Errorf("does not match pattern %s", patternTag)
			}
			return nil
		}), nil
	}

	return processor, nil
}
