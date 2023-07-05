package stevedore

import (
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

func Label(result *parser.Result) string {
	var label *parser.Node

	var endLine int

	var layer int64
	layer = 0

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
			child.Original += " layer." + strconv.FormatInt(layer, 10) + ".author = \"JamesWoolfenden\""
			label = child

			continue
		}
	}

	if label == nil {
		var newLabel parser.Node
		newLabel.Value = "LABEL"
		newLabel.Original = "LABEL layer." + strconv.FormatInt(layer, 10) + ".author = \"JamesWoolfenden\""
		newLabel.StartLine = endLine + 1
		newLabel.EndLine = endLine + 1

		var child parser.Node

		result.AST.AddChild(&child, newLabel.StartLine, newLabel.StartLine)
		result.AST.Children = append(result.AST.Children, &newLabel)
	}

	var dump string

	for _, child := range result.AST.Children {
		dump += child.Original + "\n"
	}

	return dump
}
