package stevedore

import (
	"os/user"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rs/zerolog/log"
)

func Label(result *parser.Result, file *string) string {
	var label *parser.Node

	var endLine int

	var layer int64
	layer = 0

	myUser, _ := user.Current()

	for _, child := range result.AST.Children {
		endLine = child.EndLine

		if strings.Contains(child.Value, "FROM") {
			SplitFrom := strings.SplitN(child.Original, "FROM", 2)
			image := strings.TrimSpace(SplitFrom[1])
			ParentLabel, err := GetDockerLabels(image)
			if err != nil {
				log.Info().Msgf("label error: %s", err)
			}

			for pLabel := range ParentLabel {
				if strings.Contains(pLabel, "layer") {
					splitter := strings.Split(pLabel, ".")
					version, _ := strconv.Atoi(splitter[1])
					layer = int64(version) + 1

					break
				}
			}
		}

		if strings.Contains(child.Value, "LABEL") {
			label = MakeLabel(child, layer, myUser, endLine, file)
		}
	}

	if label == nil {
		var newLabel parser.Node

		MakeLabel(&newLabel, layer, myUser, endLine, file)

		result.AST.Children = append(result.AST.Children, &newLabel)
	}

	var dump string

	for _, child := range result.AST.Children {
		dump += child.Original + "\n"
	}

	return dump
}

func MakeLabel(child *parser.Node, layer int64, myUser *user.User, endLine int, file *string) *parser.Node {
	myLayer := " layer." + strconv.FormatInt(layer, 10)
	if strings.Contains(child.Value, "LABEL") {
		child.Original = child.Original + myLayer +
			".author=" + "\"" + myUser.Name + "\"" + myLayer + ".trace=\"" + uuid.NewString() + "\""
	} else {
		child.Original = "LABEL" + myLayer +
			".author=" + "\"" + myUser.Name + "\"" + myLayer + ".trace=\"" + uuid.NewString() + "\""
	}

	child.Original += myLayer + ".tool=\"stevedore\""
	child.StartLine = endLine + 1
	child.EndLine = endLine + 1

	log.Info().Msgf("file: " + *file)
	log.Info().Msgf("label: " + child.Original)

	return child
}
