package terraform

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type terraformConfig struct {
	Paths []string `cty:"paths" steampipe:"watch"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"paths": {
		Type: schema.TypeList,
		Elem: &schema.Attribute{Type: schema.TypeString},
	},
}

func ConfigInstance() interface{} {
	return &terraformConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) terraformConfig {
	if connection == nil || connection.Config == nil {
		return terraformConfig{}
	}
	config, _ := connection.Config.(terraformConfig)
	return config
}
