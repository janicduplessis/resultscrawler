package test

import (
	"io"
	"net/http"
)

type FakeClient struct {
	Data io.ReadCloser
}

func (c *FakeClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body: c.Data,
	}, nil
}
