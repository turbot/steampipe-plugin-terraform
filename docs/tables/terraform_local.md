# Table: terraform_local

A local value assigns a name to an expression, so you can use it multiple times within a module without repeating it.

## Examples

### Basic info

```sql
select
  name,
  value,
  path
from
  terraform_local;
```

### List 'Owner' locals (case insensitive)

```sql
select
  name,
  value,
  path
from
  terraform_local
where
  name ilike 'owner'
```

### Detect secrets in local values (requires Code plugin)

```sql
select
  name as local_name,
  path as file_path,
  secret_type,
  secret,
  authenticated,
  line,
  col
from
  code_secret,
  terraform_local
where
  src = value::text;
```
