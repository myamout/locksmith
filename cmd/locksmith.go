package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/myamout/locksmith/internal/server"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "port",
				Usage:    "Port locksmith server runs on",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			p := c.Int("port")
			cfg := server.ServerConfig{
				Port: p,
			}
			server := server.NewLocksmithServer(ctx, cfg)
			slog.Info("starting server...")
			err := server.Start()

			return err
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
