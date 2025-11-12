package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/timtatt/sift/cmd"
	"github.com/timtatt/sift/internal/sift"
)

func main() {
	var cli cmd.CLI

	ctx := kong.Parse(&cli)

	if sift.IsStdinTerminal() {
		fmt.Fprintln(os.Stderr, "Warning: no input provided")
		fmt.Fprintln(os.Stderr)
		ctx.PrintUsage(false)
		os.Exit(1)
	}

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
