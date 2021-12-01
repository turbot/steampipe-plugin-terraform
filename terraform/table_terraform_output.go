package terraform

import (
	"context"
	"fmt"
	"os"
	"reflect"

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
			// TODO: Transform to expression
			{
				Name:        "value",
				Description: "The value argument takes an expression whose result is to be returned to the user.",
				Type:        proto.ColumnType_JSON,
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
	//Value       cty.Value `column:"value,jsonb"`
	//Value interface{}
}

func listOutputs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydate, defaulting to the config paths or
	// available by the optional key column
	path := h.Item.(filePath).Path

	combinedParser, err := Parser()
	if err != nil {
		plugin.Logger(ctx).Error("terraform_output.listOutputs", "create_parser_error", err)
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_output.listOutputs", "read_file_error", err, "path", path)
		return nil, err
	}

	var tfOutput terraformOutput

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_output.listOutputs", "parse_error", err, "path", path)
			return nil, err
		}

		for _, doc := range docs {
			if doc["output"] != nil {
				// For each output, scan its arguments
				for outputName, outputData := range doc["output"].(model.Document) {
					plugin.Logger(ctx).Warn("Output:", outputData)
					tfOutput, err = buildOutput(ctx, path, outputName, outputData.(model.Document))
					if err != nil {
						plugin.Logger(ctx).Error("terraform_output.listOutputs", "build_output_error", err)
						return nil, err
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

	// The starting line number is stored in "_kics__default"
	kicsLines := d["_kics_lines"]
	linesMap := kicsLines.(map[string]model.LineObject)
	defaultLine := linesMap["_kics__default"]
	tfOutput.StartLine = defaultLine.Line

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	for k, v := range d {
		switch k {
		case "description":
			tfOutput.Description = v.(string)

		case "value":
			ty := reflect.TypeOf(v)
			plugin.Logger(ctx).Warn("Value Type:", ty)
			plugin.Logger(ctx).Warn("Name:", ty.Name())
			plugin.Logger(ctx).Warn("String:", ty.String())
			plugin.Logger(ctx).Warn("Kind:", ty.Kind())

			valStr, err := convertValue(v)
			if err != nil {
				plugin.Logger(ctx).Warn("String conv error:", err)
				tfOutput.Value = "bad string conversion"
				// TODO: Re-enable error handling once conversions work
				//return tfOutput, fmt.Errorf("Failed to resolve value argument for output%s: %w", name, err)
			}
			tfOutput.Value = valStr

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
