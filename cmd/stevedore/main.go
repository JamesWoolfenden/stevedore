package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jameswoolfenden/stevedore/internal/auth"
	"github.com/jameswoolfenden/stevedore/internal/config"
	"github.com/jameswoolfenden/stevedore/internal/dockerfile"
	"github.com/jameswoolfenden/stevedore/internal/git"
	"github.com/jameswoolfenden/stevedore/src/version"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"moul.io/banner"
)

func main() {
	fmt.Println(banner.Inline("stevedore"))
	fmt.Println("version:", version.Version)

	// Initialize configuration
	cfg := config.NewConfig()
	if err := cfg.SetupLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

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
				UsageText: "stevedore label [options]",
				Action: func(c *cli.Context) error {
					return runLabel(c, cfg)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Dockerfile to parse",
						Category: "files",
					},
					&cli.StringFlag{
						Name:     "directory",
						Aliases:  []string{"d"},
						Usage:    "Directory to scan for Dockerfiles",
						Value:    ".",
						Category: "files",
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Destination for updated Dockerfiles",
						Value:    ".",
						Category: "files",
					},
					&cli.StringFlag{
						Name:     "author",
						Aliases:  []string{"a"},
						Usage:    "Override for author name",
						Value:    "",
						Category: "metadata",
					},
				},
			},
		},
		Name:     "stevedore",
		Usage:    "Update Dockerfile labels with metadata",
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

// runLabel executes the label command
func runLabel(c *cli.Context, cfg *config.Config) error {
	// Get flags
	file := c.String("file")
	directory := c.String("directory")
	output := c.String("output")
	author := c.String("author")

	// Override config with author if provided
	if author != "" {
		cfg.DefaultAuthor = author
	}

	// Initialize services
	authService := auth.NewDockerAuth()

	// Initialize git service (may be nil if not in a git repo)
	var gitService git.Service
	var err error

	workDir := directory
	if file != "" {
		workDir = file
	}

	gitService, err = git.NewGitService(workDir)
	if err != nil {
		log.Warn().Err(err).Msg("git service unavailable, will skip git metadata")
		gitService = nil
	}

	// Create labeler and parser
	labeler := dockerfile.NewLabeler(gitService, authService)
	parser := dockerfile.NewParser(labeler)
	parser.File = file
	parser.Directory = directory
	parser.Output = output
	parser.Author = cfg.DefaultAuthor

	// Execute parsing
	return parser.ParseAll()
}
