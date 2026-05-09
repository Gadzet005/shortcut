package shortcutclient

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  int
	Headers http.Header
	Body    []byte
}

func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body, v)
}
