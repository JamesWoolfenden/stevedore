package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	stevedore "github.com/jameswoolfenden/stevedore/src"
	"github.com/jameswoolfenden/stevedore/src/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"moul.io/banner"
)

func main() {
	fmt.Println(banner.Inline("stevedore"))

	fmt.Println("version:", version.Version)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var content stevedore.Parser

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
					var err error
					if content.File == nil {
						err = content.ParseAll()
					} else {
						err = content.ParseAll()
					}

					return err
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "Dockerfile to parse",
						Destination: content.File,
						Category:    "files",
					},
					&cli.StringFlag{
						Name:        "directory",
						Aliases:     []string{"d"},
						Usage:       "Destination to update Dockerfiles",
						Value:       ".",
						Destination: &content.Directory,
						Category:    "files",
					},
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       "Destination for updated Dockerfiles",
						Value:       ".",
						Destination: &content.Output,
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
