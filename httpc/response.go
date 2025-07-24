package httpc

type Response[T any] struct {
	Body ResponseBody[T]
}

type ResponseBody[T any] struct {
	Data    T       `json:"data"`
	Paging  *Paging `json:"paging,omitempty"`
	Filter  *Filter `json:"filter,omitempty"`
	Sort    *Sort   `json:"sort,omitempty"`
	Message string  `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
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

// SetData sets the data field
func (b *ResponseBuilder[T]) SetData(data T) *ResponseBuilder[T] {
	b.response.Body.Data = data
	return b
}

// SetMessage sets the message field
func (b *ResponseBuilder[T]) SetMessage(message string) *ResponseBuilder[T] {
	b.response.Body.Message = message
	return b
}

// SetError sets the error field
func (b *ResponseBuilder[T]) SetError(error string) *ResponseBuilder[T] {
	b.response.Body.Error = error
	return b
}

func (b *ResponseBuilder[T]) SetPaging(page, size, total int64) *ResponseBuilder[T] {
	b.response.Body.Paging = &Paging{
		Page:  page,
		Size:  size,
		Total: total,
	}
	return b
}

func (b *ResponseBuilder[T]) SetFilter(modifier func(*Filter)) *ResponseBuilder[T] {
	if b.response.Body.Filter == nil {
		b.response.Body.Filter = &Filter{}
	}

	modifier(b.response.Body.Filter)

	return b
}

func (b *ResponseBuilder[T]) SetSort(modifier func(*Sort)) *ResponseBuilder[T] {
	if b.response.Body.Sort == nil {
		b.response.Body.Sort = &Sort{}
	}

	modifier(b.response.Body.Sort)
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
