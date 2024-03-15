package core

import (
	"testing"
	"time"
)

func TestToMap(t *testing.T) {
	testCases := []struct {
		In   Result
		Out  map[string]string
		Name string
	}{
		{
			Result{
				Url:        "https://www.google.com",
				RPS:        1.0,
				Requests:   10,
				Errors:     0,
				Bytes:      123,
				Duration:   time.Duration(time.Second),
				Fastest:    time.Duration(time.Second),
				Slowest:    time.Duration(time.Second),
				StatusCode: 200,
				Error:      nil,
			},
			map[string]string{
				"Url":        "https://www.google.com",
				"RPS":        "1",
				"Requests":   "10",
				"Errors":     "0",
				"Bytes":      "123",
				"Duration":   "1s",
				"Fastest":    "1s",
				"Slowest":    "1s",
				"StatusCode": "200",
				"Error":      "",
			},
			"Normal TestCase",
		},
	}
	for _, v := range testCases {
		t.Run(v.Name, func(t *testing.T) {
			got := v.In.ToMap()
			for k, value := range got {
				if wantV, ok := v.Out[k]; ok && wantV != v.Out[k] {
					t.Errorf("Error %s : got %s, want %s", k, value, v.Out[k])
				}
			}
		})
	}
}
