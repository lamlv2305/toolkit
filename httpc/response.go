package httpc

type Response[T any] struct {
	Body ResponseBody[T]
}

type ResponseBody[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ResponseBuilder provides a fluent interface for building responses
type ResponseBuilder[T any] struct {
	response Response[T]
}

// NewResponse creates a new ResponseBuilder
func NewResponse[T any]() *ResponseBuilder[T] {
	return &ResponseBuilder[T]{
		response: Response[T]{
			Body: ResponseBody[T]{},
		},
	}
}

// WithData sets the data field
func (b *ResponseBuilder[T]) WithData(data T) *ResponseBuilder[T] {
	b.response.Body.Data = data
	return b
}

// WithMessage sets the message field
func (b *ResponseBuilder[T]) WithMessage(message string) *ResponseBuilder[T] {
	b.response.Body.Message = message
	return b
}

// WithError sets the error field
func (b *ResponseBuilder[T]) WithError(error string) *ResponseBuilder[T] {
	b.response.Body.Error = error
	return b
}

// Build returns the final Response
func (b *ResponseBuilder[T]) Build() Response[T] {
	return b.response
}

func (b *ResponseBuilder[T]) Response() (*Response[T], error) {
	return &b.response, nil
}

type RawReponse struct {
	Body any
}
