package terraform

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

func tableTerraformDataSource(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_data source",
		Description: "Terraform data source information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listDataSources,
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
				Description: "Data source name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "Data source type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "arguments",
				Description: "Data source arguments.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "count",
				Description: "The count meta-argument accepts a whole number, and creates that many instances of the data source or module.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "for_each",
				Description: "The for_each meta-argument accepts a map or a set of strings, and creates an instance for each item in that map or set.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "depends_on",
				Description: "Use the depends_on meta-argument to handle hidden data source or module dependencies that Terraform can't automatically infer.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "provider",
				Description: "The provider meta-argument specifies which provider configuration to use for a data source, overriding Terraform's default behavior of selecting one based on the data source type name.",
				Type:        proto.ColumnType_STRING,
			},
		},
	}
}

type terraformDataSource struct {
	Name      string
	Type      string
	Path      string
	StartLine int
	Arguments map[string]interface{}
	DependsOn []string
	// Count can be a number or refer to a local or variable
	Count   string
	ForEach string
	// A data source's provider arg will always reference a provider block
	Provider string
}

func listDataSources(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydate, defaulting to the config paths or
	// available by the optional key column
	path := h.Item.(filePath).Path

	combinedParser, err := Parser()
	if err != nil {
		plugin.Logger(ctx).Error("terraform_data_source.listDataSources", "create_parser_error", err)
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_data_source.listDataSources", "read_file_error", err, "path", path)
		return nil, err
	}

	var tfDataSource terraformDataSource

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_data_source.listDataSources", "parse_error", err, "path", path)
			return nil, err
		}

		for _, doc := range docs {
			if doc["data"] != nil {
				// Data sources are grouped by data source type
				for dataSourceType, dataSources := range doc["data"].(model.Document) {
					tfDataSource.Path = path
					tfDataSource.Type = dataSourceType
					// For each data source, scan its arguments
					for dataSourceName, dataSourceData := range dataSources.(model.Document) {
						tfDataSource, err = buildDataSource(ctx, path, dataSourceType, dataSourceName, dataSourceData.(model.Document))
						if err != nil {
							plugin.Logger(ctx).Error("terraform_data_source.listDataSources", "build_data_source_error", err)
							return nil, err
						}
						d.StreamListItem(ctx, tfDataSource)
					}
				}
			}
		}
	}

	return nil, nil
}

func buildDataSource(ctx context.Context, path string, dataSourceType string, name string, d model.Document) (terraformDataSource, error) {
	var tfDataSource terraformDataSource

	tfDataSource.Path = path
	tfDataSource.Type = dataSourceType
	tfDataSource.Name = name
	tfDataSource.Arguments = make(map[string]interface{})

	// The starting line number is stored in "_kics__default"
	kicsLines := d["_kics_lines"]
	linesMap := kicsLines.(map[string]model.LineObject)
	defaultLine := linesMap["_kics__default"]
	tfDataSource.StartLine = defaultLine.Line

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	for k, v := range d {
		switch k {
		case "count":
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_data_source.buildDataSource", "convert_count_error", err)
				return tfDataSource, err
			}
			tfDataSource.Count = valStr

		case "provider":
			if reflect.TypeOf(v).String() != "string" {
				return tfDataSource, fmt.Errorf("The 'provider' argument for data source '%s' must be of type string", name)
			}
			tfDataSource.Provider = v.(string)

		case "for_each":
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_data_source.buildDataSource", "convert_for_each_error", err)
				return tfDataSource, err
			}
			tfDataSource.ForEach = valStr

		case "depends_on":
			if reflect.TypeOf(v).String() != "[]interface{}" {
				return tfDataSource, fmt.Errorf("The 'depends_on' argument for data source '%s' must be of type list", name)
			}
			interfaces := v.([]interface{})
			s := make([]string, len(interfaces))
			for i, v := range interfaces {
				s[i] = fmt.Sprint(v)
			}
			tfDataSource.DependsOn = s

		// It's safe to add any remaining arguments since we've already removed all "_kics" arguments
		default:
			tfDataSource.Arguments[k] = v
		}
	}
	return tfDataSource, nil
}
