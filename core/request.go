package core

import (
	"io"
	"net/http"
	"time"
)

func Send(client *http.Client, r *http.Request) *Result {
	t := time.Now()
	var (
		code  int
		bytes int64
	)
	response, err := client.Do(r)
	code = response.StatusCode
	bytes, err = io.Copy(io.Discard, response.Body)
	return &Result{
		Duration: time.Since(t),
		Bytes:    bytes,
		Status:   code,
		Error:    err,
	}
}
