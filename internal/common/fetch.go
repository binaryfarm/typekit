package common

import (
	"io"
	"net/http"
	u "net/url"
)

type FetchOptions struct {
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

type Response struct {
	Status     int                 `json:"status"`
	StatusText string              `json:"statusText"`
	Body       io.ReadCloser       `json:"body"`
	BodyUsed   bool                `json:"bodyUsed"`
	Headers    map[string][]string `json:"headers"`
	Ok         bool                `json:"ok"`
	Redirected bool                `json:"redirected"`
	Type       string              `json:"type"`
	Url        string              `json:"url"`
}

func Fetch(url string, opts FetchOptions) Response {
	location, e := u.Parse(url)
	if e != nil {
		return Response{
			StatusText: "500 Internal Server Error",
			Status:     500,
			BodyUsed:   false,
			Ok:         false,
			Redirected: false,
			Type:       "basic",
		}
	}
	response, e := http.DefaultClient.Do(&http.Request{
		Method: opts.Method,
		Header: opts.Headers,
		URL:    location,
	})

	return Response{
		StatusText: response.Status,
		Status:     response.StatusCode,
		BodyUsed:   response.Body != nil,
		Ok:         (response.StatusCode >= 200 && response.StatusCode <= 299),
		Type:       "basic",
		Headers:    response.Header,
		Url:        response.Request.URL.String(),
	}
}
