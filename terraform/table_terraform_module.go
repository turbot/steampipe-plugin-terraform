package terraform

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

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
				Name:        "module_source",
				Description: "Module source",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "version",
				Description: "Module version",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "variables",
				Description: "Input variables passed to this module.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "count",
				Description: "The integer value for the count meta-argument if it's set as a number in a literal expression.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "count_src",
				Description: "The count meta-argument accepts a whole number, and creates that many instances of the resource or module.",
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

type terraformModule struct {
	Name      string
	Path      string
	StartLine int
	EndLine   int
	Source    string
	Variables map[string]interface{}
	DependsOn []string
	// Count can be a number or refer to a local or variable
	Count    int
	CountSrc string
	ForEach  string
	// A data source's provider arg will always reference a provider block
	Provider     string
	ModuleSource string
	Version      string
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

	var tfModule *terraformModule

	for _, parser := range combinedParser {
		parsedDocs, err := ParseContent(ctx, d, path, content, parser)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_module.listModules", "parse_error", err, "path", path)
			return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
		}

		for _, doc := range parsedDocs.Docs {
			if doc["module"] != nil {
				for moduleName, moduleData := range doc["module"].(model.Document) {
					tfModule, err = buildModule(ctx, path, content, moduleName, moduleData.(model.Document))
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

func buildModule(ctx context.Context, path string, content []byte, name string, d model.Document) (*terraformModule, error) {
	tfModule := new(terraformModule)

	tfModule.Path = path
	tfModule.Name = name
	tfModule.Variables = make(map[string]interface{})

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	startPosition, endPosition, source, err := getBlock(ctx, path, content, "module", []string{name})
	if err != nil {
		plugin.Logger(ctx).Error("error getting details of block", err)
		return nil, err
	}

	tfModule.StartLine = startPosition.Line
	tfModule.Source = source
	tfModule.EndLine = endPosition.Line

	for k, v := range d {
		switch k {
		case "source":
			if reflect.TypeOf(v).String() != "string" {
				return tfModule, fmt.Errorf("The 'source' argument for module '%s' must be of type string", name)
			}
			tfModule.ModuleSource = v.(string)

		case "version":
			if reflect.TypeOf(v).String() != "string" {
				return tfModule, fmt.Errorf("The 'version' argument for module '%s' must be of type string", name)
			}
			tfModule.Version = v.(string)

		case "count":
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_module.buildDataSource", "convert_count_error", err)
				return tfModule, err
			}
			tfModule.CountSrc = valStr

			// Only attempt to get the int value if the type is SimpleJSONValue
			if reflect.TypeOf(v).String() == "json.SimpleJSONValue" {
				var countVal int
				err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &countVal)
				// Log the error but don't return the err since we have count_src anyway
				if err != nil {
					plugin.Logger(ctx).Warn("terraform_module.buildResource", "convert_count_error", err)
				}
				tfModule.Count = countVal
			}

		case "provider":
			if reflect.TypeOf(v).String() != "string" {
				return tfModule, fmt.Errorf("The 'provider' argument for module '%s' must be of type string", name)
			}
			tfModule.Provider = v.(string)

		case "for_each":
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_module.buildDataSource", "convert_for_each_error", err)
				return tfModule, err
			}
			tfModule.ForEach = valStr

		case "depends_on":
			if reflect.TypeOf(v).String() != "[]interface {}" {
				return tfModule, fmt.Errorf("The 'depends_on' argument for module '%s' must be of type list", name)
			}
			interfaces := v.([]interface{})
			s := make([]string, len(interfaces))
			for i, v := range interfaces {
				s[i] = fmt.Sprint(v)
			}
			tfModule.DependsOn = s

		case "lifecycle":
			// ignoring as lifecycle block is reserved for future versions, see
			// https://developer.hashicorp.com/terraform/language/modules/syntax#meta-arguments

		default:
			// safe to add any remaining arguments since already removed all "_kics" arguments
			tfModule.Variables[k] = v

		}
	}
	return tfModule, nil
}
