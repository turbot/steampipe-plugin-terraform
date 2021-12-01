package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
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
}

func listLocals(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydate, defaulting to the config paths or
	// available by the optional key column
	path := h.Item.(filePath).Path

	combinedParser, err := Parser()
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.listLocals", "create_parser_error", err)
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.listLocals", "read_file_error", err, "path", path)
		return nil, err
	}

	var tfLocal terraformLocal

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_local.listLocals", "parse_error", err, "path", path)
			return nil, err
		}

		for _, doc := range docs {
			if doc["locals"] != nil {
				// Locals are grouped by local blocks
				switch localType := doc["locals"].(type) {

				// If more than 1 local block is defined, an array of interfaces is returned
				case []interface{}:
					for _, locals := range doc["locals"].([]interface{}) {
						// Get lines map to use when building each local row
						linesMap := locals.(model.Document)["_kics_lines"].(map[string]model.LineObject)
						for localName, localValue := range locals.(model.Document) {
							tfLocal, err = buildLocal(ctx, path, localName, localValue, locals.(model.Document), linesMap)
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
					for localName, localValue := range doc["locals"].(model.Document) {
						tfLocal, err = buildLocal(ctx, path, localName, localValue, doc["locals"].(model.Document), linesMap)
						if err != nil {
							plugin.Logger(ctx).Error("terraform_local.listLocals", "build_local_error", err)
							return nil, err
						}
						d.StreamListItem(ctx, tfLocal)
					}

				default:
					plugin.Logger(ctx).Error("terraform_local.listLocals", "unknown_type", localType)
					return nil, fmt.Errorf("Failed to list locals in %s due to unknown type", path)
				}

			}
		}
	}
	return nil, nil
}

func buildLocal(ctx context.Context, path string, name string, value interface{}, d model.Document, lineMap map[string]model.LineObject) (terraformLocal, error) {
	var tfLocal terraformLocal
	tfLocal.Path = path
	tfLocal.Name = name

	valStr, err := convertExpressionValue(value)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_local.buildLocal", "convert_value_error", err)
		return tfLocal, err
	}
	tfLocal.Value = valStr

	// Each starting line number is stored in "_kics_localName", e.g., "_kics_foo"
	lineKey := "_kics_" + name
	defaultLine := lineMap[lineKey]
	tfLocal.StartLine = defaultLine.Line

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	return tfLocal, nil
}
