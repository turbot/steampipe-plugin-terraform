---
organization: Turbot
category: ["software development"]
icon_url: "/images/plugins/turbot/terraform.svg"
brand_color: "#844FBA"
display_name: "Terraform"
short_name: "terraform"
description: "Steampipe plugin to query data from Terraform files."
og_description: "Query Terraform files with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/turbot/terraform-social-graphic.png"
---

# Terraform + Steampipe

A Terraform configuration file is used to declare resources, variables, modules, and more.

[Steampipe](https://steampipe.io) is an open source CLI to instantly query data using SQL.

Query all resources in your Terraform files:

```sql
select
  name,
  type,
  jsonb_pretty(arguments) as args
from
  terraform_resource;
```

```
> select name, type, jsonb_pretty(arguments) as args from terraform_resource;
+------------+----------------+--------------------------------------------+
| name       | type           | args                                       |
+------------+----------------+--------------------------------------------+
| app_server | aws_instance   | {                                          |
|            |                |     "ami": "ami-830c94e3",                 |
|            |                |     "tags": {                              |
|            |                |         "Name": "ExampleAppServerInstance" |
|            |                |     },                                     |
|            |                |     "instance_type": "t2.micro"            |
|            |                | }                                          |
| app_volume | aws_ebs_volume | {                                          |
|            |                |     "size": 40,                            |
|            |                |     "tags": {                              |
|            |                |         "Name": "HelloWorld"               |
|            |                |     },                                     |
|            |                |     "availability_zone": "us-west-2a"      |
|            |                | }                                          |
| app_bucket | aws_s3_bucket  | {                                          |
|            |                |     "acl": "private",                      |
|            |                |     "tags": {                              |
|            |                |         "Name": "Test bucket",             |
|            |                |         "Environment": "Dev"               |
|            |                |     },                                     |
|            |                |     "bucket": "my-app-bucket"              |
|            |                | }                                          |
+------------+----------------+--------------------------------------------+
```

## Documentation

- **[Table definitions & examples →](/plugins/turbot/terraform/tables)**

## Get started

### Install

Download and install the latest Terraform plugin:

```bash
steampipe plugin install terraform
```

### Credentials

No credentials are required.

### Configuration

Installing the latest terraform plugin will create a config file (`~/.steampipe/config/terraform.spc`) with a single connection named `terraform`:

```hcl
connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files
  # Paths can be configured with a local directory, a remote Git repository URL, or an S3 bucket URL
  # Wildcard based searches are supported, including recursive searches
  # Local paths are resolved relative to the current working directory (CWD)

  # For example:
  #  - "*.tf" matches all Terraform configuration files in the CWD
  #  - "**/*.tf" matches all Terraform configuration files in the CWD and all sub-directories
  #  - "../*.tf" matches all Terraform configuration files in the CWD's parent directory
  #  - "steampipe*.tf" matches all Terraform configuration files starting with "steampipe" in the CWD
  #  - "/path/to/dir/*.tf" matches all Terraform configuration files in a specific directory
  #  - "/path/to/dir/main.tf" matches a specific file

  # If paths includes "*", all files (including non-Terraform configuration files) in
  # the CWD will be matched, which may cause errors if incompatible file types exist

  # Defaults to CWD
  paths = [ "*.tf" ]
}
```

### Supported Path Formats

The `paths` config argument is flexible and can search for Terraform configuration files from several different sources, e.g., local directory paths, Git, S3.

The following sources are supported:

- [Local files](#configuring-local-file-paths)
- [Remote Git repositories](#configuring-remote-git-repository-urls)
- [S3](#configuring-s3-urls)

Paths may [include wildcards](https://pkg.go.dev/path/filepath#Match) and support `**` for recursive matching. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "*.tf",
    "~/*.tf",
    "github.com/turbot/steampipe-plugin-aws//aws-test/tests/aws_acm_certificate//*.tf",
    "github.com/hashicorp/terraform-guides//infrastructure-as-code//**/*.tf",
    "bitbucket.org/benturrell/terraform-arcgis-portal//modules/shared//*.tf",
    "gitlab.com/gitlab-org/configure/examples/gitlab-terraform-aws//*.tf",
    "s3::https://bucket.s3.us-east-1.amazonaws.com/test_folder//*.tf"
  ]
}
```

**Note**: If any path matches on `*` without `.tf`, all files (including non-Terraform configuration files) in the directory will be matched, which may cause errors if incompatible file types exist.

#### Configuring Local File Paths

You can define a list of local directory paths to search for terraform files. Paths are resolved relative to the current working directory. For example:

- `*.tf` matches all Terraform configuration files in the CWD.
- `**/*.tf` matches all Terraform configuration files in the CWD and all sub-directories.
- `../*.tf` matches all Terraform configuration files in the CWD's parent directory.
- `steampipe*.tf` matches all Terraform configuration files starting with "steampipe" in the CWD.
- `/path/to/dir/*.tf` matches all Terraform configuration files in a specific directory. For example:
  - `~/*.tf` matches all Terraform configuration files in the home directory.
  - `~/**/*.tf` matches all Terraform configuration files recursively in the home directory.
- `/path/to/dir/main.tf` matches a specific file.

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [ "*.tf", "~/*.tf", "/path/to/dir/main.tf" ]
}
```

#### Configuring Remote Git Repository URLs

You can also configure `paths` with any Git remote repository URLs, e.g., GitHub, BitBucket, GitLab. The plugin will then attempt to retrieve any Terraform configuration files from the remote repositories.

For example:

- `github.com/turbot/steampipe-plugin-aws//*.tf` matches all top-level Terraform configuration files in the specified repository.
- `github.com/turbot/steampipe-plugin-aws//**/*.tf` matches all Terraform configuration files in the specified repository and all subdirectories.
- `github.com/turbot/steampipe-plugin-aws//**/*.tf?ref=fix_7677` matches all Terraform configuration files in the specific tag of a repository.
- `github.com/turbot/steampipe-plugin-aws//aws-test/tests/aws_acm_certificate//*.tf` matches all Terraform configuration files in the specified folder path.

If the example formats above do not work for private repositories, this could be due to git credentials being stored by another tool, e.g., VS Code. An alternative format you can try is:

- `git::ssh://git@github.com/test_org/test_repo//*.tf`

You can specify a subdirectory after a double-slash (`//`) if you want to download only a specific subdirectory from a downloaded directory.

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [ "github.com/turbot/steampipe-plugin-aws//aws-test/tests/aws_acm_certificate//*.tf" ]
}
```

Similarly, you can define a list of GitLab and BitBucket URLs to search for Terraform configuration files:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "github.com/turbot/steampipe-plugin-aws//**/*.tf",
    "github.com/hashicorp/terraform-guides//infrastructure-as-code//**/*.tf",
    "bitbucket.org/benturrell/terraform-arcgis-portal//modules/shared//*.tf",
    "bitbucket.org/benturrell/terraform-arcgis-portal//modules//**/*.tf",
    "gitlab.com/gitlab-org/configure/examples/gitlab-terraform-aws//*.tf",
    "gitlab.com/gitlab-org/configure/examples/gitlab-terraform-aws//**/*.tf"
  ]
}
```

#### Configuring S3 URLs

You can also query all Terraform configuration files stored inside an S3 bucket (public or private) using the bucket URL.

##### Accessing a Private Bucket

In order to access your files in a private S3 bucket, you will need to configure your credentials. You can use your configured AWS profile from local `~/.aws/config`, or pass the credentials using the standard AWS environment variables, e.g., `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`.

We recommend using AWS profiles for authentication.

**Note:** Make sure that `region` is configured in the config. If not set in the config, `region` will be fetched from the standard environment variable `AWS_REGION`.

You can also authenticate your request by setting the AWS profile and region in `paths`. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "s3::https://bucket-2.s3.us-east-1.amazonaws.com//*.tf?aws_profile=<AWS_PROFILE>",
    "s3::https://bucket-2.s3.us-east-1.amazonaws.com/test_folder//*.tf?aws_profile=<AWS_PROFILE>"
  ]
}
```

**Note:**

In order to access the bucket, the IAM user or role will require the following IAM permissions:

- `s3:ListBucket`
- `s3:GetObject`
- `s3:GetObjectVersion`

If the bucket is in another AWS account, the bucket policy will need to grant access to your user or role. For example:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadBucketObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::123456789012:user/YOUR_USER"
      },
      "Action": [
        "s3:ListBucket",
        "s3:GetObject",
        "s3:GetObjectVersion"
      ],
      "Resource": [
        "arn:aws:s3:::test-bucket1",
        "arn:aws:s3:::test-bucket1/*"
      ]
    }
  ]
}
```

##### Accessing a Public Bucket

Public access granted to buckets and objects through ACLs and bucket policies allows any user access to data in the bucket. We do not recommend making S3 buckets public, but if there are specific objects you'd like to make public, please see [How can I grant public read access to some objects in my Amazon S3 bucket?](https://aws.amazon.com/premiumsupport/knowledge-center/read-access-objects-s3-bucket/).

You can query any public S3 bucket directly using the URL without passing credentials. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "s3::https://bucket-1.s3.us-east-1.amazonaws.com/test_folder//*.tf",
    "s3::https://bucket-2.s3.us-east-1.amazonaws.com/test_folder//**/*.tf"
  ]
}
```

## Get involved

- Open source: https://github.com/turbot/steampipe-plugin-terraform
- Community: [Slack Channel](https://steampipe.io/community/join)
