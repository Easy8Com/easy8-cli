package skills

import (
	"embed"
	"fmt"
)

const PrimaryName = "easy8-cli"

var Names = []string{PrimaryName, "easy-query", "git-flow"}

//go:embed easy8-cli/SKILL.md
var Content []byte

//go:embed easy8-cli/SKILL.md easy-query/SKILL.md git-flow/SKILL.md
var bundle embed.FS

func Read(name string) ([]byte, error) {
	for _, item := range Names {
		if item == name {
			return bundle.ReadFile(fmt.Sprintf("%s/SKILL.md", name))
		}
	}
	return nil, fmt.Errorf("unknown skill: %s", name)
}
