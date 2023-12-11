---
title: "Steampipe Table: terraform_output - Query Terraform Outputs using SQL"
description: "Allows users to query Terraform Outputs, thus providing a means to extract information about the outputs from Terraform state files."
---

# Table: terraform_output - Query Terraform Outputs using SQL

Terraform Outputs serve as a means to extract data from a Terraform state. These outputs can be simple strings or complex data structures such as lists or maps. They provide a way to share information between modules, access computed values, and manage resource configurations.

Output values are like the return values of a Terraform module, and have several uses:

- A child module can use outputs to expose a subset of its resource attributes to a parent module.
- A root module can use outputs to print certain values in the CLI output after running terraform apply.
- When using remote state, root module outputs can be accessed by other configurations via a terraform_remote_state data source.

## Table Usage Guide

The `terraform_output` table provides insights into the outputs from Terraform state files. As a DevOps engineer, you can explore output-specific details through this table, including the values, types, and associated state files. Utilize it to manage and monitor your Terraform infrastructure, ensuring configurations are as expected and aiding in troubleshooting.

## Examples

### Basic info
Discover the segments that contain specific values within your Terraform outputs. This can be particularly useful in understanding the distribution and organization of your data.Explore the key details of your Terraform configuration outputs. This can help you understand the values and paths associated with different elements of your configuration, and can be useful in troubleshooting or optimizing your setup.


```sql+postgres
select
  name,
  description,
  value,
  path
from
  terraform_output;
```

```sql+sqlite
select
  name,
  description,
  value,
  path
from
  terraform_output;
```

### List sensitive outputs
Discover the segments that contain sensitive information within your Terraform outputs. This can help in identifying potential security risks and take necessary precautions to protect your data.Explore which outputs in your Terraform configuration are marked as sensitive. This is useful for maintaining data security and confidentiality.


```sql+postgres
select
  name,
  description,
  path
from
  terraform_output
where
  sensitive;
```

```sql+sqlite
select
  name,
  description,
  path
from
  terraform_output
where
  sensitive = 1;
```

### List outputs referring to AWS S3 bucket ARN attributes
Explore which Terraform outputs reference AWS S3 bucket ARN attributes. This can be useful for identifying dependencies or potential configuration issues.Analyze the settings to understand the connections between your Terraform outputs and AWS S3 bucket ARN attributes. This is useful to identify potential dependencies or configurations that may impact your S3 bucket usage.


```sql+postgres
select
  name,
  description,
  value,
  path
from
  terraform_output
where
  value::text like '%aws_s3_bucket.%.arn%';
```

```sql+sqlite
select
  name,
  description,
  value,
  path
from
  terraform_output
where
  value like '%aws_s3_bucket.%.arn%';
```