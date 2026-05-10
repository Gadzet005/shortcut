package shortcutclient

import (
	"net/http"
	"net/url"
	"time"
)

type requestOptions struct {
	method  string
	headers http.Header
	query   url.Values
	timeout time.Duration
}

func defaultRequestOptions() requestOptions {
	return requestOptions{
		method:  http.MethodPost,
		headers: http.Header{},
		query:   url.Values{},
	}
}

type Option func(*requestOptions)

func WithMethod(method string) Option {
	return func(o *requestOptions) {
		o.method = method
	}
}

func WithHeader(key, value string) Option {
	return func(o *requestOptions) {
		o.headers.Set(key, value)
	}
}

func WithHeaders(headers http.Header) Option {
	return func(o *requestOptions) {
		for k, values := range headers {
			for _, v := range values {
				o.headers.Add(k, v)
			}
		}
	}
}

func WithQuery(key, value string) Option {
	return func(o *requestOptions) {
		o.query.Add(key, value)
	}
}

func WithQueryValues(values url.Values) Option {
	return func(o *requestOptions) {
		for k, vs := range values {
			for _, v := range vs {
				o.query.Add(k, v)
			}
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *requestOptions) {
		o.timeout = d
	}
}

func WithNodeOverride(node, hostPort string) Option {
	return func(o *requestOptions) {
		o.query.Add("node-rwr", node+":"+hostPort)
	}
}
