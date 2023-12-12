---
title: "Steampipe Table: terraform_local - Query Terraform Local Values using SQL"
description: "Allows users to query Terraform Local Values, particularly the final local value name and its corresponding expression."
---

# Table: terraform_local - Query Terraform Local Values using SQL

Terraform Local Values are a convenient naming mechanism that allows users to assign a name to an expression so it can be used multiple times within a module without repeating it. Local Values can be helpful to avoid repeating the same values or expressions multiple times in a Terraform configuration. If overused, they can also make configuration hard to read and understand if the reader has to continually lookup the values.

## Table Usage Guide

The `terraform_local` table provides insights into local values within Terraform. As a DevOps engineer, explore local value-specific details through this table, including the final local value name and its corresponding expression. Utilize it to uncover information about local values, such as those that are used multiple times, to avoid repetition and enhance the readability of the Terraform configuration.

## Examples

### Basic info
Analyze the settings to understand the basic information stored in your Terraform local configurations. This can assist in identifying potential configuration issues or inconsistencies.This query allows you to gain insights into the local values defined in your Terraform code. It can be useful to understand your configuration and to identify specific settings or paths that may need to be modified or reviewed.


```sql+postgres
select
  name,
  value,
  path
from
  terraform_local;
```

```sql+sqlite
select
  name,
  value,
  path
from
  terraform_local;
```

### List 'Owner' locals (case insensitive)
Analyze the settings to understand who the owners are across various paths in a case-insensitive manner. This can be beneficial in managing access rights and maintaining security protocols.Identify instances where the 'owner' locals are used in the Terraform configuration to understand the ownership details in the system. This can be particularly useful in managing and organizing resources effectively.


```sql+postgres
select
  name,
  value,
  path
from
  terraform_local
where
  name ilike 'owner';
```

```sql+sqlite
select
  name,
  value,
  path
from
  terraform_local
where
  name like 'owner';
```