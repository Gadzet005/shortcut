package shortcutclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func New(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

func (c *Client) Run(ctx context.Context, graph string, body []byte, opts ...Option) (*Response, error) {
	if c.cfg.BaseURL == "" {
		return nil, fmt.Errorf("shortcutclient: empty BaseURL")
	}
	if graph == "" {
		return nil, fmt.Errorf("shortcutclient: empty graph name")
	}

	o := defaultRequestOptions()
	for _, opt := range opts {
		opt(&o)
	}

	endpoint, err := buildURL(c.cfg.BaseURL, graph, o.query)
	if err != nil {
		return nil, fmt.Errorf("shortcutclient: build url: %w", err)
	}

	if o.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, o.method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("shortcutclient: new request: %w", err)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, values := range o.headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("shortcutclient: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("shortcutclient: read response: %w", err)
	}

	return &Response{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    respBody,
	}, nil
}

func buildURL(baseURL, graph string, query url.Values) (string, error) {
	base := strings.TrimRight(baseURL, "/")
	graphPath := strings.TrimLeft(graph, "/")

	endpoint := base + "/run/" + graphPath

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}
