package builtintypes

import (
	"reflect"
	"strconv"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

func NewIntHandler(fieldType reflect.Type) readpipeline.TypedHandler[int64] {
	return NewTypedIntHandler(fieldType.Bits())
}

func NewUintHandler(fieldType reflect.Type) readpipeline.TypedHandler[uint64] {
	return NewTypedUintHandler(fieldType.Bits())
}

func NewFloatHandler(fieldType reflect.Type) readpipeline.TypedHandler[float64] {
	return NewTypedFloatHandler(fieldType.Bits())
}

// NewTypedIntHandler returns a TypedHandler[int64] that uses standard int parsing and validation.
func NewTypedIntHandler(bits int) readpipeline.TypedHandler[int64] {
	return &typeHandlerImpl[int64]{
		Parser: func(rawValue string) (int64, error) {
			return strconv.ParseInt(rawValue, 0, bits)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[int64],
	}
}

// NewTypedUintHandler returns a TypedHandler[uint64] that uses standard uint parsing and validation.
func NewTypedUintHandler(bits int) readpipeline.TypedHandler[uint64] {
	return &typeHandlerImpl[uint64]{
		Parser: func(rawValue string) (uint64, error) {
			return strconv.ParseUint(rawValue, 0, bits)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[uint64],
	}
}

// NewTypedFloatHandler returns a TypedHandler[float64] that uses standard float parsing and validation.
func NewTypedFloatHandler(bits int) readpipeline.TypedHandler[float64] {
	return &typeHandlerImpl[float64]{
		Parser: func(rawValue string) (float64, error) {
			return strconv.ParseFloat(rawValue, bits)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[float64],
	}
}
