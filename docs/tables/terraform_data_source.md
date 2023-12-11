---
title: "Steampipe Table: terraform_data_source - Query Terraform Data Sources using SQL"
description: "Allows users to query Terraform Data Sources, specifically providing insights into the configuration and state of data sources in Terraform."
---

# Table: terraform_data_source - Query Terraform Data Sources using SQL

Terraform Data Sources are a type of resource in Terraform that allow users to fetch data from a specific source or service. This data can then be used within other resources or outputs within your Terraform configuration. It is a powerful tool that allows for the dynamic configuration of resources based on data retrieved from external sources.

## Table Usage Guide

The `terraform_data_source` table provides insights into data sources within Terraform. As a DevOps engineer, explore data source-specific details through this table, including the configuration and state of each data source. Utilize it to understand how data is being fetched and used within your Terraform configuration, and to ensure that data sources are being used effectively and securely.

## Examples

### Basic info
Explore various data sources in your Terraform configuration to identify their names, types, and paths to understand the structure and organization of your infrastructure. This can be useful when you want to review or modify your configuration.This query can help you explore the different data sources in your Terraform setup. It's useful for understanding the types of data your infrastructure is relying on and where that data is coming from.

```sql+postgres
select
  name,
  type,
  arguments,
  path
from
  terraform_data_source;
```

```sql+sqlite
select
  name,
  type,
  arguments,
  path
from
  terraform_data_source;
```

### List AWS EC2 AMIs
Determine the areas in which specific AWS EC2 AMIs are used, by analyzing the data source. This can help in understanding the distribution and application of different AMIs within your infrastructure.Explore which Amazon Machine Images (AMIs) are available in your AWS EC2 environment using this query. It helps in assessing the elements within your infrastructure and aids in making informed decisions for resource allocation and management.

```sql+postgres
select
  name,
  type,
  arguments,
  path
from
  terraform_data_source
where
  type = 'aws_ami';
```

```sql+sqlite
select
  name,
  type,
  arguments,
  path
from
  terraform_data_source
where
  type = 'aws_ami';
```

### Get filters for each AWS EC2 AMI
Discover the segments that help to identify the specific filters applied to each AWS EC2 AMI. This is beneficial for understanding the configuration and management of your EC2 AMIs, aiding in resource optimization and security.Explore which AWS EC2 AMIs have specific filters applied to them. This is useful for understanding your AMI configurations and ensuring they align with your security and operational requirements.


```sql+postgres
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

```sql+sqlite
with filters as (
select
  name,
  type,
  json_each(arguments, '$.filter') as filter,
  path
from
  terraform_data_source
where
  type = 'aws_ami'
)
select
  name,
  type,
  json_extract(filter.value, '$.name') as name,
  json_extract(filter.value, '$.values') as values,
  path
from
  filters;
```