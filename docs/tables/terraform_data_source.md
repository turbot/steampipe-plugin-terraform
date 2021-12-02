# Table: terraform_data_source

Data sources allow Terraform use information defined outside of Terraform, defined by another separate Terraform configuration, or modified by functions.

## Examples

### Basic info

```sql
select
  name,
  type,
  arguments,
  path
from
  terraform_data_source;
```

### List AWS EC2 AMIs

```sql
select
  name,
  type,
  arguments,
  path
from
  terraform_data_source
where
  type = 'aws_ami'
```

### Get filters for each AWS EC2 AMI

```sql
with filters as (
select
  name,
  type,
  jsonb_array_elements(arguments -> 'filter') as filter,
  path
from
  terraform_data_source
where
  type = 'aws_ami'
)
select
  name,
  type,
  filter -> 'name' as name,
  filter -> 'values' as values,
  path
from
  filters;
```
