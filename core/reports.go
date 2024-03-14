package core

import (
	"fmt"
	"io"
)

type Report struct {
	Result
	testStatus bool
	testType   string
	testErrors []error
}

func (r *Report) Fprint(out io.Writer) {
	p := func(format string, args ...any) {
		fmt.Fprintf(out, format+"\n", args...)
	}
	testStatus := "FAILURE"
	if r.testStatus {
		testStatus = "SUCCESS"
	}
	p("Test [%s]", r.Url)
	p("\tTest Status: %s", testStatus)
	p("\tTest Type: %s", r.testType)
	p("\tTest Errors: %v", r.testErrors)
	p("\tDuration: %s", r.Duration)
}

func (r *Report) Success() bool {
	return r.testStatus
}
