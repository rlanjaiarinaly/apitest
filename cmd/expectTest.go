/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"apitest/core"
	"errors"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	ErrStatusCode = errors.New("Status Code mismatched")
)

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
	expects, err := core.ReadConfig(file)
	if err != nil {
		return err
	}
	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: runtime.NumCPU(),
		},
	}
	for report := range expects.CompareOutput(&client) {
		report.Fprint(out)
	}

	return nil
}
