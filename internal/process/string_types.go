package process

import (
	"reflect"
)

// NewStringHandler returns a Handler that simply returns the raw value.
// Strings support the min and max tags for lexical ordering and the pattern tag for regex
func NewStringHandler(_ reflect.Type) Handler {
	return TypeHandler[string]{
		Parser: func(rawValue string) (value string, err error) {
			return rawValue, nil
		},
		ValidationWrapper: NewCompositeWrapper(WrapProcessUsingPatternTag, WrapProcessUsingRangeTags[string]),
	}
}
