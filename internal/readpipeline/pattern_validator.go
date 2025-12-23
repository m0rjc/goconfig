package readpipeline

import (
	"fmt"
	"reflect"
	"regexp"
)

// WrapProcessUsingPatternTag applies the pattern tag validation if present.
func WrapProcessUsingPatternTag(tags reflect.StructTag, processor FieldProcessor[string]) (FieldProcessor[string], error) {
	patternTag, hasPattern := tags.Lookup("pattern")

	if hasPattern {
		pattern, err := regexp.Compile(patternTag)
		if err != nil {
			return nil, err
		}
		return Pipe(processor, func(value string) error {
			if !pattern.MatchString(value) {
				return fmt.Errorf("does not match pattern %s", patternTag)
			}
			return nil
		}), nil
	}

	return processor, nil
}
