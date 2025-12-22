package process

import (
	"reflect"
	"strconv"
)

func NewIntHandler(fieldType reflect.Type) Handler {
	return TypeHandler[int64]{
		Parser: func(rawValue string) (int64, error) {
			// Use base 0 to allow input like 0xFF
			return strconv.ParseInt(rawValue, 0, fieldType.Bits())
		},
		ValidationWrapper: WrapProcessUsingRangeTags[int64],
	}
}

func NewUintHandler(fieldType reflect.Type) Handler {
	return TypeHandler[uint64]{
		Parser: func(rawValue string) (uint64, error) {
			return strconv.ParseUint(rawValue, 0, fieldType.Bits())
		},
		ValidationWrapper: WrapProcessUsingRangeTags[uint64],
	}
}

func NewFloatHandler(fieldType reflect.Type) Handler {
	return TypeHandler[float64]{
		Parser: func(rawValue string) (float64, error) {
			return strconv.ParseFloat(rawValue, fieldType.Bits())
		},
		ValidationWrapper: WrapProcessUsingRangeTags[float64],
	}
}
