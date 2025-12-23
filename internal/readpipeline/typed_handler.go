package readpipeline

import "reflect"

// typeHandlerImpl is the strongly typed handler for the given pipeline.
// It implements the typeless PipelineBuilder interface for the pipeline by boxing and unboxing the value as required.
type typeHandlerImpl[T any] struct {
	// Parser is the strongly typed version of the FieldProcessor that acts as input for this readpipeline
	Parser FieldProcessor[T]
	// ValidationWrapper is a factory that wraps the FieldProcessor with validation stages
	ValidationWrapper Wrapper[T]
}

func (h *typeHandlerImpl[T]) BuildPipeline(tags reflect.StructTag) (FieldProcessor[T], error) {
	pipeline := h.Parser
	if pipeline == nil {
		return nil, nil
	}

	wrapper := h.ValidationWrapper
	if wrapper != nil {
		var err error
		pipeline, err = wrapper(tags, pipeline)
		if err != nil {
			return nil, err
		}
	}

	return pipeline, nil
}
