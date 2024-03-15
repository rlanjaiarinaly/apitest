package core

import (
	"fmt"
	"io"
	"strings"
)

type TestError []error

type Report struct {
	Result
	testStatus bool
	testType   string
	testErrors TestError
}

func (te *TestError) String() string {
	var r strings.Builder
	for _, err := range *te {
		fmt.Fprintln(&r, "\t\t- "+err.Error())
	}
	return strings.TrimRight(r.String(), "\n")
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
	if len(r.testErrors) > 0 {
		p("\tTest Errors:\n%s", r.testErrors.String())
	} else {
		p("\tTest Errors: <nil>")
	}
	p("\tDuration: %s", r.Duration)
}

func (r *Report) String() string {
	var s strings.Builder
	r.Fprint(&s)
	return s.String()
}

func (r *Report) Success() bool {
	return r.testStatus
}
