package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableTerraformLocal(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_local",
		Description: "Terraform local information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listLocals,
			KeyColumns:    plugin.OptionalColumns([]string{"path"}),
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "Local name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "value",
				Description: "Local value.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
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

type terraformLocal struct {
	Name      string
	Value     string
	Path      string
	StartLine int
	EndLine   int
	Source    string
}

func listLocals(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydate, defaulting to the config paths or
	// available by the optional key column
	data := h.Item.(filePath)
	path := data.Path

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.listLocals", "read_file_error", err, "path", path)
		return nil, err
	}

	// Return if the path is a TF plan path
	if data.IsTFPlanFilePath && !isTerraformPlan(content) {
		return nil, nil
	}

	combinedParser, err := Parser()
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.listLocals", "create_parser_error", err)
		return nil, err
	}

	for _, parser := range combinedParser {
		parsedDocs, err := ParseContent(ctx, d, path, content, parser)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_local.listLocals", "parse_error", err, "path", path)
			return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
		}

		for _, doc := range parsedDocs.Docs {
			if doc["locals"] != nil {
				// Locals are grouped by local blocks
				switch localType := doc["locals"].(type) {

				// If more than 1 local block is defined, an array of interfaces is returned
				case []interface{}:
					for _, locals := range doc["locals"].([]interface{}) {
						// Get lines map to use when building each local row
						linesMap := locals.(model.Document)["_kics_lines"].(map[string]model.LineObject)
						// Remove all "_kics" arguments now that we have the lines map
						sanitizeDocument(locals.(model.Document))
						for localName, localValue := range locals.(model.Document) {
							tfLocal, err := buildLocal(ctx, path, content, localName, localValue, linesMap)
							if err != nil {
								plugin.Logger(ctx).Error("terraform_local.listLocals", "build_local_error", err)
								return nil, err
							}
							d.StreamListItem(ctx, tfLocal)
						}
					}

				// If only 1 local block is defined, a model.Document is returned
				case model.Document:
					// Get lines map to use when building each local row
					linesMap := doc["locals"].(model.Document)["_kics_lines"].(map[string]model.LineObject)
					// Remove all "_kics" arguments now that we have the lines map
					sanitizeDocument(doc["locals"].(model.Document))
					for localName, localValue := range doc["locals"].(model.Document) {
						tfLocal, err := buildLocal(ctx, path, content, localName, localValue, linesMap)
						if err != nil {
							plugin.Logger(ctx).Error("terraform_local.listLocals", "build_local_error", err)
							return nil, err
						}
						d.StreamListItem(ctx, tfLocal)
					}

				default:
					plugin.Logger(ctx).Error("terraform_local.listLocals", "unknown_type", localType)
					return nil, fmt.Errorf("failed to list locals in %s due to unknown type", path)
				}

			}
		}
	}
	return nil, nil
}

func buildLocal(ctx context.Context, path string, content []byte, name string, value interface{}, lineMap map[string]model.LineObject) (*terraformLocal, error) {
	tfLocal := new(terraformLocal)
	tfLocal.Path = path
	tfLocal.Name = name

	valStr, err := convertExpressionValue(value)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.buildLocal", "convert_value_error", err)
		return nil, err
	}
	tfLocal.Value = valStr

	start, end, source, err := getBlock(ctx, path, content, "locals", []string{})
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.buildLocal", "getBlock", err)
		return nil, err
	}
	tfLocal.StartLine = start.Line
	tfLocal.EndLine = end.Line
	tfLocal.Source = source

	return tfLocal, nil
}
