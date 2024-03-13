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
	"strings"

	"github.com/spf13/cobra"
)

var (
	ERR_METHOD_NOT_SUPPORTED = errors.New("GET is the only method supported for the moment")
)

// testrouteCmd represents the testroute command
var testrouteCmd = &cobra.Command{
	Use:   "testroute URL",
	Short: "Test a single route",
	Long: `Test a single route by specifying the : 
  - url
  - the flags with the request as described in the flags below`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.Usage()
			return nil
		}
		url := args[0]
		number, err := cmd.Flags().GetInt("number")
		if err != nil {
			return err
		}
		concurrency, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			return err
		}
		method, err := cmd.Flags().GetString("method")
		if err != nil {
			return err
		}
		err = testRouteAction(os.Stdout, url, method, concurrency, number)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testrouteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testrouteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testrouteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	testrouteCmd.Flags().IntP("number", "n", 1, "The number of request to send to the route")
	testrouteCmd.Flags().IntP("concurrency", "c", runtime.NumCPU(), "The concurrency level")
	testrouteCmd.Flags().StringP("method", "m", "GET", "Request Method")
}

func testRouteAction(out io.Writer, url string, method string, concurrency, number int) error {
	client := core.Client{
		Concurrency: concurrency,
	}
	method, err := handleMethod(method)
	if err != nil {
		return err
	}
	r, err := http.NewRequest(method, url, http.NoBody)
	if err != nil {
		return err
	}
	result := client.Do(r, number)
	fmt.Fprint(out, result)
	return nil
}

func handleMethod(method string) (string, error) {
	method = strings.ToUpper(method)
	switch method {
	case "GET":
	default:
		return "", ERR_METHOD_NOT_SUPPORTED
	}
	return method, nil
}
