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

func tableTerraformProvider(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "terraform_provider",
		Description: "Terraform provider information.",
		List: &plugin.ListConfig{
			ParentHydrate: tfConfigList,
			Hydrate:       listProviders,
			KeyColumns:    plugin.OptionalColumns([]string{"path"}),
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "Provider name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "arguments",
				Description: "Provider arguments.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "alias",
				Description: "The alias meta-argument to provide an extra name segment.",
				Type:        proto.ColumnType_STRING,
			},
			// Version is deprecated as of Terraform 0.13, but some older files may still use it
			{
				Name:        "version",
				Description: "The version meta-argument specifies a version constraint for a provider, and works the same way as the version argument in a required_providers block.",
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

type terraformProvider struct {
	Name      string
	Path      string
	StartLine int
	EndLine   int
	Source    string
	Arguments map[string]interface{}
	Alias     string
	Version   string
}

func listProviders(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// The path comes from a parent hydate, defaulting to the config paths or
	// available by the optional key column
	data := h.Item.(filePath)
	path := data.Path

	// Return if the path is a TF plan path
	if data.IsTFPlanFilePath {
		return nil, nil
	}

	// If the path was requested through qualifier then match it exactly. Globs
	// are not supported in this context since the output value for the column
	// will never match the requested value.
	if d.EqualsQualString("path") != "" && d.EqualsQualString("path") != path {
		return nil, nil
	}

	combinedParser, err := Parser()
	if err != nil {
		plugin.Logger(ctx).Error("terraform_provider.listProviders", "create_parser_error", err)
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		plugin.Logger(ctx).Error("terraform_provider.listProviders", "read_file_error", err, "path", path)
		return nil, err
	}

	var tfProvider terraformProvider

	for _, parser := range combinedParser {
		parsedDocs, err := ParseContent(ctx, d, path, content, parser)
		if err != nil {
			plugin.Logger(ctx).Error("terraform_provider.listProviders", "parse_error", err, "path", path)
			return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
		}

		for _, doc := range parsedDocs.Docs {
			if doc["provider"] != nil {
				// Providers are grouped by provider name
				for providerName, providers := range doc["provider"].(model.Document) {
					// If more than 1 provider with the same name, an array of interfaces is returned
					switch providerType := providers.(type) {

					case []interface{}:
						for _, providerData := range providers.([]interface{}) {
							// For each provider, scan its arguments
							tfProvider, err = buildProvider(ctx, path, content, providerName, providerData.(model.Document))
							if err != nil {
								plugin.Logger(ctx).Error("terraform_provider.listProviders", "build_provider_error", err)
								return nil, err
							}
							d.StreamListItem(ctx, tfProvider)
						}

						// If only 1 provider has the name, a model.Document is returned
					case model.Document:
						// For each provider, scan its arguments
						tfProvider, err = buildProvider(ctx, path, content, providerName, providers.(model.Document))
						if err != nil {
							plugin.Logger(ctx).Error("terraform_provider.listProviders", "build_provider_error", err)
							return nil, err
						}
						d.StreamListItem(ctx, tfProvider)

					default:
						plugin.Logger(ctx).Error("terraform_provider.listProviders", "unknown_type", providerType)
						return nil, fmt.Errorf("Failed to list providers due to unknown type for provider %s", providerName)
					}

				}
			}
		}
	}

	return nil, nil
}

func buildProvider(ctx context.Context, path string, content []byte, name string, d model.Document) (terraformProvider, error) {
	var tfProvider terraformProvider
	tfProvider.Path = path
	tfProvider.Name = name
	tfProvider.Arguments = make(map[string]interface{})

	// Remove all "_kics" arguments
	sanitizeDocument(d)

	start, end, source, err := getBlock(ctx, path, content, "provider", []string{name})
	if err != nil {
		plugin.Logger(ctx).Error("terraform_provider.buildProvider", "getBlock", err)
		return tfProvider, err
	}
	tfProvider.StartLine = start.Line
	tfProvider.EndLine = end.Line
	tfProvider.Source = source

	for k, v := range d {
		switch k {
		case "alias":
			if reflect.TypeOf(v).String() != "string" {
				return tfProvider, fmt.Errorf("The 'alias' argument for provider '%s' must be of type string", name)
			}
			tfProvider.Alias = v.(string)

		case "version":
			if reflect.TypeOf(v).String() != "string" {
				return tfProvider, fmt.Errorf("The 'version' argument for provider '%s' must be of type string", name)
			}
			tfProvider.Version = v.(string)

		// It's safe to add any remaining arguments since we've already removed all "_kics" arguments
		default:
			tfProvider.Arguments[k] = v
		}
	}

	return tfProvider, nil
}
