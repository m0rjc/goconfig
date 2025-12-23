package readpipeline

import (
	"cmp"
	"fmt"
	"reflect"
)

// orderedValidator is a validator that checks a value is within a range. The value must be comparable.
type orderedValidator[T cmp.Ordered] func(value T) error

func newMinValidator[T cmp.Ordered](minimum T) orderedValidator[T] {
	return func(value T) error {
		if value < minimum {
			return fmt.Errorf("below minimum %v", minimum)
		}
		return nil
	}
}

func newMaxValidator[T cmp.Ordered](maximum T) orderedValidator[T] {
	return func(value T) error {
		if value > maximum {
			return fmt.Errorf("above maximum %v", maximum)
		}
		return nil
	}
}

func newRangeValidator[T cmp.Ordered](minimum, maximum T) orderedValidator[T] {
	return func(value T) error {
		if value < minimum || value > maximum {
			return fmt.Errorf("must be between %v and %v", minimum, maximum)
		}
		return nil
	}
}

// WrapProcessUsingRangeTags applies the min and max tags to an ordered readpipeline.
func WrapProcessUsingRangeTags[T cmp.Ordered](tags reflect.StructTag, processor FieldProcessor[T]) (FieldProcessor[T], error) {
	minTag, hasMin := tags.Lookup("min")
	maxTag, hasMax := tags.Lookup("max")

	var minimum, maximum T
	var err error
	if hasMin {
		minimum, err = processor(minTag)
		if err != nil {
			return nil, fmt.Errorf("min tag: %v", err)
		}
	}
	if hasMax {
		maximum, err = processor(maxTag)
		if err != nil {
			return nil, fmt.Errorf("max tag: %v", err)
		}
	}

	if hasMin && hasMax {
		return Pipe(processor, Validator[T](newRangeValidator(minimum, maximum))), nil
	}
	if hasMin {
		return Pipe(processor, Validator[T](newMinValidator(minimum))), nil
	}
	if hasMax {
		return Pipe(processor, Validator[T](newMaxValidator(maximum))), nil
	}
	return processor, nil
}
