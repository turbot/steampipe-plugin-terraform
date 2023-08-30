package terraform

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	p "github.com/Checkmarx/kics/pkg/parser/json"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
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
				Name:        "mode",
				Description: "The type of resource Terraform creates, either a resource (managed) or data source (data).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "arguments",
				Description: "Resource arguments.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "instances",
				Description: "The attributes of the resource.",
				Type:        proto.ColumnType_JSON,
			},
			// Meta-arguments
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

type terraformResource struct {
	Name      string
	Type      string
	Path      string
	Mode      string
	StartLine int
	Source    string
	EndLine   int
	Arguments map[string]interface{}
	DependsOn []string
	// Count can be a number or refer to a local or variable
	Count    int
	CountSrc string
	ForEach  string
	// A resource's provider arg will always reference a provider block
	Provider  string
	Lifecycle map[string]interface{}
	Instances map[string]interface{}
}

func listResources(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydrate, defaulting to the config paths or
	// available by the optional key column
	pathInfo := h.Item.(filePath)
	path := pathInfo.Path

	// Read the content from the file
	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_resource.listResources", "read_file_error", err, "path", path)
		return nil, err
	}

	// if the file contains TF plan then set IsTFPlanFilePath to true
	if isTerraformPlan(content) {
		pathInfo.IsTFPlanFilePath = true
	}

	var docs []model.Document

	// Check if the file contains TF plan or state
	if pathInfo.IsTFPlanFilePath || pathInfo.IsTFStateFilePath {
		// Initialize the JSON parser
		jsonParser := p.Parser{}

		// Parse the file content using the JSON parser
		var str string
		documents, _, err := jsonParser.Parse(str, content)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_resource.listResources", "parse_error", err, "path", path)
			return nil, fmt.Errorf("failed to parse plan or state file %s: %v", path, err)
		}
		docs = append(docs, documents...)
	} else {
		// Build the terraform parser
		combinedParser, err := Parser()
		if err != nil {
			plugin.Logger(ctx).Error("terraform_resource.listResources", "create_parser_error", err)
			return nil, err
		}

		for _, parser := range combinedParser {
			parsedDocs, err := ParseContent(ctx, d, path, content, parser)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_resource.listResources", "parse_error", err, "path", path)
				return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
			}
			docs = append(docs, parsedDocs.Docs...)
		}
	}

	// Stream the data
	for _, doc := range docs {
		if doc["resource"] != nil {
			// Resources are grouped by resource type
			for resourceType, resources := range convertModelDocumentToMapInterface(doc["resource"]) {
				// For each resource, scan its arguments
				for resourceName, resourceData := range convertModelDocumentToMapInterface(resources) {
					tfResource, err := buildResource(ctx, pathInfo.IsTFPlanFilePath, content, path, resourceType, resourceName, convertModelDocumentToMapInterface(resourceData))
					if err != nil {
						plugin.Logger(ctx).Error("terraform_resource.listResources", "build_resource_error", err)
						return nil, err
					}
					d.StreamListItem(ctx, tfResource)
				}
			}
		} else if doc["resources"] != nil { // state file returns resources
			for _, resource := range doc["resources"].([]interface{}) {
				resourceData := convertModelDocumentToMapInterface(resource)
				tfResource, err := buildResource(ctx, pathInfo.IsTFStateFilePath, content, path, resourceData["type"].(string), resourceData["name"].(string), resourceData)
				if err != nil {
					plugin.Logger(ctx).Error("terraform_resource.listResources", "build_resource_error", err)
					return nil, err
				}
				d.StreamListItem(ctx, tfResource)
			}
		}
	}

	return nil, nil
}

