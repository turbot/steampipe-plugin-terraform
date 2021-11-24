package terraform

import (
	"context"
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
				Name:        "path",
				Description: "Path to the file.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Local name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "value",
				Description: "Local value.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
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
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tfLocal terraformLocal

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			panic(err)
		}

		for _, doc := range docs {
			if doc["locals"] != nil {
				plugin.Logger(ctx).Warn("Local top level:", doc["locals"])
				plugin.Logger(ctx).Warn("Local top level model:", doc["locals"].([]interface{}))
				// Locals are grouped by local blocks
				for _, locals := range doc["locals"].([]interface{}) {
					plugin.Logger(ctx).Warn("Locals:", locals)
					// Get lines map to use when building each local row
					linesMap := locals.(model.Document)["_kics_lines"].(map[string]model.LineObject)

					for localName, localValue := range locals.(model.Document) {
						tfLocal, err = buildLocal(ctx, path, localName, localValue, locals.(model.Document), linesMap)
						if err != nil {
							panic(err)
						}
						d.StreamListItem(ctx, tfLocal)
					}
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
	plugin.Logger(ctx).Warn("Val:", value)
	// TODO: Convert value into the correct (?) type
	//tfLocal.Value = value
	tfLocal.Value = "test"

	// Each starting line number is stored in "_kics_localName", e.g., "_kics_foo"
	lineKey := "_kics_" + name
	defaultLine := lineMap[lineKey]
	tfLocal.StartLine = defaultLine.Line

	// Remove all "_kics" properties
	sanitizeDocument(d)

	return tfLocal, nil
}
