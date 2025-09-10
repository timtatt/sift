package main

import (
	"context"
	"log"

	"github.com/timtatt/sift/internal/sift"
)

func main() {

	// TODO: setup logger

	ctx := context.Background()

	err := sift.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}

}
