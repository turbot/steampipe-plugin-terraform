# Table: terraform_provider

Each provider adds a set of resource types and/or data sources that Terraform can manage.

Most providers configure a specific infrastructure platform (either cloud or self-hosted). Providers can also offer local utilities for tasks like generating random numbers for unique resource names.

## Examples

### Basic info

```sql
select
  name,
  alias,
  arguments,
  path
from
  terraform_provider;
```

### List providers using deprecated 'version' argument

```sql
select
  name,
  alias,
  version,
  path
from
  terraform_provider
where
  version is not null;
```

### List AWS providers with their regions

```sql
select
  name,
  alias,
  arguments ->> 'region' as region,
  path
from
  terraform_provider
where
  name = 'aws';
```
