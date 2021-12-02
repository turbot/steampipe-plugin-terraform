# Table: terraform_output

Output values are like the return values of a Terraform module, and have several uses:

- A child module can use outputs to expose a subset of its resource attributes to a parent module.
- A root module can use outputs to print certain values in the CLI output after running terraform apply.
- When using remote state, root module outputs can be accessed by other configurations via a terraform_remote_state data source.

## Examples

### Basic info

```sql
select
  name,
  description,
  value,
  path
from
  terraform_output;
```

### List sensitive outputs

```sql
select
  name,
  description,
  path
from
  terraform_output
where
  sensitive;
```

### List outputs referring to AWS S3 bucket ARN attributes

```sql
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
