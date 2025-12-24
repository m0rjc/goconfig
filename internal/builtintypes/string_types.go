package builtintypes

import (
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

// NewStringHandler returns a TypedHandler[string] that simply returns the raw value.
// Strings support the min and max tags for lexical ordering and the pattern tag for regex
func NewStringHandler(_ reflect.Type) readpipeline.TypedHandler[string] {
	return NewTypedStringHandler()
}

// NewTypedStringHandler returns a TypedHandler[string] that uses standard string parsing and validation.
func NewTypedStringHandler() readpipeline.TypedHandler[string] {
	return &typeHandlerImpl[string]{
		Parser: func(rawValue string) (string, error) {
			return rawValue, nil
		},
		ValidationWrapper: readpipeline.NewCompositeWrapper(WrapProcessUsingPatternTag, WrapProcessUsingRangeTags[string]),
	}
}
