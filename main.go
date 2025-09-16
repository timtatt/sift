package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/timtatt/sift/internal/sift"
)

func main() {

	// TODO: setup logger

	ctx := context.Background()

	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug view")
	flag.Parse()

	err := sift.Run(ctx, sift.SiftOptions{
		Debug: debug,
	})

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

}
