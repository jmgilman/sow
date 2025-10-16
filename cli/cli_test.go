// Package main provides integration tests for the sow CLI using testscript.
package main

import (
	"os"
	"testing"

	"github.com/jmgilman/sow/cli/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"sow": func() {
			// Execute the sow CLI
			// We need to handle errors gracefully since testscript
			// expects the return code to indicate success/failure
			rootCmd := cmd.NewRootCmd()
			if err := rootCmd.Execute(); err != nil {
				// Print error to stderr since SilenceErrors is true
				rootCmd.PrintErrln("Error:", err)
				os.Exit(1)
			}
		},
	})
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
