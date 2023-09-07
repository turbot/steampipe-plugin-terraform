## v0.8.0 [2023-09-07]

_Enhancements_

- Added support to parse Terraform plan and state files to get resource information. ([#40](https://github.com/turbot/steampipe-plugin-terraform/pull/40))

## v0.7.0 [2023-06-20]

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.5.0](https://github.com/turbot/steampipe-plugin-sdk/blob/v5.5.0/CHANGELOG.md#v550-2023-06-16) which significantly reduces API calls and boosts query performance, resulting in faster data retrieval. This update significantly lowers the plugin initialization time of dynamic plugins by avoiding recursing into child folders when not necessary. ([#41](https://github.com/turbot/steampipe-plugin-terraform/pull/41))

## v0.6.0 [2023-06-07]

_Bug fixes_

- Fixed the `arguments` column of `terraform_module` table to correctly return data instead of `null`. ([#36](https://github.com/turbot/steampipe-plugin-terraform/pull/36)) (Thanks [@rollwagen](https://github.com/rollwagen) for the contribution!!)

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.4.1](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v541-2023-05-05) which fixes increased plugin initialization time due to multiple connections causing the schema to be loaded repeatedly. ([#38](https://github.com/turbot/steampipe-plugin-terraform/pull/38))

## v0.5.0 [2023-04-11]

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.3.0](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v530-2023-03-16) which includes fixes for query cache pending item mechanism and aggregator connections not working for dynamic tables. ([#33](https://github.com/turbot/steampipe-plugin-terraform/pull/33))

## v0.4.0 [2023-02-16]

_What's new?_

- New tables added
  - [terraform_module](https://hub.steampipe.io/plugins/turbot/terraform/tables/terraform_module) ([#28](https://github.com/turbot/steampipe-plugin-terraform/pull/28)) (Thanks [@rollwagen](https://github.com/rollwagen) for the contribution!!)

## v0.3.0 [2022-11-16]

_What's new?_

- Added support for retrieving Terraform configuration files from remote Git repositories and S3 buckets. For more information, please see [Supported Path Formats](https://hub.steampipe.io/plugins/turbot/terraform#supported-path-formats). ([#25](https://github.com/turbot/steampipe-plugin-terraform/pull/25))
- Added file watching support for files included in the `paths` config argument. ([#25](https://github.com/turbot/steampipe-plugin-terraform/pull/25))

_Enhancements_

- Added `end_line` and `source` columns to all tables. ([#25](https://github.com/turbot/steampipe-plugin-terraform/pull/25))

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v5.0.0](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v500-2022-11-16) which includes support for fetching remote files with go-getter and file watching. ([#25](https://github.com/turbot/steampipe-plugin-terraform/pull/25))

## v0.2.0 [2022-09-09]

_Dependencies_

- Recompiled plugin with [steampipe-plugin-sdk v4.1.6](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v416-2022-09-02) which includes several caching and memory management improvements. ([#21](https://github.com/turbot/steampipe-plugin-terraform/pull/21))
- Recompiled plugin with Go version `1.19`. ([#21](https://github.com/turbot/steampipe-plugin-terraform/pull/21))

## v0.1.0 [2022-04-28]

_Enhancements_

- Added support for native Linux ARM and Mac M1 builds. ([#17](https://github.com/turbot/steampipe-plugin-terraform/pull/17))
- Recompiled plugin with [steampipe-plugin-sdk v3.1.0](https://github.com/turbot/steampipe-plugin-sdk/blob/main/CHANGELOG.md#v310--2022-03-30) and Go version `1.18`. ([#16](https://github.com/turbot/steampipe-plugin-terraform/pull/16))

## v0.0.5 [2022-02-10]

_What's new?_

- File loading and matching through the `paths` argument has been updated to make the plugin easier to use:
  - The `paths` argument is no longer commented out by default for new plugin installations and now defaults to the current working directory
  - Home directory expansion (`~`) is now supported
  - Recursive directory searching (`**`) is now supported
- Previously, when using wildcard matching (`*`), non-Terraform configuration files were automatically excluded to prevent parsing errors. These files are no longer automatically excluded to allow for a wider range of matches. If your current configuration uses wildcard matching, e.g., `paths = [ "/path/to/my/files/*" ]`, please update it to include the file extension, e.g., `paths = [ "/path/to/my/files/*.tf" ]`.

## v0.0.4 [2022-02-01]

_Bug fixes_

- Fixed: Add lock to file parsing function to prevent concurrent map read/write errors ([#9](https://github.com/turbot/steampipe-plugin-terraform/pull/9))

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
