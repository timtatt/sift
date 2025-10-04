package cmd

import (
	"context"

	"github.com/timtatt/sift/internal/sift"
)

type CLI struct {
	Debug          bool `name:"debug" short:"d" help:"enable debug view"`
	RawLogs        bool `name:"raw" short:"r" help:"disable prettified logs"`
	NonInteractive bool `name:"non-interactive" short:"n" help:"disable interactive mode"`
}

func (c *CLI) Run() error {
	ctx := context.Background()

	return sift.Run(ctx, sift.SiftOptions{
		Debug:          c.Debug,
		NonInteractive: c.NonInteractive,
		PrettifyLogs:   !c.RawLogs,
	})
}
