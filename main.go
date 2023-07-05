package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
	"os"
	"sort"
	stevedore "stevedore/src"
	"stevedore/src/version"
	"time"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var file string

	var directory string

	app := &cli.App{
		EnableBashCompletion: true,
		Flags:                []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:      "version",
				Aliases:   []string{"v"},
				Usage:     "Outputs the application version",
				UsageText: "stevedore version",
				Action: func(*cli.Context) error {
					fmt.Println(version.Version)

					return nil
				},
			},
			{
				Name:      "label",
				Aliases:   []string{"l"},
				Usage:     "Updates Dockerfiles labels",
				UsageText: "stevedore label",
				Action: func(*cli.Context) error {
					err := stevedore.Parse()
					return err
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "Dockerfile to parse",
						Destination: &file,
						Category:    "files",
					},
					&cli.StringFlag{
						Name:        "directory",
						Aliases:     []string{"d"},
						Usage:       "Destination to update Dockerfiles",
						Value:       ".",
						Destination: &directory,
						Category:    "files",
					},
				},
			},
		},
		Name:     "stevedore",
		Usage:    "Update Dockerfile labels",
		Compiled: time.Time{},
		Authors:  []*cli.Author{{Name: "James Woolfenden", Email: "jim.wolf@duck.com"}},
		Version:  version.Version,
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("stevedore failure")
	}
}
