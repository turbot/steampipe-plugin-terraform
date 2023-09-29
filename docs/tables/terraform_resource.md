# Table: terraform_resource

Each resource block describes one or more infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

## Examples

### Basic info

```sql
select
  name,
  type,
  address,
  attributes_std,
  path
from
  terraform_resource;
```

### List AWS IAM roles

```sql
select
  name,
  type,
  address,
  attributes_std,
  path
from
  terraform_resource
where
  type = 'aws_iam_role';
```

### List AWS IAM `assume_role_policy` Statements

```sql
select
  path,
  name,
  address,
  (attributes_std ->> 'assume_role_policy')::jsonb -> 'Statement' as statement
from
  terraform_resource
where
  type = 'aws_iam_role'
```

### Get AMI for each AWS EC2 instance

```sql
select
  address,
  name,
  attributes_std ->> 'ami' as ami,
  path
from
  terraform_resource
where
  type = 'aws_instance';
```

### List AWS CloudTrail trails that are not encrypted

```sql
select
  address,
  name,
  path
from
  terraform_resource
where
  type = 'aws_cloudtrail'
  and attributes_std -> 'kms_key_id' is null;
```

### List Azure storage accounts that allow public blob access

```sql
select
  address,
  name,
  case
    when attributes_std -> 'allow_blob_public_access' is null then false
    else (attributes_std -> 'allow_blob_public_access')::boolean
  end as allow_blob_public_access,
  path
from
  terraform_resource
where
  type = 'azurerm_storage_account'
  -- Optional arg that defaults to false
  and (attributes_std -> 'allow_blob_public_access')::boolean;
```

### List Azure MySQL servers that don't enforce SSL

```sql
select
  address,
  name,
  attributes_std -> 'ssl_enforcement_enabled' as ssl_enforcement_enabled,
  path
from
  terraform_resource
where
  type = 'azurerm_mysql_server'
  and not (attributes_std -> 'ssl_enforcement_enabled')::boolean;
```

### List Azure MySQL servers with public network access enabled

```sql
select
  address,
  name,
  case
    when attributes_std -> 'public_network_access_enabled' is null then true
    else (attributes_std -> 'public_network_access_enabled')::boolean
  end as public_network_access_enabled,
  path
from
  terraform_resource
where
  type in ('azurerm_mssql_server', 'azurerm_mysql_server')
  -- Optional arg that defaults to true
  and (attributes_std -> 'public_network_access_enabled' is null or (attributes_std -> 'public_network_access_enabled')::boolean);
```

### List resources from a plan file

```sql
select
  name,
  type,
  address,
  attributes_std,
  path
from
  terraform_resource
where
  path = '/path/to/tfplan.json';
```

### List resources from a state file

```sql
select
  name,
  type,
  address,
  attributes_std,
  path
from
  terraform_resource
where
  path = '/path/to/terraform.tfstate';
```
