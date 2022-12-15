# Table: terraform_module

Modules are containers for multiple resources that are used together.

The `source` argument in a module block tells Terraform where to find
the source code for the desired child module. Due to name clashes, the
column name for the `source` argument is `module_source`.

Registry modules support versioning via the `version` argument.

## Examples

### Basic info

```sql
select
  name,
  module_source,
  version
from
  terraform_module;
```

### List all modules that reference a source on 'gitlab.com' but don't use a version number as reference

```sql
select
  name,
  split_part(module_source,'=',-1) as ref
from
  terraform_module
where
  module_source like '%gitlab.com%'
  and not split_part(module_source,'=',-1) ~ '^[0-9]';
```
