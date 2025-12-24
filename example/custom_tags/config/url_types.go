package config

import (
	"errors"
	"net/url"
	"reflect"

	"github.com/m0rjc/goconfig"
)

// ErrMustBeSecureUrl is returned when a URL does not use the HTTPS scheme, indicating it must be secure.
var ErrMustBeSecureUrl = errors.New("must be a secure URL")

// typeUrlPtr is a custom type parser for URL values.
// Start with the base string type so that we gain the built-in regex validator
func newUrlCustomType() goconfig.TypedHandler[*url.URL] {
	basicStringType := goconfig.DefaultStringType[string]()
	typeUrlPtr := goconfig.TransformCustomType[string, *url.URL](
		basicStringType,
		func(rawValue string) (*url.URL, error) {
			value, err := url.ParseRequestURI(rawValue)
			if err != nil {
				return nil, err
			}
			return value, nil
		})

	// Each if these methods creates a new type handler by decorating the existing one.
	// They do not modify the existing one.
	typeUrlPtr = goconfig.AddDynamicValidation(typeUrlPtr, func(tags reflect.StructTag, pipeline goconfig.FieldProcessor[*url.URL]) (goconfig.FieldProcessor[*url.URL], error) {
		secureTag := tags.Get("secure")
		if secureTag == "true" {
			pipeline = goconfig.AddValidatorToPipeline(pipeline, func(value *url.URL) error {
				if value.Scheme != "https" {
					return ErrMustBeSecureUrl
				}
				return nil
			})
		}
		return pipeline, nil
	})

	return typeUrlPtr
}
