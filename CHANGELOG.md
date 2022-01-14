## v0.0.3 [2022-01-14]

_Enhancements_

- Recompiled plugin with [steampipe-plugin-sdk v1.8.3](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v183--2021-12-23)

_Bug fixes_

- Fixed `terraform_local` queries intermittently failing due to a value conversion error

## v0.0.2 [2021-12-16]

_Enhancements_

- Recompiled plugin with Go version 1.17 ([#4](https://github.com/turbot/steampipe-plugin-terraform/pull/4))

## v0.0.1 [2021-12-02]

_What's new?_

- New tables added

  - [terraform_data_source](https://hub.steampipe.io/plugins/turbot/terraform/tables/terraform_data_source)
  - [terraform_local](https://hub.steampipe.io/plugins/turbot/terraform/tables/terraform_local)
  - [terraform_output](https://hub.steampipe.io/plugins/turbot/terraform/tables/terraform_output)
  - [terraform_provider](https://hub.steampipe.io/plugins/turbot/terraform/tables/terraform_provider)
  - [terraform_resource](https://hub.steampipe.io/plugins/turbot/terraform/tables/terraform_resource)
