package terraform

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
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
				Description: "DataSource name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "DataSource type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "properties",
				Description: "DataSource properties.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "count",
				Description: "The count meta-argument accepts a whole number, and creates that many instances of the data source or module.",
				Type:        proto.ColumnType_INT,
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
	Name       string
	Type       string
	Path       string
	StartLine  int
	Properties map[string]interface{}
	DependsOn  []string
	Count      int
	ForEach    map[string]interface{}
	// A data source's provider arg will always reference a provider block
	Provider string
}

func listDataSources(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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

	var tfDataSource terraformDataSource

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			panic(err)
		}

		for _, doc := range docs {
			if doc["data"] != nil {
				// Data sources are grouped by data source type
				for dataSourceType, dataSources := range doc["data"].(model.Document) {
					plugin.Logger(ctx).Warn("Data source:", dataSources)
					tfDataSource.Path = path
					tfDataSource.Type = dataSourceType
					// For each dataSource, scan its properties
					for dataSourceName, dataSourceData := range dataSources.(model.Document) {
						tfDataSource, err = buildDataSource(ctx, path, dataSourceType, dataSourceName, dataSourceData.(model.Document))
						if err != nil {
							panic(err)
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
	tfDataSource.Properties = make(map[string]interface{})

	for k, v := range d {
		switch k {
		// The starting line number is stored in "_kics__default"
		case "_kics_lines":
			linesMap := v.(map[string]model.LineObject)
			defaultLine := linesMap["_kics__default"]
			tfDataSource.StartLine = defaultLine.Line

		case "count":
			var countVal int
			err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &countVal)
			if err != nil {
				// TODO: Return error normally instead
				panic(err)
			}
			tfDataSource.Count = countVal

		case "provider":
			tfDataSource.Provider = v.(string)

		case "for_each":
			tfDataSource.ForEach = v.(model.Document)

		case "depends_on":
			interfaces := v.([]interface{})
			s := make([]string, len(interfaces))
			for i, v := range interfaces {
				s[i] = fmt.Sprint(v)
			}
			tfDataSource.DependsOn = s

		// Avoid adding _kicks properties and meta-arguments directly
		// TODO: Handle map type properties to avoid including _kics properties
		default:
			if !strings.HasPrefix(k, "_kics") {
				tfDataSource.Properties[k] = v
			}
		}
	}
	return tfDataSource, nil
}
