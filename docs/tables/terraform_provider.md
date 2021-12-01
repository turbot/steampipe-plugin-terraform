# Table: terraform_provider

Each provider adds a set of resource types and/or data sources that Terraform can manage.

Most providers configure a specific infrastructure platform (either cloud or self-hosted). Providers can also offer local utilities for tasks like generating random numbers for unique resource names.

## Examples

### Basic info

```sql
select
  name,
  alias,
  arguments
from
  terraform_provider;
```
