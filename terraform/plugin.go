/*
Package terraform implements a steampipe plugin for terraform.

This plugin provides data that Steampipe uses to present foreign
tables that represent Terraform resources.
*/
package terraform

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin/transform"
)

const pluginName = "steampipe-plugin-terraform"

// Plugin creates this (terraform) plugin
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             pluginName,
		DefaultTransform: transform.FromCamel().NullIfZero(),
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		TableMap: map[string]*plugin.Table{
			"terraform_data_source": tableTerraformDataSource(ctx),
			"terraform_local":       tableTerraformLocal(ctx),
			"terraform_output":      tableTerraformOutput(ctx),
			"terraform_provider":    tableTerraformProvider(ctx),
			"terraform_resource":    tableTerraformResource(ctx),
		},
	}

	return p
}
