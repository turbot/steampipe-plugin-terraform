![image](https://hub.steampipe.io/images/plugins/turbot/terraform-social-graphic.png)

# Terraform Plugin for Steampipe

Use SQL to query data from Terraform configuration files.

- **[Get started →](https://hub.steampipe.io/plugins/turbot/terraform)**
- Documentation: [Table definitions & examples](https://hub.steampipe.io/plugins/turbot/terraform/tables)
- Community: [Slack Channel](https://steampipe.io/community/join)
- Get involved: [Issues](https://github.com/turbot/steampipe-plugin-terraform/issues)

## Quick start

Install the plugin with [Steampipe](https://steampipe.io):

```shell
steampipe plugin install terraform
```

Run a query:

```sql
select
  name,
  type,
  arguments
from
  terraform_resource;
```

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
