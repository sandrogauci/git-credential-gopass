package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/urfave/cli/v2"
)

const (
	name = "gopass-git-credentials"
)

// Version is the released version of gopass.
var version string

func main() {
	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context.
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	// reading from stdin?
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
		ctx = ctxutil.WithInteractive(ctx, false)
		ctx = ctxutil.WithStdin(ctx, true)
	}

	gp, err := api.New(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize gopass API: %s\n", err)
		os.Exit(1)
	}

	gc := &gc{
		gp: gp,
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = `Use "gopass" as git's credential.helper`
	app.Description = "" +
		"This command allows you to cache your git-credentials with gopass." +
		"Activate by using `git config --global credential.helper gopass`"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "store",
			Usage: "First part of path to find the secret.",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:   "get",
			Hidden: true,
			Action: gc.Get,
			Before: gc.Before,
		},
		{
			Name:   "store",
			Hidden: true,
			Action: gc.Store,
			Before: gc.Before,
		},
		{
			Name:   "erase",
			Hidden: true,
			Action: gc.Erase,
			Before: gc.Before,
		},
		{
			Name:        "configure",
			Description: "This command configures git-credential-gopass as git's credential.helper",
			Action:      gc.Configure,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "global",
					Usage: "Install for current user",
				},
				&cli.BoolFlag{
					Name:  "local",
					Usage: "Install for current repository only",
				},
				&cli.BoolFlag{
					Name:  "system",
					Usage: "Install for all users, requires superuser rights",
				},
				&cli.StringFlag{
					Name:  "store",
					Usage: "First part of path to find the secret.",
				},
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
