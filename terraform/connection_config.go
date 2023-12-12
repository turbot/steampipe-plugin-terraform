package terraform

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type terraformConfig struct {
	ConfigurationFilePaths []string `cty:"configuration_file_paths" steampipe:"watch"`
	Paths                  []string `cty:"paths" steampipe:"watch"`
	PlanFilePaths          []string `cty:"plan_file_paths" steampipe:"watch"`
	StateFilePaths         []string `cty:"state_file_paths" steampipe:"watch"`
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
