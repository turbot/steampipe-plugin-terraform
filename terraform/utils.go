package terraform

import (
	"context"
	_ "embed" // Embed kics CLI img and scan-flags
	json "encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	"github.com/bmatcuk/doublestar"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type filePath struct {
	Path string
}

// Use when parsing any TF file to prevent concurrent map read and write errors
var parseMutex = sync.Mutex{}

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
	if terraformConfig.Paths == nil {
		return nil, errors.New("paths must be configured")
	}

	// Gather file path matches for the glob
	var matches []string
	paths := terraformConfig.Paths
	for _, i := range paths {
		// Check to resolve ~ to home dir
		if strings.HasPrefix(i, "~") {
			// File system context
			home, err := os.UserHomeDir()
			if err != nil {
				plugin.Logger(ctx).Error("utils.tfConfigList", "os.UserHomeDir error. ~ will not be expanded in paths.", err)
			}

			// Resolve ~ to home dir
			if home != "" {
				if i == "~" {
					i = home
				} else if strings.HasPrefix(i, "~/") {
					i = filepath.Join(home, i[2:])
				}
			}
		}

		// Get full path
		fullPath, err := filepath.Abs(i)
		if err != nil {
			plugin.Logger(ctx).Error("utils.tfConfigList", "failed to fetch absolute path", err, "path", i)
			return nil, err
		}

		iMatches, err := doublestar.Glob(fullPath)
		if err != nil {
			// Fail if any path is an invalid glob
			return nil, fmt.Errorf("Path is not a valid glob: %s", i)
		}
		matches = append(matches, iMatches...)
	}

	// Sanitize the matches to likely Terraform files
	for _, i := range matches {
		// Check if file or directory
		fileInfo, err := os.Stat(i)
		if err != nil {
			plugin.Logger(ctx).Error("utils.tfConfigList", "error getting file info", err, "path", i)
			return nil, err
		}

		// Ignore directories
		if fileInfo.IsDir() {
			continue
		}

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
		d.StreamListItem(ctx, filePath{Path: i})
	}

	return nil, nil
}

func Parser() ([]*parser.Parser, error) {

	combinedParser, err := parser.NewBuilder().
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

// For any arguments that can be a TF expression, convert to string for easier handling
func convertExpressionValue(v interface{}) (valStr string, err error) {
	switch v := v.(type) {
	// Numbers and bools
	case ctyjson.SimpleJSONValue:
		val, err := v.MarshalJSON()
		if err != nil {
			return "", fmt.Errorf("Failed to convert SimpleJSONValue value %v: %w", v, err)
		}
		valStr = string(val)

	case string:
		val, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("Failed to convert string value %v: %w", v, err)
		}
		valStr = string(val)

	// Maps
	case model.Document:
		val, err := v.MarshalJSON()
		if err != nil {
			return "", fmt.Errorf("Failed to convert model.Document value %v: %w", v, err)
		}
		valStr = string(val)

	// Arrays
	case []interface{}:
		var valStrs []string
		for _, iValue := range v {
			tempVal, err := convertExpressionValue(iValue)
			if err != nil {
				return "", fmt.Errorf("Failed to convert []interface{} value %v: %w", v, err)
			}
			valStrs = append(valStrs, tempVal)
		}
		valStr = fmt.Sprintf("[%s]", strings.Join(valStrs, ","))

	default:
		return "", fmt.Errorf("Failed to convert value %v due to unknown type: %T", v, v)
	}
	return valStr, nil
}

func ParseContent(ctx context.Context, d *plugin.QueryData, path string, content []byte, p *parser.Parser) (parser.ParsedDocument, error) {
	// Only allow parsing of one file at a time to prevent concurrent map read
	// and write errors
	parseMutex.Lock()
	defer parseMutex.Unlock()

	parsedDocs, err := p.Parse(path, content)
	if err != nil {
		plugin.Logger(ctx).Error("utils.ParseContent", "parse_error", err, "path", path)
		return parser.ParsedDocument{}, err
	}

	return parsedDocs, nil
}
