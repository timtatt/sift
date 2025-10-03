package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/timtatt/sift/internal/sift"
)

func main() {

	ctx := context.Background()

	var debug bool
	var nonInteractive bool
	flag.BoolVar(&debug, "debug", false, "enable debug view")
	flag.BoolVar(&nonInteractive, "non-interactive", false, "skip alternate screen and show inline view only")
	flag.BoolVar(&nonInteractive, "n", false, "skip alternate screen and show inline view only (shorthand)")
	flag.Parse()

	err := sift.Run(ctx, sift.SiftOptions{
		Debug:          debug,
		NonInteractive: nonInteractive,
	})

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

}
