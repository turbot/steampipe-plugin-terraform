---
title: "Steampipe Table: terraform_variable - Query Terraform Variables using SQL"
description: "Allows users to query Terraform Variables, providing insights into variable definitions and their properties within Terraform configurations."
---

# Table: terraform_variable - Query Terraform Variables using SQL

Terraform Variables define the parameters that can be customized during the execution of a Terraform configuration. They allow users to provide input values for Terraform configurations, making them reusable and more flexible. Variables can have default values, descriptions, types, and validations to ensure proper usage.

## Table Usage Guide

The `terraform_variable` table provides insights into the variables defined in your Terraform configurations. As a DevOps engineer, you can explore variable-specific details through this table, including their names, types, default values, descriptions, and validation rules. Utilize it to manage and monitor your Terraform infrastructure, ensuring configurations are as expected and aiding in troubleshooting.

## Examples

### Basic Info
Explore the key details of your Terraform configuration variables. This can help you understand the values and paths associated with different elements of your configuration, and can be useful in troubleshooting or optimizing your setup.

```sql+postgres
select
  name,
  description,
  type,
  default_value,
  path
from
  terraform_variable;
```

```sql+sqlite
select
  name,
  description,
  type,
  default_value,
  path
from
  terraform_variable;
```

### List Variables with Validation Rules
Identify the variables that have validation rules applied. This is useful for ensuring that the constraints on variable values are properly understood and managed.

```sql+postgres
select
  name,
  validation,
  type
from
  terraform_variable
where
  validation is not null;
```

```sql+sqlite
select
  name,
  validation,
  type
from
  terraform_variable
where
  validation is not null;
```

### Sensitive Variables
Discover which variables in your Terraform configuration are marked as sensitive. This is useful for maintaining data security and confidentiality.

```sql+postgres
select
  name,
  description,
  sensitive
from
  terraform_variable
where
  sensitive;
```

```sql+sqlite
select
  name,
  description,
  sensitive
from
  terraform_variable
where
  sensitive = 1;
```
