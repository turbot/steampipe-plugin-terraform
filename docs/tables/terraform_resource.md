---
title: "Steampipe Table: terraform_resource - Query Terraform Resources using SQL"
description: "Allows users to query Terraform Resources, specifically the configuration, state, and provider details, providing insights into resource management and potential configuration issues."
---

# Table: terraform_resource - Query Terraform Resources using SQL

Terraform is an open-source infrastructure as code software tool that enables users to define and provision a data center infrastructure using a high-level configuration language. It supports a multitude of providers such as AWS, GCP, Azure, and more. The Terraform Resources are the main component in a Terraform configuration, and they describe one or more infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

## Table Usage Guide

The `terraform_resource` table provides insights into Terraform Resources within the Terraform environment. As a DevOps engineer, explore resource-specific details through this table, including configuration, state, and provider details. Utilize it to uncover information about resources, such as their current state, the provider they are associated with, and the details of their configuration.

## Examples

### Basic info
Explore the fundamental details of your Terraform resources to gain a better understanding of their configuration and location. This can be beneficial in managing resources and assessing their setup.Explore which resources are currently in use within your Terraform configuration. This allows you to gain insights into the types, addresses, and paths of these resources, aiding you in your infrastructure management tasks.


```sql+postgres
select
  name,
  type,
  address,
  attributes_std,
  path
from
  terraform_resource;
```

```sql+sqlite
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
Explore the configuration of your AWS infrastructure by identifying the roles assigned within it. This can help in understanding the access and permissions structure, aiding in security audits and compliance checks.Explore the various roles within your AWS IAM setup to understand their configurations and attributes. This could help in managing access control and ensuring security protocols are being followed.


```sql+postgres
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

```sql+sqlite
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
Explore which AWS Identity and Access Management (IAM) roles have specific permissions. This is particularly useful for auditing security and compliance purposes, as it allows you to identify potential vulnerabilities in your IAM roles' permissions.Analyze the settings to understand the policies associated with AWS IAM roles. This can be useful to identify instances where specific roles have been granted certain permissions, ensuring secure and appropriate access control within your AWS environment.


```sql+postgres
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

```sql+sqlite
select
  path,
  name,
  address,
  json_extract(attributes_std, '$.assume_role_policy.Statement') as statement
from
  terraform_resource
where
  type = 'aws_iam_role'
```

### Get AMI for each AWS EC2 instance
Explore which AWS EC2 instances are associated with each Amazon Machine Image (AMI). This can help identify instances that may be using outdated or unsecured AMIs, supporting better security and compliance management.Explore which Amazon Machine Images (AMIs) are used for each Amazon Web Services (AWS) Elastic Compute Cloud (EC2) instance. This is useful for understanding the software configurations of your EC2 instances.


```sql+postgres
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

```sql+sqlite
select
  address,
  name,
  json_extract(attributes_std, '$.ami') as ami,
  path
from
  terraform_resource
where
  type = 'aws_instance';
```

### List AWS CloudTrail trails that are not encrypted
Analyze the settings to understand which AWS CloudTrail trails are not encrypted, helping to identify potential security risks in your AWS environment.Determine the areas in which AWS CloudTrail trails are not encrypted to ensure data security and compliance. This is crucial for identifying potential security vulnerabilities in your AWS environment.


```sql+postgres
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

```sql+sqlite
select
  address,
  name,
  path
from
  terraform_resource
where
  type = 'aws_cloudtrail'
  and json_extract(attributes_std, '$.kms_key_id') is null;
```

### List Azure storage accounts that allow public blob access
Explore which Azure storage accounts permit public access to their blobs. This is useful in identifying potential security vulnerabilities where sensitive data might be exposed.Explore which Azure storage accounts permit public blob access. This can be useful in identifying potential security risks and ensuring that sensitive data is not inadvertently exposed to the public.


```sql+postgres
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

```sql+sqlite
select
  address,
  name,
  case
    when json_extract(attributes_std, '$.allow_blob_public_access') is null then 0
    else json_extract(attributes_std, '$.allow_blob_public_access')
  end as allow_blob_public_access,
  path
from
  terraform_resource
where
  type = 'azurerm_storage_account'
  and json_extract(attributes_std, '$.allow_blob_public_access');
```

### List Azure MySQL servers that don't enforce SSL
Explore which Azure MySQL servers are potentially vulnerable by identifying those that do not enforce SSL. This can help enhance security by pinpointing areas to strengthen encryption measures.Determine the areas in which Azure MySQL servers are not enforcing SSL. This is useful to identify potential security vulnerabilities and ensure all servers are adhering to best practices for secure connections.


```sql+postgres
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

```sql+sqlite
select
  address,
  name,
  json_extract(attributes_std, '$.ssl_enforcement_enabled') as ssl_enforcement_enabled,
  path
from
  terraform_resource
where
  type = 'azurerm_mysql_server'
  and not json_extract(attributes_std, '$.ssl_enforcement_enabled');
```

### List Azure MySQL servers with public network access enabled
Determine the Azure MySQL servers that have public network access enabled. This can be useful for identifying potential security risks and ensuring that your servers are configured according to your organization's security policies.Determine the Azure MySQL servers that have public network access enabled. This helps in identifying potential security risks by highlighting servers that are exposed to the public internet.


```sql+postgres
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

```sql+sqlite
select
  address,
  name,
  case
    when json_extract(attributes_std, '$.public_network_access_enabled') is null then 1
    else json_extract(attributes_std, '$.public_network_access_enabled')
  end as public_network_access_enabled,
  path
from
  terraform_resource
where
  type in ('azurerm_mssql_server', 'azurerm_mysql_server')
  and (json_extract(attributes_std, '$.public_network_access_enabled') is null or json_extract(attributes_std, '$.public_network_access_enabled'));
```

### List resources from a plan file
This query allows you to analyze the resources outlined in a specific Terraform plan file. It helps in gaining insights into the different elements like name, type, and address, which can be beneficial for understanding the structure and configuration of your infrastructure.Explore which resources are included in a specific plan file. This can help identify instances where certain resources may need to be added, removed, or modified, providing insights into the overall configuration of your project.

```sql+postgres
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

```sql+sqlite
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
Explore which resources are contained within a specific state file. This is useful for understanding the structure and content of your Terraform infrastructure without needing to navigate through multiple files or directories.Determine the resources within a specific state file in Terraform. This is useful for understanding the components of your infrastructure and their attributes, especially when managing large-scale deployments.


```sql+postgres
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

```sql+sqlite
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