package terraform

import (
	"context"
	"encoding/json"
	"fmt"
)

type TerraformPlanResource struct {
	Name    string                 `cty:"name"`
	Type    string                 `cty:"type"`
	Mode    string                 `cty:"mode"`
	Values  map[string]interface{} `cty:"values"`
	Address string                 `cty:"address"`
}

type TerraformPlanPlannedValuesRootModule struct {
	Resources []TerraformPlanResource `json:"resources"`
}

type TerraformPlanPlannedValues struct {
	RootModule TerraformPlanPlannedValuesRootModule `json:"root_module"`
}

type TerraformPlanContentStruct struct {
	PlannedValues TerraformPlanPlannedValues `json:"planned_values"`
}

func getTerraformPlanContentFromBytes(rawContent []byte) (*TerraformPlanContentStruct, error) {
	var planContent *TerraformPlanContentStruct
	err := json.Unmarshal(rawContent, &planContent)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the plan file content: %v", err)
	}
	return planContent, nil
}

func buildTerraformPlanResource(ctx context.Context, path string, resource TerraformPlanResource) (*terraformResource, error) {
	tfResource := new(terraformResource)

	tfResource.Path = path
	tfResource.Type = resource.Type
	tfResource.Name = resource.Name
	tfResource.Address = resource.Address
	tfResource.Mode = resource.Mode
	tfResource.Arguments = resource.Values
	tfResource.AttributesStd = tfResource.Arguments

	startLine, endLine, source := findBlockLinesFromJSON(ctx, path, "resources", resource.Address, resource.Type)
	tfResource.StartLine = startLine
	tfResource.EndLine = endLine
	tfResource.Source = source

	return tfResource, nil
}
