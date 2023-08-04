package stevedore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type Parser struct {
	File      *string
	Output    string
	Directory string
}

func (content *Parser) ParseAll() error {
	if content.File != nil {
		err2 := content.Parse()
		if err2 != nil {
			return err2
		}
	} else {
		err := filepath.Walk(content.Directory,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() && strings.Contains(info.Name(), "Dockerfile") {
					file := info.Name()
					content.File = &file
					err2 := content.Parse()
					if err2 != nil {
						return err2
					}
				}

				return nil
			})
		if err != nil {
			return fmt.Errorf("walk path failed %w", err)
		}
	}

	return nil
}

func (content *Parser) Parse() error {
	Parse := Dockerfile{nil, *content.File, ""}

	err := Parse.ParseFile()
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	dump, err := Parse.Label()
	if err != nil {
		return err
	}

	fileOut := filepath.Join(content.Output, filepath.Base(*content.File))

	err = os.WriteFile(fileOut, []byte(dump), 0o644)
	if err != nil {
		return fmt.Errorf("writefile error: %w", err)
	}

	log.Info().Msgf("updated: %s", "Dockerfile")

	return nil
}
