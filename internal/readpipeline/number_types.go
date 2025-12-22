package readpipeline

import (
	"reflect"
	"strconv"
)

func NewIntHandler(fieldType reflect.Type) PipelineBuilder {
	return WrapTypedHandler(NewTypedIntHandler(fieldType.Bits()))
}

func NewUintHandler(fieldType reflect.Type) PipelineBuilder {
	return WrapTypedHandler(NewTypedUintHandler(fieldType.Bits()))
}

func NewFloatHandler(fieldType reflect.Type) PipelineBuilder {
	return WrapTypedHandler(NewTypedFloatHandler(fieldType.Bits()))
}

// NewTypedIntHandler returns a TypedHandler[int64] that uses standard int parsing and validation.
func NewTypedIntHandler(bits int) TypedHandler[int64] {
	return typeHandlerImpl[int64]{
		Parser: func(rawValue string) (int64, error) {
			return strconv.ParseInt(rawValue, 0, bits)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[int64],
	}
}

// NewTypedUintHandler returns a TypedHandler[uint64] that uses standard uint parsing and validation.
func NewTypedUintHandler(bits int) TypedHandler[uint64] {
	return typeHandlerImpl[uint64]{
		Parser: func(rawValue string) (uint64, error) {
			return strconv.ParseUint(rawValue, 0, bits)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[uint64],
	}
}

// NewTypedFloatHandler returns a TypedHandler[float64] that uses standard float parsing and validation.
func NewTypedFloatHandler(bits int) TypedHandler[float64] {
	return typeHandlerImpl[float64]{
		Parser: func(rawValue string) (float64, error) {
			return strconv.ParseFloat(rawValue, bits)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[float64],
	}
}
