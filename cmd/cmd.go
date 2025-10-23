package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/timtatt/sift/internal/sift"
)

type CLI struct {
	Debug          bool `name:"debug" short:"d" help:"enable debug view"`
	RawLogs        bool `name:"raw" short:"r" help:"disable prettified logs"`
	NonInteractive bool `name:"non-interactive" short:"n" help:"disable interactive mode"`
	Version        bool `name:"version" short:"v" help:"print version"`
}

func (c *CLI) Run() error {
	ctx := context.Background()

	if c.Version {
		fmt.Print(sift.Version)
		os.Exit(0)
	}

	return sift.Run(ctx, sift.SiftOptions{
		Debug:          c.Debug,
		NonInteractive: c.NonInteractive,
		PrettifyLogs:   !c.RawLogs,
	})
}
