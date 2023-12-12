---
title: "Steampipe Table: terraform_provider - Query Terraform Providers using SQL"
description: "Allows users to query Terraform Providers, specifically the details about the provider plugins in the Terraform state."
---

# Table: terraform_provider - Query Terraform Providers using SQL

Terraform Providers are plugins that Terraform uses to manage resources. A provider is responsible for understanding API interactions and exposing resources. Providers generally are an IaaS (e.g., Alibaba Cloud, AWS, GCP, Microsoft Azure, OpenStack), PaaS (e.g., Heroku), or SaaS services (e.g., Terraform Cloud, DNSimple, Cloudflare).

## Table Usage Guide

The `terraform_provider` table provides insights into provider plugins in the Terraform state. As a DevOps engineer, explore provider-specific details through this table, including the provider type, version, and associated metadata. Utilize it to uncover information about providers, such as their versions, and the verification of provider configurations.

## Examples

### Basic info
Explore the specifics of your Terraform provider, including its name, alias, and associated arguments. This can help you better understand the configuration and structure of your Terraform environment.Explore the basic information about your Terraform providers to understand their names, aliases, and arguments, and to locate their paths. This can help streamline your configuration management and troubleshooting processes.

```sql+postgres
select
  name,
  alias,
  arguments,
  path
from
  terraform_provider;
```

```sql+sqlite
select
  name,
  alias,
  arguments,
  path
from
  terraform_provider;
```

### List providers using deprecated 'version' argument
This example helps you identify the instances where deprecated 'version' arguments are still being used in your Terraform providers. This can aid in ensuring your configuration is up-to-date and compliant with current best practices.Discover the segments that are utilizing outdated 'version' arguments in their configuration. This aids in identifying areas for potential updates and improvements.

```sql+postgres
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

```sql+sqlite
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
Explore the configuration of your AWS providers to understand the specific regions in which they operate. This can be beneficial in managing and optimizing your resource allocation across different geographical locations.Explore which AWS providers are configured across different regions. This can assist in managing resources and ensuring efficient distribution across various geographical locations.


```sql+postgres
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

```sql+sqlite
select
  name,
  alias,
  json_extract(arguments, '$.region') as region,
  path
from
  terraform_provider
where
  name = 'aws';
```