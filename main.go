package main

import (
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe-plugin-terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: terraform.Plugin})
}
