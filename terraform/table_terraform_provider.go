package terraform

import (
	"context"
	"os"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
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
				Name:        "path",
				Description: "Path to the file.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Provider name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_line",
				Description: "Starting line number.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "properties",
				Description: "Provider properties.",
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
		},
	}
}

type terraformProvider struct {
	Name       string
	Path       string
	StartLine  int
	Properties map[string]interface{}
	Alias      string
	Version    string
}

func listProviders(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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

	var tfProvider terraformProvider

	for _, parser := range combinedParser {
		docs, _, err := parser.Parse(path, content)
		if err != nil {
			panic(err)
		}

		for _, doc := range docs {
			if doc["provider"] != nil {
				//plugin.Logger(ctx).Warn("Provider top level:", doc["provider"])
				//plugin.Logger(ctx).Warn("Provider top level model:", doc["provider"].(model.Document))

				// Providers are grouped by provider name
				for providerName, providers := range doc["provider"].(model.Document) {
					//plugin.Logger(ctx).Warn("Provider name:", providerName)
					//plugin.Logger(ctx).Warn("Providers:", providers)

					// If more than 1 provider with the same name, an array of interfaces is returned
					switch providerType := providers.(type) {
					case []interface{}:
						for _, providerData := range providers.([]interface{}) {
							// For each provider, scan its properties
							tfProvider, err = buildProvider(path, providerName, providerData.(model.Document))
							if err != nil {
								panic(err)
							}
							d.StreamListItem(ctx, tfProvider)
						}
						break

						// If only 1 provider has the name, a model.Document is returned
					case model.Document:
						// For each provider, scan its properties
						//tfProvider, err = buildProvider(path, providerName, providers.(model.Document))
						tfProvider, err = buildProvider(path, providerName, providers.(model.Document))
						if err != nil {
							panic(err)
						}
						d.StreamListItem(ctx, tfProvider)
						break

					default:
						plugin.Logger(ctx).Warn("Type:", providerType)
						panic("Unexpected type")
					}

				}
			}
		}
	}

	return nil, nil
}

func buildProvider(path string, name string, d model.Document) (terraformProvider, error) {
	var tfProvider terraformProvider
	tfProvider.Path = path
	tfProvider.Name = name
	tfProvider.Properties = make(map[string]interface{})

	for k, v := range d {
		// The starting line number for a provider is stored in "_kics__default"
		if k == "_kics_lines" {
			// TODO: Fix line number check
			//tfProvider.StartLine = v.(map[string]interface{})["_kics__default"].(map[string]model.LineObject)["_kics_line"]
			tfProvider.StartLine = 999
		}

		if k == "alias" {
			tfProvider.Alias = v.(string)
		}

		if k == "version" {
			tfProvider.Version = v.(string)
		}

		// Avoid adding _kicks properties and meta-arguments directly
		// TODO: Handle map type properties to avoid including _kics properties
		if !strings.HasPrefix(k, "_kics") && k != "alias" && k != "version" {
			tfProvider.Properties[k] = v
		}
	}

	return tfProvider, nil
}
