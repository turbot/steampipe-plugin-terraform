package terraform

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	p "github.com/Checkmarx/kics/pkg/parser/json"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func tableTerraformVariable(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_variable",
		Description: "Terraform variable information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listVariables,
			KeyColumns:    plugin.OptionalColumns([]string{"path"}),
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The variable name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The variable type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "default_value",
				Description: "The default value for the variable.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "description",
				Description: "Because the variable values of a module are part of its user interface, you can briefly describe the purpose of each value using the optional description argument.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "sensitive",
				Description: "An variable can be marked as containing sensitive material using the optional sensitive argument.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "validation",
				Description: "The validation applied on the variable.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "end_line",
				Description: "Ending line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "source",
				Description: "The block source code.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "path",
				Description: "Path to the file.",
				Type:        proto.ColumnType_STRING,
			},
		},
	}
}

type terraformVariable struct {
	Name         string
	Type         string
	Path         string
	StartLine    int
	EndLine      int
	Source       string
	Description  string
	Sensitive    bool
	DefaultValue string
	Validation   string
}

func listVariables(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydrate, defaulting to the config paths or
	// available by the optional key column
	pathInfo := h.Item.(filePath)
	path := pathInfo.Path

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_variable.listVariables", "read_file_error", err, "path", path)
		return nil, err
	}

	// Return if the path is a TF plan path
	if pathInfo.IsTFPlanFilePath || isTerraformPlan(content) {
		return nil, nil
	}

	var docs []model.Document

	// Check if the file contains TF state
	if pathInfo.IsTFStateFilePath {
		// Initialize the JSON parser
		jsonParser := p.Parser{}

		// Parse the file content using the JSON parser
		var str string
		documents, _, err := jsonParser.Parse(str, content)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_variable.listVariables", "state_parse_error", err, "path", path)
			return nil, fmt.Errorf("failed to parse state file %s: %v", path, err)
		}

		docs = append(docs, documents...)
	} else {
		// Build the terraform parser
		combinedParser, err := Parser()
		if err != nil {
			plugin.Logger(ctx).Error("terraform_variable.listVariables", "create_parser_error", err)
			return nil, err
		}

		for _, parser := range combinedParser {
			parsedDocs, err := ParseContent(ctx, d, path, content, parser)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_variable.listVariables", "parse_error", err, "path", path)
				return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
			}
			docs = append(docs, parsedDocs.Docs...)
		}
	}

	for _, doc := range docs {
		if doc["variable"] != nil {
			// For each variable, scan its arguments
			for variableName, variableData := range doc["variable"].(model.Document) {
				tfVariable, err := buildVariable(ctx, pathInfo.IsTFStateFilePath, path, content, variableName, variableData.(model.Document))
				if err != nil {
					plugin.Logger(ctx).Error("terraform_variable.listVariables", "build_variable_error", err)
					return nil, err
				}
				d.StreamListItem(ctx, tfVariable)
			}
		} else if doc["variables"] != nil {
			// For each variable, scan its arguments
			for varName, variableData := range convertModelDocumentToMapInterface(doc["variables"]) {
				// if !strings.HasPrefix(varName, "_kics") {
				tfVar, err := buildVariable(ctx, pathInfo.IsTFStateFilePath, path, content, varName, convertModelDocumentToMapInterface(variableData))
				if err != nil {
					plugin.Logger(ctx).Error("terraform_variable.listVariables", "build_variable_error", err)
					return nil, err
				}
				d.StreamListItem(ctx, tfVar)
				// }
			}
		}
	}

	return nil, nil
}

func buildVariable(ctx context.Context, isTFStateFilePath bool, path string, content []byte, name string, d model.Document) (terraformVariable, error) {
	var tfVar terraformVariable

	tfVar.Path = path
	tfVar.Name = name

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	if isTFStateFilePath {
		startLine, endLine, source, err := findBlockLinesFromJSON(ctx, path, "variables", name)
		if err != nil {
			return tfVar, err
		}

		tfVar.StartLine = startLine
		tfVar.EndLine = endLine
		tfVar.Source = source
	} else {
		start, end, source, err := getBlock(ctx, path, content, "variable", []string{name})
		if err != nil {
			plugin.Logger(ctx).Error("terraform_variable.buildVariable", "getBlock", err)
			return tfVar, err
		}
		tfVar.StartLine = start.Line
		tfVar.EndLine = end.Line
		tfVar.Source = source
		val, err := extractValidationBlock(source)
		if err != nil {
			plugin.Logger(ctx).Debug("No validation block found...")
		} else {
			tfVar.Validation = val
		}
	}
	for k, v := range d {
		switch k {
		case "description":
			if reflect.TypeOf(v).String() != "string" {
				return tfVar, fmt.Errorf("the 'description' argument for variable '%s' must be of type string", name)
			}
			tfVar.Description = v.(string)

		case "default":
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_variable.buildVariable", "convert_value_error", err)
				return tfVar, err
			}
			tfVar.DefaultValue = valStr

		case "sensitive":
			// Numbers and bools are both parsed as SimpleJSONValue, so we type check
			// through the gocty conversion error handling
			var sensitiveVal bool
			err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &sensitiveVal)
			if err != nil {
				return tfVar, fmt.Errorf("failed to resolve 'sensitive' argument for variable '%s': %w", name, err)
			}

		case "type":

			tfVar.Type = formatVariableTypeString(v.(string))

		}
	}
	return tfVar, nil
}

// Cleanup the value for the variable type
// formatString uses regex to remove "${" and "}" from the input string.
func formatVariableTypeString(input string) string {
	re := regexp.MustCompile(`^\$\{(.+)\}$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractValidationBlock(tfVar string) (string, error) {
	// Define a regex pattern to match the validation blocks
	validationBlockPattern := `validation\s*\{[^}]+\}`

	// Compile the regex pattern
	re, err := regexp.Compile(validationBlockPattern)
	if err != nil {
		return "", err
	}

	// Find all validation blocks in the given string
	validationBlocks := re.FindAllString(tfVar, -1)
	if len(validationBlocks) == 0 {
		return "", fmt.Errorf("no validation blocks found")
	}

	return strings.Join(validationBlocks, "\n\n"), nil
}
