# Table: terraform_resource

Each resource block describes one or more infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

## Examples

### Basic info

```sql
select
  name,
  type,
  arguments
from
  terraform_resource;
```

### List EC2 instance resources

```sql
select
  name,
  type,
  arguments
from
  terraform_resource
where
  type = 'aws_instance';
```
