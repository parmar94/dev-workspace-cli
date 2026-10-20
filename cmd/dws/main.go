package main

import (
	"os"

	"github.com/parmar94/dev-workspace-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
