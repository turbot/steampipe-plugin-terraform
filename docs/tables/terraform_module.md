---
title: "Steampipe Table: terraform_module - Query Terraform Modules using SQL"
description: "Allows users to query Terraform Modules, specifically the module source, version, and other metadata, providing insights into the configuration and usage of Terraform modules."
---

# Table: terraform_module - Query Terraform Modules using SQL

Terraform Modules are a set of Terraform resources that are grouped together and managed as a single entity. They provide a way to encapsulate common service configurations and reuse them across multiple environments or projects. Modules help in managing complex infrastructure setups by breaking them down into smaller, manageable components.

## Table Usage Guide

The `terraform_module` table provides insights into Terraform Modules within Terraform. As a DevOps engineer, explore module-specific details through this table, including source, version, and other metadata. Utilize it to uncover information about modules, such as their configuration, usage across different environments or projects, and the management of complex infrastructure setups.

**Important Notes**

- The `source` argument in a module block tells Terraform where to find
the source code for the desired child module. Due to name clashes, the
column name for the `source` argument is `module_source`.
- Registry modules support versioning via the `version` argument.

## Examples

### Basic info
Explore the different modules in your Terraform configuration to understand their source and version. This can help ensure you're using the most up-to-date and secure modules in your infrastructure.Discover the segments that are using different versions of Terraform modules. This can help in managing updates and ensuring consistency across your infrastructure.


```sql+postgres
select
  name,
  module_source,
  version
from
  terraform_module;
```

```sql+sqlite
select
  name,
  module_source,
  version
from
  terraform_module;
```

### List all modules that reference a source on 'gitlab.com' but don't use a version number as reference
This example highlights the identification of Terraform modules that reference sources on 'gitlab.com' but do not utilize a version number for referencing. This is useful for ensuring proper version control and avoiding potential inconsistencies or conflicts in your infrastructure setup.Explore modules that link to 'gitlab.com' but lack a specified version number. This is useful for identifying potential areas of instability in your infrastructure, as modules without version numbers can introduce unpredictability.


```sql+postgres
select
  name,
  split_part(module_source,'=',-1) as ref
from
  terraform_module
where
  module_source like '%gitlab.com%'
  and not split_part(module_source,'=',-1) ~ '^[0-9]';
```

```sql+sqlite
Error: SQLite does not support split_part function and regular expression matching like '~'.
```