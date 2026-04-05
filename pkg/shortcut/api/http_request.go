package shortcutapi

import "net/url"

// TODO: сделать proto файл

type HttpRequest struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Query   url.Values          `json:"query"`
	Body    []byte              `json:"body"`
}
