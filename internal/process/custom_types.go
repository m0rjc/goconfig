package process

import "reflect"

// NewCustomHandler creates a custom handler that delegates to the custom parser. It will use the default handler
// for validation if present. customParser cannot be nil. defaultHandler can be nil.
func NewCustomHandler(customParser FieldProcessor[any], defaultHandler Handler) Handler {
	return &customHandler{
		customParser:   customParser,
		defaultHandler: defaultHandler,
	}
}

type customHandler struct {
	customParser     FieldProcessor[any]
	customValidators []Validator[any]
	defaultHandler   Handler
}

// GetParser returns the custom parser.
func (c customHandler) GetParser() FieldProcessor[any] {
	return c.customParser
}

// AddValidatorsToPipeline will apply validators from the default handler if present. It allows wholly custom types
// to be a no-op
func (c customHandler) AddValidatorsToPipeline(tags reflect.StructTag, p FieldProcessor[any]) (FieldProcessor[any], error) {
	if c.defaultHandler != nil {
		return c.defaultHandler.AddValidatorsToPipeline(tags, p)
	}
	return p, nil
}
