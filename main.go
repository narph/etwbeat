package main

import (
	"os"

	"github.com/narph/etwbeat/cmd"

	_ "github.com/narph/etwbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