func buildResource(ctx context.Context, isTFFilePath bool, content []byte, path string, resourceType string, name string, d model.Document) (*terraformResource, error) {
	tfResource := new(terraformResource)

	tfResource.Path = path
	tfResource.Type = resourceType
	tfResource.Name = name
	tfResource.Arguments = make(map[string]interface{})
	tfResource.Lifecycle = make(map[string]interface{})
	tfResource.Instances = make(map[string]interface{})

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	if isTFFilePath {
		file, err := os.Open(path)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_resource.buildResource", "open_file_error", err, "path", path)
			return tfResource, err
		}
		startLine, endLine, source := findBlockLinesFromJSON(file, "resources", resourceType, name)
		tfResource.StartLine = startLine
		tfResource.EndLine = endLine
		tfResource.Source = source
	} else {
		startPosition, endPosition, source, err := getBlock(ctx, path, content, "resource", []string{resourceType, name})
		if err != nil {
			plugin.Logger(ctx).Error("error getting details of block", err)
			return nil, err
		}

		tfResource.StartLine = startPosition.Line
		tfResource.Source = source
		tfResource.EndLine = endPosition.Line
	}
	// TODO: Can we return source code as well?
	for k, v := range d {
		switch k {
		case "count":
			// The count_src column can handle numbers or strings (expressions)
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_resource.buildResource", "convert_count_error", err)
				return tfResource, err
			}
			tfResource.CountSrc = valStr

			// Only attempt to get the int value if the type is SimpleJSONValue
			if reflect.TypeOf(v).String() == "json.SimpleJSONValue" {
				var countVal int
				err := gocty.FromCtyValue(v.(ctyjson.SimpleJSONValue).Value, &countVal)
				// Log the error but don't return the err since we have count_src anyway
				if err != nil {
					plugin.Logger(ctx).Warn("terraform_resource.buildResource", "convert_count_error", err)
				}
				tfResource.Count = countVal
			}

		case "provider":
			if reflect.TypeOf(v).String() != "string" {
				return tfResource, fmt.Errorf("The 'provider' argument for resource '%s' must be of type string", name)
			}
			tfResource.Provider = v.(string)

		case "name":
			if reflect.TypeOf(v).String() != "string" {
				return tfResource, fmt.Errorf("The 'name' argument for resource '%s' must be of type string", name)
			}
			if tfResource.Name == "" {
				tfResource.Name = v.(string)
			}

		case "type":
			if reflect.TypeOf(v).String() != "string" {
				return tfResource, fmt.Errorf("The 'type' argument for resource '%s' must be of type string", name)
			}
			if tfResource.Name == "" {
				tfResource.Type = v.(string)
			}

		case "mode":
			if reflect.TypeOf(v).String() != "string" {
				return tfResource, fmt.Errorf("The 'mode' argument for resource '%s' must be of type string", name)
			}
			tfResource.Mode = v.(string)

		case "for_each":
			valStr, err := convertExpressionValue(v)
			if err != nil {
				plugin.Logger(ctx).Error("terraform_resource.buildResource", "convert_for_each_error", err)
				return tfResource, err
			}
			tfResource.ForEach = valStr

		case "lifecycle":
			if reflect.TypeOf(v).String() != "model.Document" {
				return tfResource, fmt.Errorf("The 'lifecycle' argument for resource '%s' must be of type map", name)
			}
			for k, v := range v.(model.Document) {
				if !strings.HasPrefix(k, "_kics") {
					tfResource.Lifecycle[k] = v
				}
			}

		case "depends_on":
			if reflect.TypeOf(v).String() != "[]interface {}" {
				return tfResource, fmt.Errorf("The 'depends_on' argument for resource '%s' must be of type list", name)
			}
			interfaces := v.([]interface{})
			s := make([]string, len(interfaces))
			for i, v := range interfaces {
				s[i] = fmt.Sprint(v)
			}
			tfResource.DependsOn = s

		case "instances":
			if reflect.TypeOf(v).String() != "[]interface {}" {
				return tfResource, fmt.Errorf("The 'instances' argument for resource '%s' must be of type list", name)
			}
			for _, v := range v.([]interface{}) {
				convertedValue := convertModelDocumentToMapInterface(v)
				cleanedValue := removeKicsLabels(convertedValue).(map[string]interface{})
				for property, value := range cleanedValue {
					tfResource.Instances[property] = value
				}
			}

		// It's safe to add any remaining arguments since we've already removed all "_kics" arguments
		default:
			tfResource.Arguments[k] = v
		}
	}
	return tfResource, nil
}

// convertModelDocumentToMapInterface takes the documents in model.Document format and converts it into map[string]interface{}
func convertModelDocumentToMapInterface(data interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	switch item := data.(type) {
	case model.Document:
		result = item
	case map[string]interface{}:
		result = item
	}
	return result
}

func removeKicsLabels(data interface{}) interface{} {
	if dataMap, isMap := data.(map[string]interface{}); isMap {
		for key, value := range dataMap {
			if strings.HasPrefix(key, "_kics") {
				delete(dataMap, key)
			} else {
				dataMap[key] = removeKicsLabels(value)
			}
		}
		return dataMap
	} else if dataList, isList := data.([]interface{}); isList {
		for i, item := range dataList {
			dataList[i] = removeKicsLabels(item)
		}
		return dataList
	}
	return data
}
