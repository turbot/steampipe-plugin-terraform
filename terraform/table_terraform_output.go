package terraform

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func tableTerraformOutput(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_output",
		Description: "Terraform output information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listOutputs,
			KeyColumns:    plugin.OptionalColumns([]string{"path"}),
		},
		Columns: []*plugin.Column{
			{
				Name:        "path",
				Description: "Path to the file.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Output name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "value",
				Description: "The value argument takes an expression whose result is to be returned to the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Because the output values of a module are part of its user interface, you can briefly describe the purpose of each value using the optional description argument.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "sensitive",
				Description: "An output can be marked as containing sensitive material using the optional sensitive argument.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "depends_on",
				Description: "Use the depends_on meta-argument to handle hidden output or module dependencies that Terraform can't automatically infer.",
				Type:        proto.ColumnType_JSON,
			},
		},
	}
}

type terraformOutput struct {
	Name        string
	Path        string
	StartLine   int
	DependsOn   []string
	Description string
	Sensitive   bool
	Value       string
}

func listOutputs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydate, defaulting to the config paths or
	// available by the optional key column
	path := h.Item.(filePath).Path

	combinedParser, err := Parser()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tfOutput terraformOutput

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			panic(err)
		}

		for _, doc := range docs {
			if doc["output"] != nil {
				// For each output, scan its properties
				for outputName, outputData := range doc["output"].(model.Document) {
					plugin.Logger(ctx).Warn("Output:", outputData)
					tfOutput, err = buildOutput(ctx, path, outputName, outputData.(model.Document))
					if err != nil {
						panic(err)
					}
					d.StreamListItem(ctx, tfOutput)
				}
			}
		}
	}

	return nil, nil
}

func buildOutput(ctx context.Context, path string, name string, d model.Document) (terraformOutput, error) {
	var tfOutput terraformOutput

	tfOutput.Path = path
	tfOutput.Name = name

	for k, v := range d {
		switch k {
		// The starting line number is stored in "_kics__default"
		case "_kics_lines":
			linesMap := v.(map[string]model.LineObject)
			defaultLine := linesMap["_kics__default"]
			tfOutput.StartLine = defaultLine.Line

		case "description":
			tfOutput.Description = v.(string)

		case "value":
			switch v.(type) {
			// TODO: Can we always assume if SimpleJSONValue it's an int?
			case ctyjson.SimpleJSONValue:
				var val int
				err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &val)
				if err != nil {
					// TODO: Return error normally instead
					panic(err)
				}
				tfOutput.Value = strconv.Itoa(val)
				break

			// If not SimpleJSONValue, assume string
			default:
				tfOutput.Value = v.(string)
			}

		case "sensitive":
			var sensitiveVal bool
			err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &sensitiveVal)
			if err != nil {
				// TODO: Return error normally instead
				panic(err)
			}
			tfOutput.Sensitive = sensitiveVal

		case "depends_on":
			interfaces := v.([]interface{})
			s := make([]string, len(interfaces))
			for i, v := range interfaces {
				s[i] = fmt.Sprint(v)
			}
			tfOutput.DependsOn = s
		}
	}
	return tfOutput, nil
}
