/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"apitest/core"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	ErrStatusCode = errors.New("Status Code mismatched")
)

type expectS struct {
	ExpectS []Expect `yaml:"expects"`
}

type input struct {
	Url    string `yaml:"url"`
	Method string `yaml:"method"`
}

type expectedOutput struct {
	StatusCode int `yaml:"statusCode"`
}

type Expect struct {
	Input          input          `yaml:"input"`
	ExpectedOutput expectedOutput `yaml:"expectedOutput"`
}

type Report struct {
	core.Result
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

// expectTestCmd represents the expectTest command
var expectTestCmd = &cobra.Command{
	Use:   "expectTest",
	Short: "Test route against some expected output",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filepath, err := cmd.Flags().GetString("filepath")
		if err != nil {
			return err
		}
		return testExpectedRouteAction(os.Stdout, filepath)
	},
}

func init() {
	rootCmd.AddCommand(expectTestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// expectTestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// expectTestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	expectTestCmd.Flags().StringP("filepath", "f", "specs.yaml", "Path of the file containing the input and expected output")
}

func testExpectedRouteAction(out io.Writer, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	expects, err := readConfig(file)
	if err != nil {
		return err
	}
	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: runtime.NumCPU(),
		},
	}
	for report := range expects.compareOutput(&client) {
		report.Fprint(out)
	}

	return nil
}

func readConfig(file io.Reader) (*expectS, error) {
	var expect expectS
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

func (e *expectS) compareOutput(client *http.Client) chan *Report {
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
	req, err := http.NewRequest(e.Input.Method, e.Input.Url, http.NoBody)
	if err != nil {
		return &Report{
			testErrors: []error{err},
		}
	}
	result := core.Send(client, req)
	report := &Report{testStatus: true, Result: *result}
	if result.StatusCode != e.ExpectedOutput.StatusCode {
		report.testStatus = false
		report.testErrors = append(report.testErrors, fmt.Errorf("%q : got %d expected %d", ErrStatusCode, result.StatusCode, e.ExpectedOutput.StatusCode))
	}
	return report
}
