package readpipeline

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

func NewUrlTypedHandler() TypedHandler[*url.URL] {
	return &typeHandlerImpl[*url.URL]{
		Parser:            url.ParseRequestURI,
		ValidationWrapper: wrapUrlPipeline,
	}
}

func wrapUrlPipeline(tags reflect.StructTag, pipeline FieldProcessor[*url.URL]) (FieldProcessor[*url.URL], error) {
	patternTag := tags.Get("pattern")
	if patternTag != "" {
		pattern, err := regexp.Compile(patternTag)
		if err != nil {
			return nil, err
		}
		pipeline = Pipe(pipeline, func(value *url.URL) error {
			if !pattern.MatchString(value.String()) {
				return fmt.Errorf("does not match pattern %s", patternTag)
			}
			return nil
		})
	}

	// scheme is a command separated list of acceptable schemes, for example `http,https` or `imaps`
	schemeTag := tags.Get("scheme")
	if schemeTag != "" {
		schemes := strings.Split(schemeTag, ",")
		pipeline = Pipe(pipeline, func(value *url.URL) error {
			for _, scheme := range schemes {
				if scheme == value.Scheme {
					return nil
				}
			}
			return fmt.Errorf("scheme must be one of %s", strings.Join(schemes, ", "))
		})
	}

	return pipeline, nil
}
