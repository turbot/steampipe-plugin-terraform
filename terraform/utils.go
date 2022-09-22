package terraform

import (
	"context"
	_ "embed" // Embed kics CLI img and scan-flags
	json "encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
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
	quals := d.EqualsQuals
	if quals["path"] != nil {
		d.StreamListItem(ctx, filePath{Path: quals["path"].GetStringValue()})
		return nil, nil
	}

	// #2 - paths in config

	// Fail if no paths are specified
	terraformConfig := GetConfig(d.Connection)
	if terraformConfig.Paths == nil {
		return nil, errors.New("paths must be configured")
	}

	// Gather file path matches for the glob
	var matches []string
	paths := terraformConfig.Paths
	for _, i := range paths {

		// List the files in the given source directory
		files, err := d.GetSourceFiles(i)
		if err != nil {
			return nil, err
		}
		matches = append(matches, files...)
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

func getBlock(ctx context.Context, path string, content []byte, blockType string, matchLabels []string) (startPos hcl.Pos, endPos hcl.Pos, source string, _ error) {
	parser := hclparse.NewParser()
	file, _ := parser.ParseHCL(content, path)
	fileContent, _, diags := file.Body.PartialContent(terraformSchema)
	if diags.HasErrors() {
		return hcl.InitialPos, hcl.InitialPos, "", errors.New(diags.Error())
	}
	for _, block := range fileContent.Blocks.OfType(blockType) {
		if isBlockMatch(block, blockType, matchLabels) {
			syntaxBody, ok := block.Body.(*hclsyntax.Body)
			if !ok {
				// this should never happen
				plugin.Logger(ctx).Info("could not cast to hclsyntax")
				break
			}

			startPos = syntaxBody.SrcRange.Start
			endPos = syntaxBody.SrcRange.End
			source = strings.Join(
				strings.Split(
					string(content),
					"\n",
				)[(syntaxBody.SrcRange.Start.Line-1):syntaxBody.SrcRange.End.Line],
				"\n",
			)

			break
		}
	}
	return
}

func isBlockMatch(block *hcl.Block, blockType string, matchLabels []string) bool {
	if !strings.EqualFold(block.Type, blockType) {
		return false
	}

	if len(block.Labels) != len(matchLabels) {
		return false
	}
	for mIdx, matchLabel := range matchLabels {
		if !strings.EqualFold(block.Labels[mIdx], matchLabel) {
			return false
		}
	}
	return true
}

var terraformSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "terraform",
		},
		{
			// This one is not really valid, but we include it here so we
			// can create a specialized error message hinting the user to
			// nest it inside a "terraform" block.
			Type: "required_providers",
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
		{
			Type: "locals",
		},
		{
			Type:       "output",
			LabelNames: []string{"name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
		},
		{
			Type: "moved",
		},
	},
}
