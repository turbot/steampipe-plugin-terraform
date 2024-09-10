package terraform

import (
	"bufio"
	"context"
	_ "embed" // Embed kics CLI img and scan-flags
	json "encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	filehelpers "github.com/turbot/go-kit/files"
)

type filePath struct {
	Path              string
	IsTFPlanFilePath  bool
	IsTFStateFilePath bool
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

		path := d.EqualsQualString("path")

		// check if state file is provide in the qual
		if strings.HasSuffix(path, ".tfstate") {
			d.StreamListItem(ctx, filePath{Path: path, IsTFStateFilePath: true})
			return nil, nil
		}

		d.StreamListItem(ctx, filePath{Path: path})
		return nil, nil
	}

	// #2 - paths in config

	// Fail if no paths are specified
	terraformConfig := GetConfig(d.Connection)
	if terraformConfig.Paths == nil && terraformConfig.ConfigurationFilePaths == nil && terraformConfig.PlanFilePaths == nil && terraformConfig.StateFilePaths == nil {
		return nil, nil
	}

	// Gather file path matches for the glob
	var paths, matches []string

	// TODO:: Remove backward compatibility for the argument 'Paths'
	if terraformConfig.Paths != nil {
		paths = terraformConfig.Paths
	} else {
		paths = terraformConfig.ConfigurationFilePaths
	}
	configurationFilePaths := paths

	for _, i := range configurationFilePaths {

		// List the files in the given source directory
		files, err := d.GetSourceFiles(i)
		if err != nil {
			plugin.Logger(ctx).Error("tfConfigList.configurationFilePaths", "get_source_files_error", err)

			// If the specified path is unavailable, then an empty row should populate
			if strings.Contains(err.Error(), "failed to get directory specified by the source") {
				continue
			}
			return nil, err
		}
		matches = append(matches, files...)
	}

	// Sanitize the matches to ignore the directories
	for _, i := range matches {

		// Ignore directories
		if filehelpers.DirectoryExists(i) {
			continue
		}
		d.StreamListItem(ctx, filePath{Path: i})
	}

	// Gather TF plan file path matches for the glob
	var matchedPlanFilePaths []string
	planFilePaths := terraformConfig.PlanFilePaths
	for _, i := range planFilePaths {

		// List the files in the given source directory
		files, err := d.GetSourceFiles(i)
		if err != nil {
			plugin.Logger(ctx).Error("tfConfigList.planFilePaths", "get_source_files_error", err)

			// If the specified path is unavailable, then an empty row should populate
			if strings.Contains(err.Error(), "failed to get directory specified by the source") {
				continue
			}
			return nil, err
		}
		matchedPlanFilePaths = append(matchedPlanFilePaths, files...)
	}

	// Sanitize the matches to ignore the directories
	for _, i := range matchedPlanFilePaths {

		// Ignore directories
		if filehelpers.DirectoryExists(i) {
			continue
		}
		d.StreamListItem(ctx, filePath{
			Path:             i,
			IsTFPlanFilePath: true,
		})
	}

	// Gather TF state file path matches for the glob
	var matchedStateFilePaths []string
	stateFilePaths := terraformConfig.StateFilePaths
	for _, i := range stateFilePaths {

		// List the files in the given source directory
		files, err := d.GetSourceFiles(i)
		if err != nil {
			plugin.Logger(ctx).Error("tfConfigList.stateFilePaths", "get_source_files_error", err)

			// If the specified path is unavailable, then an empty row should populate
			if strings.Contains(err.Error(), "failed to get directory specified by the source") {
				continue
			}
			return nil, err
		}
		matchedStateFilePaths = append(matchedStateFilePaths, files...)
	}

	// Sanitize the matches to ignore the directories
	for _, i := range matchedStateFilePaths {

		// Ignore directories
		if filehelpers.DirectoryExists(i) {
			continue
		}
		d.StreamListItem(ctx, filePath{
			Path:              i,
			IsTFStateFilePath: true,
		})
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

		// check if the arguments interface is nil
		if v != nil {
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

	parsedDocs, err := p.Parse(path, content, false, false)
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

func isTerraformPlan(content []byte) bool {
	var data map[string]interface{}
	err := json.Unmarshal(content, &data)
	if err != nil {
		return false
	}

	// Check for fields that are common in Terraform plans
	_, hasResourceChanges := data["resource_changes"]
	_, hasFormatVersion := data["format_version"]

	return hasResourceChanges && hasFormatVersion
}

// findBlockLinesFromJSON locates the start and end lines of a specific block or nested element within a block.
// The file should contain structured data (e.g., JSON) and this function expects to search for blocks with specific names.
func findBlockLinesFromJSON(ctx context.Context, path string, blockName string, pathName ...string) (int, int, string, error) {
	var currentLine, startLine, endLine int
	var bracketCounter, startCounter int

	// These boolean flags indicate which part of the structured data we're currently processing.
	inBlock, inOutput, inTargetBlock := false, false, false

	file, err := os.Open(path)
	if err != nil {
		plugin.Logger(ctx).Error("findBlockLinesFromJSON", "file_error", err)
		return startLine, endLine, "", err
	}

	// Move the file pointer to the start of the file.
	_, _ = file.Seek(0, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Detect the start of the desired block, path, response, etc.
		// Depending on the blockName and provided pathName, different conditions are checked.

		// Generic block detection
		if !inBlock && (trimmedLine == fmt.Sprintf(`"%s": {`, blockName) || trimmedLine == fmt.Sprintf(`"%s": [`, blockName)) {
			inBlock = true
			startLine = currentLine
			continue
		} else if inBlock && blockName == "outputs" && trimmedLine == fmt.Sprintf(`"%s": {`, pathName[0]) {
			// Different output block detection within the "outputs" block
			inOutput = true
			bracketCounter = 1
			startLine = currentLine
			continue
		} else if inBlock && blockName == "resources" {
			if inBlock && strings.Contains(trimmedLine, "{") {
				bracketCounter++
				startCounter = currentLine
			}
			if inBlock && strings.Contains(trimmedLine, "}") {
				bracketCounter--
			}

			// Get the start line info for the plan file data
			// For terraform plan we need a special handling since
			// if we use the count or for_each, in that case the resource configurations in the terraform plan can have more than 1 resource object with same name and type.
			// So, to avoid the conflict use address and type instead which is unique and only applicable for terraform plan file.
			if inBlock && strings.Contains(trimmedLine, fmt.Sprintf(`"address": "%s"`, pathName[0])) {
				peekCounter := 1
				nameFound := false

				for {
					peekLine, _ := readLineN(file, currentLine+peekCounter)
					if strings.Contains(peekLine, fmt.Sprintf(`"type": "%s"`, pathName[1])) {
						nameFound = true
						break
					}
					if strings.Contains(peekLine, "}") {
						break
					}
					peekCounter++
				}

				if nameFound {
					inTargetBlock = true
					startLine = startCounter // Assume the opening brace is at the start of this resource
				}
			}

			// Get the start line info from terraform state file.
			// Match the type and name of the resource to get the start position
			if inBlock && strings.Contains(trimmedLine, fmt.Sprintf(`"type": "%s"`, pathName[0])) {
				peekCounter := 1
				nameFound := false

				for {
					peekLine, _ := readLineN(file, currentLine+peekCounter)
					if strings.Contains(peekLine, fmt.Sprintf(`"name": "%s"`, pathName[1])) {
						nameFound = true
						break
					}
					if strings.Contains(peekLine, "}") {
						break
					}
					peekCounter++
				}

				if nameFound {
					inTargetBlock = true
					startLine = startCounter // Assume the opening brace is at the start of this resource
				}
			}
		}
		// If we are within a block, we need to track the opening and closing brackets
		// to determine where the block ends.
		if inBlock && inOutput && !inTargetBlock {
			bracketCounter += strings.Count(line, "{")
			bracketCounter -= strings.Count(line, "}")

			if bracketCounter == 0 {
				endLine = currentLine
				break
			}
		}

		if inBlock && inTargetBlock && bracketCounter == 0 {
			endLine = currentLine
			break
		}
	}
	source := getSourceFromFile(file, startLine, endLine)

	if startLine != 0 && endLine == 0 {
		// If we found the start but not the end, reset the start to indicate the block doesn't exist in entirety.
		startLine = 0
	}

	// By default (when created), the file content is not properly formatted with indentation and all the content remains in line 1
	if file != nil && startLine == 0 && endLine == 0 {
		// Set the start line as 1, and
		// end line as the current line (i.e. total lines)
		startLine = 1
		endLine = currentLine

		content, err := os.ReadFile(path)
		if err != nil {
			plugin.Logger(ctx).Error("findBlockLinesFromJSON", "read_file_error", err)
			return startLine, endLine, source, err
		}
		contentStr := string(content)

		// Regex pattern to extract the resources list from the file
		pattern := `"planned_values":{.*"root_module":{"resources":(.*)}},"resource_changes"`

		// Compile the regular expression
		re := regexp.MustCompile(pattern)

		// Find the match in the JSON string
		matches := re.FindStringSubmatch(contentStr)

		// Check if the resources block is present in the plan file content store the resources list
		var resources []interface{}
		if len(matches) >= 2 {
			plannedValues := matches[1]
			err := json.Unmarshal([]byte(plannedValues), &resources)
			if err != nil {
				plugin.Logger(ctx).Error("findBlockLinesFromJSON", "unmarshal_error", err)
				return startLine, endLine, source, err
			}
		}

		// Go through the resources and check for the desired one
		for _, r := range resources {
			if strings.Contains(fmt.Sprint(r), pathName[0]) && strings.Contains(fmt.Sprint(r), pathName[1]) {
				if data, ok := r.(map[string]interface{}); ok {
					// Marshal the map to JSON
					jsonBytes, err := json.Marshal(data)
					if err != nil {
						plugin.Logger(ctx).Error("findBlockLinesFromJSON", "unmarshal_error", err)
						return startLine, endLine, source, err
					}

					// Convert the JSON bytes to a string
					jsonString := string(jsonBytes)
					// And, set the value as source
					source = jsonString
				}
			}
		}
	}

	return startLine, endLine, source, nil
}

func getSourceFromFile(file *os.File, startLine int, endLine int) string {
	var source string
	_, _ = file.Seek(0, 0) // Go to the start
	scanner := bufio.NewScanner(file)
	currentSourceLine := 0
	for scanner.Scan() {
		currentSourceLine++
		if currentSourceLine >= startLine && currentSourceLine <= endLine {
			source += scanner.Text() + "\n"
		}
		if currentSourceLine > endLine {
			break
		}
	}
	return source
}

func readLineN(file *os.File, lineNum int) (string, error) {
	_, _ = file.Seek(0, 0) // Go to the start
	scanner := bufio.NewScanner(file)
	currentLine := 0
	for scanner.Scan() {
		currentLine++
		if currentLine == lineNum {
			return scanner.Text(), nil
		}
	}
	return "", nil
}

// Transform function to return nil if an empty map
func NullIfEmptyMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if data, isMap := d.Value.(map[string]interface{}); isMap {
		if len(data) == 0 {
			return nil, nil
		}
	}
	return d.Value, nil
}
