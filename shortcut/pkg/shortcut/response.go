package shortcut

import (
	"bytes"
	"encoding/json"
	"net/http"

	multipartutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/multipart"
)

type Response struct {
	items map[string][]byte
	err   error
}

func NewResponse() *Response {
	return &Response{
		items: make(map[string][]byte),
	}
}

func (r *Response) AddItem(name string, data []byte) *Response {
	r.items[name] = data
	return r
}

func (r *Response) AddItemJSON(name string, v any) *Response {
	data, err := json.Marshal(v)
	if err != nil {
		r.err = NewErrorWithCause(http.StatusInternalServerError, "failed to marshal json", err)
		return r
	}

	r.AddItem(name, data)
	return r
}

func (r *Response) Send(c *Context) error {
	if r.err != nil {
		return r.err
	}

	var buf bytes.Buffer

	contentType, err := multipartutils.WriteMultipartData(&buf, r.items)
	if err != nil {
		r.err = NewErrorWithCause(http.StatusInternalServerError, "failed to write multipart data", err)
		return r.err
	}

	c.base.Data(http.StatusOK, contentType, buf.Bytes())
	return nil
}
