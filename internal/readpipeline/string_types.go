package readpipeline

import (
	"reflect"
)

// NewStringHandler returns a PipelineBuilder that simply returns the raw value.
// Strings support the min and max tags for lexical ordering and the pattern tag for regex
func NewStringHandler(_ reflect.Type) PipelineBuilder {
	return NewTypedStringHandler()
}

// NewTypedStringHandler returns a TypedHandler[string] that uses standard string parsing and validation.
func NewTypedStringHandler() TypedHandler[string] {
	return typeHandlerImpl[string]{
		Parser: func(rawValue string) (string, error) {
			return rawValue, nil
		},
		ValidationWrapper: NewCompositeWrapper(WrapProcessUsingPatternTag, WrapProcessUsingRangeTags[string]),
	}
}
