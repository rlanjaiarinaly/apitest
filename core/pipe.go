package core

import (
	"net/http"
	"sync"
)

// The type of function to send the request to the server and return the result
type SendFunc func(*http.Request) *Result

// Produce is the first entrypoint, it takes as argument :
//   - a channel where to send the request produced
//   - the number of request to produce
//     a function fn responsible for producing the request
func Produce(out chan<- *http.Request, n int, fn func() *http.Request) {
	for ; n > 0; n-- {
		out <- fn()
	}
}

// Part of the orchestrator pattern in golang
func produce(n int, fn func() *http.Request) <-chan *http.Request {
	out := make(chan *http.Request)

	go func() {
		defer close(out)
		Produce(out, n, fn)
	}()

	return out
}

func Split(in <-chan *http.Request, out chan<- *Result, concurrency int, fn SendFunc) {
	send := func() {
		for r := range in {
			out <- fn(r)
		}
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Create <concurrency> goroutines which listen for request from the "in" channel
	for ; concurrency > 0; concurrency-- {
		go func() {
			defer wg.Done()
			send()
		}()
	}
	wg.Wait()

}

func split(in <-chan *http.Request, concurrency int, fn SendFunc) <-chan *Result {
	out := make(chan *Result)

	go func() {
		defer close(out)
		Split(in, out, concurrency, fn)
	}()

	return out
}
