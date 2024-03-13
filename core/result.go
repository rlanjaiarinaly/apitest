package core

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Result struct {
	RPS         float64       // Request per second
	Requests    int           // Number of requests
	Errors      int           // Numbers of errors occuring
	Bytes       int64         // Number of bytes downloaded
	Duration    time.Duration // Duration of single or all requests
	Fastest     time.Duration // Fastest request duration
	Slowest     time.Duration
	AverageTime time.Duration // slowest request duration
	Status      int           // http status for a request
	Error       error         // not nil if the request presented some error
}

func (r *Result) Merge(o *Result) {
	r.Requests++
	r.Bytes += o.Bytes
	if r.Fastest == 0 || o.Duration < r.Fastest {
		r.Fastest = o.Duration
	}
	if o.Duration > r.Fastest {
		r.Slowest = o.Duration
	}
	switch {
	case o.Error != nil:
		fallthrough
	case o.Status >= http.StatusBadRequest:
		r.Errors++
	}
}

// Set the result total duration and the RPS as well
func (r *Result) Finalize(total time.Duration) *Result {
	r.Duration = total
	r.RPS = float64(r.Requests) / total.Seconds()
	r.AverageTime = total / time.Duration(r.Requests)
	return r
}

func (r *Result) Fprint(out io.Writer) {
	p := func(format string, args ...any) {
		fmt.Fprintf(out, format, args...)
	}
	p("\nSummary:\n")
	p("\tSuccess    : %.0f%%\n", r.success())
	p("\tRPS        : %.1f\n", r.RPS)
	p("\tRequests   : %d\n", r.Requests)
	p("\tErrors     : %d\n", r.Errors)
	p("\tBytes      : %d\n", r.Bytes)
	p("\tDuration   : %s\n", round(r.Duration))
	if r.Requests > 1 {
		p("\tFastest    : %s\n", round(r.Fastest))
		p("\tSlowest    : %s\n", round(r.Slowest))
		p("\tAverage    : %s\n", r.AverageTime)
	}
}

// Calculate the percentage of the successful request
func (r *Result) success() float64 {
	rr, e := float64(r.Requests), float64(r.Errors)
	return (rr - e) / rr * 100
}

func round(t time.Duration) time.Duration {
	return t.Round(time.Microsecond)
}

func (r *Result) String() string {
	var s strings.Builder
	r.Fprint(&s)
	return s.String()
}
