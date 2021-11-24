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

func tableTerraformResource(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_resource",
		Description: "Terraform resource information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listResources,
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
				Description: "Resource name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "Resource type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "arguments",
				Description: "Resource arguments.",
				Type:        proto.ColumnType_JSON,
			},
			// Meta-arguments
			{
				Name:        "count",
				Description: "The count meta-argument accepts a whole number, and creates that many instances of the resource or module.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "for_each",
				Description: "The for_each meta-argument accepts a map or a set of strings, and creates an instance for each item in that map or set.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "depends_on",
				Description: "Use the depends_on meta-argument to handle hidden resource or module dependencies that Terraform can't automatically infer.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "lifecycle",
				Description: "The lifecycle meta-argument is a nested block that can appear within a resource block.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "provider",
				Description: "The provider meta-argument specifies which provider configuration to use for a resource, overriding Terraform's default behavior of selecting one based on the resource type name.",
				Type:        proto.ColumnType_STRING,
			},
		},
	}
}

type terraformResource struct {
	Name      string
	Type      string
	Path      string
	StartLine int
	Arguments map[string]interface{}
	DependsOn []string
	Count     int
	ForEach   map[string]interface{}
	// A resource's provider arg will always reference a provider block
	Provider  string
	Lifecycle map[string]interface{}
}

func listResources(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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

	var tfResource terraformResource

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			panic(err)
		}

		for _, doc := range docs {
			if doc["resource"] != nil {
				// Resources are grouped by resource type
				for resourceType, resources := range doc["resource"].(model.Document) {
					plugin.Logger(ctx).Warn("Resource:", resources)
					// For each resource, scan its arguments
					for resourceName, resourceData := range resources.(model.Document) {
						tfResource, err = buildResource(ctx, path, resourceType, resourceName, resourceData.(model.Document))
						if err != nil {
							panic(err)
						}
						d.StreamListItem(ctx, tfResource)
					}
				}
			}
		}
	}

	return nil, nil
}

func buildResource(ctx context.Context, path string, resourceType string, name string, d model.Document) (terraformResource, error) {
	var tfResource terraformResource

	tfResource.Path = path
	tfResource.Type = resourceType
	tfResource.Name = name
	tfResource.Arguments = make(map[string]interface{})
	tfResource.Lifecycle = make(map[string]interface{})

	// The starting line number is stored in "_kics__default"
	kicsLines := d["_kics_lines"]
	linesMap := kicsLines.(map[string]model.LineObject)
	defaultLine := linesMap["_kics__default"]
	tfResource.StartLine = defaultLine.Line

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	// TODO: Can we return source code as well?
	for k, v := range d {
		switch k {
		case "count":
			var countVal int
			err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &countVal)
			if err != nil {
				// TODO: Return error normally instead
				panic(err)
			}
			tfResource.Count = countVal

		case "provider":
			tfResource.Provider = v.(string)

		case "for_each":
			tfResource.ForEach = v.(model.Document)

		case "lifecycle":
			for k, v := range v.(model.Document) {
				if !strings.HasPrefix(k, "_kics") {
					tfResource.Lifecycle[k] = v
				}
			}

		case "depends_on":
			interfaces := v.([]interface{})
			s := make([]string, len(interfaces))
			for i, v := range interfaces {
				s[i] = fmt.Sprint(v)
			}
			tfResource.DependsOn = s

		// It's safe to add any remaining arguments since we've already removed all "_kics" arguments
		default:
			tfResource.Arguments[k] = v
		}
	}
	return tfResource, nil
}
