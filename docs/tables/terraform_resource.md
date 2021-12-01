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

### List AWS IAM roles

```sql
select
  name,
  type,
  arguments
from
  terraform_resource
where
  type = 'aws_iam_role';
```

### Get AMI for each AWS EC2 instance

```sql
select
  name,
  arguments ->> 'ami' as ami
from
  terraform_resource
where
  type = 'aws_instance';
```

### List AWS CloudTrail trails that are not encrypted

```sql
select
  name
from
  terraform_resource
where
  type = 'aws_cloudtrail'
  and arguments -> 'kms_key_id' is null;
```

### List Azure storage accounts that allow public blob access

```sql
select
  name,
  case
    when arguments -> 'allow_blob_public_access' is null then false
    else (arguments -> 'allow_blob_public_access')::boolean
  end as allow_blob_public_access
from
  terraform_resource
where
  type = 'azurerm_storage_account'
  -- Optional arg that defaults to false
  and (arguments -> 'allow_blob_public_access')::boolean;
```

### List Azure MySQL servers that don't enforce SSL

```sql
select
  name,
  arguments -> 'ssl_enforcement_enabled' as ssl_enforcement_enabled
from
  terraform_resource
where
  type = 'azurerm_mysql_server'
  and not (arguments -> 'ssl_enforcement_enabled')::boolean;
```

### List Azure MySQL servers with public network access enabled

```sql
select
  name,
  case
    when arguments -> 'public_network_access_enabled' is null then true
    else (arguments -> 'public_network_access_enabled')::boolean
  end as public_network_access_enabled
from
  terraform_resource
where
  type in ('azurerm_mssql_server', 'azurerm_mysql_server')
  -- Optional arg that defaults to true
  and (arguments -> 'public_network_access_enabled' is null or (arguments -> 'public_network_access_enabled')::boolean);
```

