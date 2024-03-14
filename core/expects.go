package core

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrStatusCode = errors.New("Status Code mismatched")
)

type ExpectS struct {
	ExpectS []Expect `yaml:"expects"`
}

type input struct {
	Url    string         `yaml:"url"`
	Method string         `yaml:"method"`
	Header []headerParams `yaml:"headers"`
	Params []headerParams `yaml:"params"`
}

type headerParams struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ExpectedOutput struct {
	StatusCode int `yaml:"statusCode"`
}

type Expect struct {
	Input          input          `yaml:"input"`
	ExpectedOutput ExpectedOutput `yaml:"expectedOutput"`
}

func ReadConfig(file io.Reader) (*ExpectS, error) {
	var expect ExpectS
	if file != nil {
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&expect); err != nil {
			return nil, err
		}
	}
	return &expect, nil
}

func (e *Expect) String() string {
	return fmt.Sprintf("INPUT:\n\turl: %s\n\tmethod: %s\nOUTPUT:\n\tstatusCode: %d", e.Input.Url, e.Input.Method, e.ExpectedOutput.StatusCode)
}

func (e *ExpectS) CompareOutput(client *http.Client) chan *Report {
	out := make(chan *Report, len(e.ExpectS))
	defer close(out)

	wg := sync.WaitGroup{}
	wg.Add(len(e.ExpectS))

	for _, expect := range e.ExpectS {
		go func(expect Expect) {
			defer wg.Done()
			out <- expect.compareOuput(client)
		}(expect)
	}
	wg.Wait()
	return out
}

func (e *Expect) compareOuput(client *http.Client) *Report {
	req, err := e.createURL(e.Input.Url, e.Input.Method)
	if err != nil {
		return &Report{
			testErrors: []error{err},
		}
	}
	e.parseHeader(req)
	result := Send(client, req)
	report := &Report{testStatus: true, Result: *result}
	if result.StatusCode != e.ExpectedOutput.StatusCode {
		report.testStatus = false
		report.testErrors = append(report.testErrors, fmt.Errorf("%q : got %d expected %d", ErrStatusCode, result.StatusCode, e.ExpectedOutput.StatusCode))
	}
	return report
}

func (e *Expect) createURL(baseURL, method string) (*http.Request, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = e.parseParams().Encode()
	req, err := http.NewRequest(method, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	e.parseHeader(req)
	return req, nil
}
func (e *Expect) parseHeader(r *http.Request) {
	for _, header := range e.Input.Header {
		r.Header.Set(header.Name, header.Value)
	}
}

func (e *Expect) parseParams() url.Values {
	parameters := url.Values{}
	for _, params := range e.Input.Params {
		parameters.Add(params.Name, params.Value)
	}
	return parameters
}
