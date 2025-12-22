package process

import "time"

var durationTypeHandler = TypeHandler[time.Duration]{
	Parser: func(rawValue string) (time.Duration, error) {
		return time.ParseDuration(rawValue)
	},
	ValidationWrapper: WrapProcessUsingRangeTags[time.Duration],
}
