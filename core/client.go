package core

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Client struct {
	Concurrency int // Concurrency Level
}

func (c *Client) client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: c.Concurrency,
		},
	}
}

func (c *Client) Do(r *http.Request, n int) *Result {
	t := time.Now()

	sum := c.do(r, n)

	return sum.Finalize(time.Since(t))
}

func (c *Client) do(r *http.Request, n int) *Result {
	var client = c.client()
	defer client.CloseIdleConnections()

	p := produce(n, func() *http.Request {

		reqCloned := r.Clone(context.Background())
		if r.Body != http.NoBody {
			if b, err := r.GetBody(); err == nil {
				reqCloned.Body = io.NopCloser(b)
			}
		}
		return reqCloned
	})

	var sum Result
	for result := range split(p, c.Concurrency, c.send(client)) {
		sum.Merge(result)
	}
	return &sum
}

func (c *Client) send(client *http.Client) SendFunc {
	return func(r *http.Request) *Result {
		return Send(client, r)
	}
}
