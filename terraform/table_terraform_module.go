package terraform

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableTerraformModule(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_module",
		Description: "Terraform module information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listModules,
			KeyColumns:    plugin.OptionalColumns([]string{"path"}),
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "Module name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "source",
				Description: "Module source",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "version",
				Description: "Module version",
				Type:        proto.ColumnType_STRING,
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

type terraformModule struct {
	Name      string
	Path      string
	StartLine int
	Source    string
	Version   string
}

func listModules(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydrate, defaulting to the config paths
	// or available by the optional key column
	path := h.Item.(filePath).Path

	combinedParser, err := Parser()
	if err != nil {
		plugin.Logger(ctx).Error("terraform_module.listModules", "create_parser_error", err)
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_module.listModules", "read_file_error", err, "path", path)
		return nil, err
	}

	var tfModule terraformModule

	for _, parser := range combinedParser {
		parsedDocs, err := ParseContent(ctx, d, path, content, parser)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_module.listModules", "parse_error", err, "path", path)
			return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
		}

		for _, doc := range parsedDocs.Docs {
			if doc["module"] != nil {
				for moduleName, moduleData := range doc["module"].(model.Document) {
					tfModule, err = buildModule(ctx, path, moduleName, moduleData.(model.Document))
					if err != nil {
						plugin.Logger(ctx).Error("terraform_module.listModules", "build_module_error", err)
						return nil, err
					}
					d.StreamListItem(ctx, tfModule)
				}
			}
		}
	}

	return nil, nil
}

func buildModule(_ context.Context, path string, name string, d model.Document) (terraformModule, error) {
	var tfModule terraformModule

	tfModule.Path = path
	tfModule.Name = name

	// The starting line number is stored in "_kics__default"
	kicsLines := d["_kics_lines"]
	linesMap := kicsLines.(map[string]model.LineObject)
	defaultLine := linesMap["_kics__default"]
	tfModule.StartLine = defaultLine.Line

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	for k, v := range d {
		switch k {
		case "source":
			if reflect.TypeOf(v).String() != "string" {
				return tfModule, fmt.Errorf("The 'source' argument for module '%s' must be of type string", name)
			}
			tfModule.Source = v.(string)

		case "version":
			if reflect.TypeOf(v).String() != "string" {
				return tfModule, fmt.Errorf("The 'version' argument for module '%s' must be of type string", name)
			}
			tfModule.Version = v.(string)
		}
	}
	return tfModule, nil
}
