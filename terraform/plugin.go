/*
Package terraform implements a steampipe plugin for terraform.

This plugin provides data that Steampipe uses to present foreign
tables that represent Terraform resources.
*/
package terraform

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
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
			"terraform_resource": tableTerraformResource(ctx),
		},
	}

	return p
}
