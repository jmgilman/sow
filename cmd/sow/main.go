package main

import (
	"os"

	"github.com/your-org/sow/internal/commands"
)

func main() {
	rootCmd := commands.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
