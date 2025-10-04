package main

import (
	"github.com/alecthomas/kong"
	"github.com/timtatt/sift/cmd"
)

func main() {
	var cli cmd.CLI

	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
