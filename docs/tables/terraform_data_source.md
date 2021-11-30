# Table: terraform_data_source

Data sources allow Terraform use information defined outside of Terraform, defined by another separate Terraform configuration, or modified by functions.

## Examples

### Basic info

```sql
select
  name,
  type,
  arguments
from
  terraform_data_source;
```
