package shortcutapi

// TODO: сделать proto файл

type HttpResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}
