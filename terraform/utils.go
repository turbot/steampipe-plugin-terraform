package terraform

import (
	"context"
	_ "embed" // Embed kics CLI img and scan-flags
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Checkmarx/kics/pkg/parser"
	jsonParser "github.com/Checkmarx/kics/pkg/parser/json"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	yamlParser "github.com/Checkmarx/kics/pkg/parser/yaml"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

type filePath struct {
	Path string
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

func Parser() ([]*parser.Parser, error) {

	combinedParser, err := parser.NewBuilder().
		Add(&jsonParser.Parser{}).
		Add(&yamlParser.Parser{}).
		Add(terraformParser.NewDefault()).
		Build([]string{"Terraform"}, []string{""})
	if err != nil {
		return nil, err
	}

	return combinedParser, nil
}
