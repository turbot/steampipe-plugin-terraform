package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
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
				Transform:   transform.FromField("Content.Name"),
			},
			{
				Name:        "resource_type",
				Description: "Resource type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Content.Type"),
			},
			{
				Name:        "properties",
				Description: "Resource properties.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Content.Properties"),
			},
		},
	}
}

type filePath struct {
	Path string
}

type terraformResourceItem struct {
	Path    string
	Content terraformResource
}

func tfConfigList(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	// #1 - Path via qual

	// If the path was requested through qualifier then match it exactly. Globs
	// are not supported in this context since the output value for the column
	// will never match the requested value.
	quals := d.KeyColumnQuals
	if quals["path"] != nil {
		d.StreamListItem(ctx, filePath{Path: quals["path"].GetStringValue()})
		return nil, nil
	}

	// #2 - Glob paths in config

	// Fail if no paths are specified
	terraformConfig := GetConfig(d.Connection)
	if &terraformConfig == nil || terraformConfig.Paths == nil {
		return nil, errors.New("paths must be configured")
	}

	// Gather file path matches for the glob
	var matches []string
	paths := terraformConfig.Paths
	for _, i := range paths {
		iMatches, err := filepath.Glob(i)
		if err != nil {
			// Fail if any path is an invalid glob
			return nil, errors.New(fmt.Sprintf("Path is not a valid glob: %s", i))
		}
		matches = append(matches, iMatches...)
	}

	// Sanitize the matches to likely Terraform files
	for _, i := range matches {

		// If the file path is an exact match to a matrix path then it's always
		// treated as a match - it was requested exactly
		hit := false
		for _, j := range paths {
			if i == j {
				hit = true
				break
			}
		}
		if hit {
			d.StreamListItem(ctx, filePath{Path: i})
			continue
		}

		// This file was expanded from the glob, so check it's likely to be
		// of the right type based on the extension.
		ext := filepath.Ext(i)
		if ext == ".tf" || ext == ".tf.json" {
			d.StreamListItem(ctx, filePath{Path: i})
		}
	}

	return nil, nil
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
					tfResource.Type = resourceType
					// For each resource, scan its properties
					for resourceName, resourceData := range resources.(model.Document) {
						tfResource.Name = resourceName
						tfResource.Properties = make(map[string]interface{})
						for k, v := range resourceData.(model.Document) {
							// Avoid adding properties like _kics_lines for now
							if !strings.HasPrefix(k, "_kics") {
								tfResource.Properties[k] = v
							}
						}
					}

					// Stream each resource
					item := terraformResourceItem{
						Path:    path,
						Content: tfResource,
					}
					d.StreamListItem(ctx, item)
				}
			}
		}
	}

	return nil, nil
}
