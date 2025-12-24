package config

import (
	"errors"
	"net/url"

	"github.com/m0rjc/goconfig"
)

// SecureURL is a custom type based on url.URL that represents a URL restricted to secure HTTPS schemes only.
type SecureURL url.URL

// ErrMustBeSecureUrl is returned when a URL does not use the HTTPS scheme, indicating it must be secure.
var ErrMustBeSecureUrl = errors.New("must be a secure URL")

// String returns the string representation of the SecureURL.
// This is required to allow printing
func (s *SecureURL) String() string {
	if s == nil {
		return ""
	}
	u := url.URL(*s)
	return u.String()
}

// typeUrlPtr is a custom type parser for URL values.
// Start with the base string type so that we gain the built-in regex validator
var typeUrlPtr = goconfig.TransformCustomType[string, *url.URL](
	goconfig.DefaultStringType[string](),
	func(rawValue string) (*url.URL, error) {
		value, err := url.ParseRequestURI(rawValue)
		if err != nil {
			return nil, err
		}
		return value, nil
	})

// typeSecureUrlPtr adds a validator to ensure the URL is secure before casting it to SecureURL.
var typeSecureUrlPtr = goconfig.CastCustomType[*url.URL, *SecureURL](
	goconfig.AddValidators(typeUrlPtr, func(value *url.URL) error {
		if value.Scheme != "https" {
			return ErrMustBeSecureUrl
		}
		return nil
	}),
)
