package main

import (
	"context"
	"fmt"
	"os"

	"github.com/timtatt/sift/internal/sift"
)

func main() {

	// TODO: setup logger

	ctx := context.Background()

	err := sift.Run(ctx)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

}
