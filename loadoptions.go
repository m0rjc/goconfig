package goconfig

import "reflect"

// loadOptions holds the configuration options for Load.
type loadOptions struct {
	// parsers allows the parser for a given key to be overridden
	parsers map[string]Parser // key is fieldPath
	// keyStore reads the values. Default to os.GetEnv()
	keyStore KeyStore
	// validatorFactories provide validators for a field
	validatorFactories []ValidatorFactory
	// validators maps field paths to their validator functions
	validators map[string][]Validator
}

func (opts *loadOptions) addValidator(fieldPath string, validator Validator) {
	if opts.validators == nil {
		opts.validators = make(map[string][]Validator)
	}
	opts.validators[fieldPath] = append(opts.validators[fieldPath], validator)
}

func (opts *loadOptions) addValidatorFactory(factory ValidatorFactory) {
	if opts.validatorFactories == nil {
		opts.validatorFactories = make([]ValidatorFactory, 0, 1)
	}
	opts.validatorFactories = append(opts.validatorFactories, factory)
}

func (opts *loadOptions) addParser(path string, parser Parser) {
	if opts.parsers == nil {
		opts.parsers = make(map[string]Parser)
	}
	opts.parsers[path] = parser
}

func (opts *loadOptions) getCustomParser(path string) Parser {
	return opts.parsers[path]
}

func (opts *loadOptions) getCustomValidators(path string, fieldType reflect.StructField) ([]Validator, error) {
	validators := make([]Validator, 0)
	supplied, ok := opts.validators[path]
	if ok {
		validators = append(validators, supplied...)
	}

	if opts.validatorFactories != nil {
		registry := func(v Validator) {
			validators = append(validators, v)
		}
		for _, factory := range opts.validatorFactories {
			if err := factory(fieldType, registry); err != nil {
				return nil, err
			}
		}
	}

	return validators, nil
}
