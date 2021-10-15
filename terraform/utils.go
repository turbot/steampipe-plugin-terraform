package terraform

import (
	_ "embed" // Embed kics CLI img and scan-flags
	"fmt"

	"github.com/Checkmarx/kics/pkg/parser"
	jsonParser "github.com/Checkmarx/kics/pkg/parser/json"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	yamlParser "github.com/Checkmarx/kics/pkg/parser/yaml"
)

func parse(data string) (string, error) {
	combinedParser, err := parser.NewBuilder().
		Add(&jsonParser.Parser{}).
		Add(&yamlParser.Parser{}).
		Add(terraformParser.NewDefault()).
		Build([]string{""}, []string{""})
	if err != nil {
		return "", err
	}

	for _, parser := range combinedParser {
		doc, kind, err := parser.Parse("test.json", nil)
		fmt.Println("Doc:", doc)
		fmt.Println("Kind:", kind)
		fmt.Println("Err:", err)
	}

	return "", nil
}
