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
