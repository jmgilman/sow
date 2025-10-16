package main

import (
	"os"
	"testing"

	"github.com/jmgilman/sow/cli/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"sow": func() int {
			// Execute the sow CLI
			// We need to handle errors gracefully since testscript
			// expects the return code to indicate success/failure
			rootCmd := cmd.NewRootCmd()
			if err := rootCmd.Execute(); err != nil {
				// Print error to stderr since SilenceErrors is true
				rootCmd.PrintErrln("Error:", err)
				return 1
			}
			return 0
		},
	}))
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
