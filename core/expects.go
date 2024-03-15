package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrStatusCode = errors.New("Status Code mismatched")
	ErrBodyEmpty  = errors.New("Body empty for a POST Request")
)

type ExpectS struct {
	ExpectS []Expect `yaml:"expects"`
}

type input struct {
	Url    string         `yaml:"url"`
	Method string         `yaml:"method"`
	Header []headerParams `yaml:"headers"`
	Params []headerParams `yaml:"params"`
	Body   string         `yaml:"body"`
}

type headerParams struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ExpectedOutput struct {
	StatusCode int `yaml:"statusCode"`
}

type Expect struct {
	Input           input          `yaml:"input"`
	ExpectedOutput  ExpectedOutput `yaml:"expectedOutput"`
	PerformanceTest ExpectedPerf   `yaml:"performanceTest"`
}

func ReadConfig(file io.Reader) (*ExpectS, error) {
	var expect ExpectS
	if file != nil {
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&expect); err != nil {
			return nil, err
		}
	}
	for i, v := range expect.ExpectS {
		if v.ExpectedOutput.StatusCode == 0 {
			expect.ExpectS[i].ExpectedOutput.StatusCode = 200
		}
		if v.PerformanceTest.Success > 100 {
			expect.ExpectS[i].PerformanceTest.Success = 100
		}
	}
	return &expect, nil
}

func (e *Expect) String() string {
	return fmt.Sprintf("INPUT:\n\turl: %s\n\tmethod: %s\nOUTPUT:\n\tstatusCode: %d", e.Input.Url, e.Input.Method, e.ExpectedOutput.StatusCode)
}

func (e *ExpectS) CompareOutput(client *http.Client, perfClient *Client) chan *Report {
	out := make(chan *Report, len(e.ExpectS))
	defer close(out)

	wg := sync.WaitGroup{}
	wg.Add(len(e.ExpectS))

	for _, expect := range e.ExpectS {
		go func(expect Expect) {
			defer wg.Done()
			out <- expect.compareOuput(client, perfClient)
			log.Println(expect.PerformanceTest)
		}(expect)
	}
	wg.Wait()
	return out
}

func (e *Expect) compareOuput(client *http.Client, perfClient *Client) *Report {
	req, err := e.createURL(e.Input.Url, e.Input.Method)
	if err != nil {
		return &Report{
			testErrors: []error{err},
		}
	}
	e.parseHeader(req)
	result := Send(client, req)

	report := &Report{testStatus: true, Result: *result}
	if matchErrors := e.matchTest(result); len(matchErrors) > 0 {
		report.testStatus = false
		report.testErrors = append(report.testErrors, matchErrors...)
	}
	if report.testStatus && e.ExpectedOutput.StatusCode == 200 {
		perfResult := e.PerformanceTest.makePerfTest(perfClient, req)
		if perfErrors := e.PerformanceTest.validate(perfResult); len(perfErrors) > 0 {
			report.testStatus = false
			report.testErrors = append(report.testErrors, perfErrors...)
		}
		report.Duration = perfResult.Duration
	}
	return report
}

func (e *Expect) createURL(baseURL, method string) (*http.Request, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = e.parseParams().Encode()
	method = strings.ToUpper(method)
	var body io.Reader = http.NoBody
	if method == "POST" || method == "PUT" || method == "PATCH" {
		// if e.Input.Body == "" {
		// 	return nil, fmt.Errorf("%q URL: %s", ErrBodyEmpty, u.String())
		// }
		body = bytes.NewBuffer([]byte(e.Input.Body))
	}
	req, err := http.NewRequest(method, u.String(), body)
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

func (e *ExpectedOutput) toMap() map[string]string {
	r := map[string]string{}
	rType := reflect.TypeOf(e)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	structVal := reflect.ValueOf(*e)
	for i := 0; i < rType.NumField(); i++ {
		fieldName := rType.Field(i).Name
		r[fieldName] = fmt.Sprint(structVal.FieldByName(fieldName))
	}
	return r
}

func (e *Expect) matchTest(result *Result) []error {
	errorsResult := []error{}
	expectedMap := e.ExpectedOutput.toMap()
	resultMap := result.ToMap()
	for k, v := range expectedMap {
		if v != resultMap[k] {
			errorsResult = append(errorsResult, fmt.Errorf("Mismatched %s : got %s, want %s", k, resultMap[k], v))
		}
	}
	return errorsResult
}
