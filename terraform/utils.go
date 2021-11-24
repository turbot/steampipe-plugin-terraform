package terraform

import (
	"context"
	_ "embed" // Embed kics CLI img and scan-flags
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser"
	jsonParser "github.com/Checkmarx/kics/pkg/parser/json"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	yamlParser "github.com/Checkmarx/kics/pkg/parser/yaml"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
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

// Remove all "_kics" arguments to avoid noisy data
func sanitizeDocument(d model.Document) {
	// Deep sanitize
	for k, v := range d {
		if strings.HasPrefix(k, "_kics") {
			delete(d, k)
		}

		if reflect.TypeOf(v).String() == "model.Document" {
			sanitizeDocument(v.(model.Document))
		}

		// Some map arguments are returned as "[]interface {}" types from the parser
		if reflect.TypeOf(v).String() == "[]interface {}" {
			for _, v := range v.([]interface{}) {
				if reflect.TypeOf(v).String() == "model.Document" {
					sanitizeDocument(v.(model.Document))
				}
			}
		}
	}
}

// CtyToJSON converts a cty value to its JSON representation
func CtyToJSON(val cty.Value) (string, error) {

	if !val.IsWhollyKnown() {
		return "", fmt.Errorf("cannot serialize unknown values")
	}

	if val.IsNull() {
		return "{}", nil
	}

	buf, err := json.Marshal(val, val.Type())
	if err != nil {
		return "", err
	}

	return string(buf), nil

}

// ctyToPostgresString convert a cty value into a postgres representation of the value
func ctyToPostgresString(v cty.Value) (valStr string, err error) {
	ty := v.Type()
	switch {
	case ty.IsTupleType(), ty.IsListType():
		{

			var array []string
			if array, err = ctyTupleToArrayOfPgStrings(v); err == nil {
				valStr = fmt.Sprintf("array[%s]", strings.Join(array, ","))
			}
			return
		}
	}

	switch ty {
	case cty.Bool:
		var target bool
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%v", target)
		}
	case cty.Number:
		var target int
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%d", target)
			return
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				valStr = fmt.Sprintf("%d", target)
			}
		}
	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("'%s'", target)
		}

	default:
		var json string
		// wrap as postgres string
		if json, err = CtyToJSON(v); err == nil {
			valStr = fmt.Sprintf("'%s'::jsonb", json)
		}

	}

	return valStr, err
}

func ctyTupleToArrayOfPgStrings(val cty.Value) ([]string, error) {
	var res []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		// decode the value into a postgres compatible
		valStr, err := ctyToPostgresString(v)
		if err != nil {
			return nil, err
		}

		res = append(res, valStr)
	}
	return res, nil
}

func ctyObjectToMapOfPgStrings(val cty.Value) (map[string]string, error) {
	res := make(map[string]string)
	it := val.ElementIterator()
	for it.Next() {
		k, v := it.Element()

		// decode key
		var key string
		gocty.FromCtyValue(k, &key)

		// decode the value into a postgres compatible
		valStr, err := ctyToPostgresString(v)
		if err != nil {
			err := fmt.Errorf("invalid value provided for param '%s': %v", key, err)
			return nil, err
		}

		res[key] = valStr
	}
	return res, nil
}
