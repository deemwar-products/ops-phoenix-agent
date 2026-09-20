package main

import (
	"os"

	"github.com/deemwar-products/sre-agent/internal/cli"
)

func main() {
	exitCode := cli.Run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
