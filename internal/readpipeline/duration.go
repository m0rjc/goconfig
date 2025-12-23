package readpipeline

import "time"

var durationTypeHandler = NewTypedDurationHandler()

// NewTypedDurationHandler returns a TypedHandler[time.Duration] that uses standard duration parsing and validation.
func NewTypedDurationHandler() TypedHandler[time.Duration] {
	return &typeHandlerImpl[time.Duration]{
		Parser: func(rawValue string) (time.Duration, error) {
			return time.ParseDuration(rawValue)
		},
		ValidationWrapper: WrapProcessUsingRangeTags[time.Duration],
	}
}
