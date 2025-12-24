package config

import (
	"github.com/m0rjc/goconfig"
)

// init initializes and registers custom types for URL handling and validation, including secure HTTPS URL enforcement.
func init() {
	goconfig.RegisterCustomType(typeUrlPtr)
	goconfig.RegisterCustomType(typeSecureUrlPtr)
}
