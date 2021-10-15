package terraform

import (
	_ "embed" // Embed kics CLI img and scan-flags
	"os"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser"
	jsonParser "github.com/Checkmarx/kics/pkg/parser/json"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	yamlParser "github.com/Checkmarx/kics/pkg/parser/yaml"
)

type terraformResource struct {
	Name       string
	Type       string
	Properties map[string]interface{}
}

func Parse(path string) ([]terraformResource, error) {
	var tfResources []terraformResource
	var tfResource terraformResource

	combinedParser, err := parser.NewBuilder().
		Add(&jsonParser.Parser{}).
		Add(&yamlParser.Parser{}).
		Add(terraformParser.NewDefault()).
		Build([]string{"Terraform"}, []string{"aws"})
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			panic(err)
		}

		for _, doc := range docs {
			if doc["resource"] != nil {
				for resourceType, resources := range doc["resource"].(model.Document) {
					tfResource.Type = resourceType
					for resourceName, resourceData := range resources.(model.Document) {
						tfResource.Name = resourceName
						tfResource.Properties = make(map[string]interface{})
						for k, v := range resourceData.(model.Document) {
							// Avoid adding properties like _kics_lines for now
							if !strings.HasPrefix(k, "_kics") {
								tfResource.Properties[k] = v
							}
						}
					}
					tfResources = append(tfResources, tfResource)
				}
			}
		}
	}
	return tfResources, nil
}
