![image](https://hub.steampipe.io/images/plugins/turbot/terraform-social-graphic.png)

# Terraform Plugin for Steampipe

Use SQL to query data from Terraform configuration files.

- **[Get started →](https://hub.steampipe.io/plugins/turbot/terraform)**
- Documentation: [Table definitions & examples](https://hub.steampipe.io/plugins/turbot/terraform/tables)
- Community: [Join #steampipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/turbot/steampipe-plugin-terraform/issues)

## Quick start

Install the plugin with [Steampipe](https://steampipe.io):

```shell
steampipe plugin install terraform
```

Configure your [config file](https://hub.steampipe.io/plugins/turbot/terraform#configuration) to include directories with Terraform configuration files. If no directory is specified, the current working directory will be used.

Run steampipe:

```shell
steampipe query
```

Query all resources in your Terraform files:

```sql
select
  name,
  type,
  jsonb_pretty(arguments) as args
from
  terraform_resource;
```

```
> select name, type, jsonb_pretty(arguments) as args from terraform_resource;
+------------+----------------+--------------------------------------------+
| name       | type           | args                                       |
+------------+----------------+--------------------------------------------+
| app_server | aws_instance   | {                                          |
|            |                |     "ami": "ami-830c94e3",                 |
|            |                |     "tags": {                              |
|            |                |         "Name": "ExampleAppServerInstance" |
|            |                |     },                                     |
|            |                |     "instance_type": "t2.micro"            |
|            |                | }                                          |
| app_volume | aws_ebs_volume | {                                          |
|            |                |     "size": 40,                            |
|            |                |     "tags": {                              |
|            |                |         "Name": "HelloWorld"               |
|            |                |     },                                     |
|            |                |     "availability_zone": "us-west-2a"      |
|            |                | }                                          |
| app_bucket | aws_s3_bucket  | {                                          |
|            |                |     "acl": "private",                      |
|            |                |     "tags": {                              |
|            |                |         "Name": "Test bucket",             |
|            |                |         "Environment": "Dev"               |
|            |                |     },                                     |
|            |                |     "bucket": "my-app-bucket"              |
|            |                | }                                          |
+------------+----------------+--------------------------------------------+
```

## Engines

This plugin is available for the following engines:

| Engine        | Description
|---------------|------------------------------------------
| [Steampipe](https://steampipe.io/docs) | The Steampipe CLI exposes APIs and services as a high-performance relational database, giving you the ability to write SQL-based queries to explore dynamic data. Mods extend Steampipe's capabilities with dashboards, reports, and controls built with simple HCL. The Steampipe CLI is a turnkey solution that includes its own Postgres database, plugin management, and mod support.
| [Postgres FDW](https://steampipe.io/docs/steampipe_postgres/index) | Steampipe Postgres FDWs are native Postgres Foreign Data Wrappers that translate APIs to foreign tables. Unlike Steampipe CLI, which ships with its own Postgres server instance, the Steampipe Postgres FDWs can be installed in any supported Postgres database version.
| [SQLite Extension](https://steampipe.io/docs//steampipe_sqlite/index) | Steampipe SQLite Extensions provide SQLite virtual tables that translate your queries into API calls, transparently fetching information from your API or service as you request it.
| [Export](https://steampipe.io/docs/steampipe_export/index) | Steampipe Plugin Exporters provide a flexible mechanism for exporting information from cloud services and APIs. Each exporter is a stand-alone binary that allows you to extract data using Steampipe plugins without a database.
| [Turbot Pipes](https://turbot.com/pipes/docs) | Turbot Pipes is the only intelligence, automation & security platform built specifically for DevOps. Pipes provide hosted Steampipe database instances, shared dashboards, snapshots, and more.

## Developing

Prerequisites:

- [Steampipe](https://steampipe.io/downloads)
- [Golang](https://golang.org/doc/install)

Clone:

```sh
git clone https://github.com/turbot/steampipe-plugin-terraform.git
cd steampipe-plugin-terraform
```

Build, which automatically installs the new version to your `~/.steampipe/plugins` directory:

```
make
```

Configure the plugin:

```
cp config/* ~/.steampipe/config
vi ~/.steampipe/config/terraform.spc
```

Try it!

```
steampipe query
> .inspect terraform
```

Further reading:

- [Writing plugins](https://steampipe.io/docs/develop/writing-plugins)
- [Writing your first table](https://steampipe.io/docs/develop/writing-your-first-table)

## Contributing

Please see the [contribution guidelines](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/turbot/steampipe/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [Apache 2.0 open source license](https://github.com/turbot/steampipe-plugin-terraform/blob/main/LICENSE).

`help wanted` issues:

- [Steampipe](https://github.com/turbot/steampipe/labels/help%20wanted)
- [Terraform Plugin](https://github.com/turbot/steampipe-plugin-terraform/labels/help%20wanted)
