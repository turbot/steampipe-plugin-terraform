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
  # All paths are resolved relative to the current working directory (CWD)
  # Defaults to CWD
  paths = [ "*.tf" ]
}
```

### Setting up paths

The argument `paths` in the config is flexible in searching Terraform configuration files from several different sources (i.e. directory paths, Git, S3, etc.). This removes the burden of knowing how to download from various sources from the implementer.

Paths support the following protocols:

- [Local files](#configuring-local-file-paths)
- [Remote Git repositories](#configuring-remote-git-repositories-urls)
- [S3](#configuring-s3-urls)

Paths may [include wildcards](https://pkg.go.dev/path/filepath#Match) and support `**` for recursive matching. Defaults to the current working directory. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "*.tf",
    "~/*.tf",
    "github.com/turbot/steampipe-plugin-aws//aws-test/tests/aws_acm_certificate//*.tf",
    "github.com/hashicorp/terraform-guides//infrastructure-as-code//**/*.tf",
    "bitbucket.org/benturrell/terraform-arcgis-portal//modules/shared//*.tf",
    "gitlab.com/YourProject/YourRepository//YourFolder//*.tf",
    "s3::https://bucket.s3.us-east-1.amazonaws.com/test_folder//*.tf"
  ]
}
```

#### Configuring local file paths

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

**Note:** If paths include `*`, all files (including non-Terraform configuration files) in the CWD will be matched, which may cause errors if incompatible file types exist.

#### Configuring remote Git repositories URLs

Not only local files, but you can also configure `paths` with any remote repository URLs (i.e. GitHub, BitBucket, GitLab) to search various Terraform configuration files in it.

You can also mention [wildcards](https://pkg.go.dev/path/filepath#Match) and support `**` for recursive matching. For example:

- `github.com/turbot/steampipe-plugin-aws//*.tf` matches all top-level Terraform configuration files in the specified repository.
- `github.com/turbot/steampipe-plugin-aws//**/*tf` matches all Terraform configuration files in the specified repository and all sub-directories.
- `github.com/turbot/steampipe-plugin-aws//**/*tf?ref=fix_7677` matches all Terraform configuration files in the specific tag of a repository.

You can specify a subdirectory after a double-slash (`//`) if you want to download only a specific subdirectory from a downloaded directory.

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [ "github.com/turbot/steampipe-plugin-aws//aws-test/tests/aws_acm_certificate//*.tf" ]
}
```

Similarly, you can define a list of GitLab and BitBucket URLs to search for the Terraform configuration files.

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "github.com/turbot/steampipe-plugin-aws//**/*tf",
    "github.com/hashicorp/terraform-guides//infrastructure-as-code//**/*.tf",
    "bitbucket.org/benturrell/terraform-arcgis-portal//modules/shared//*.tf",
    "bitbucket.org/benturrell/terraform-arcgis-portal//modules//**/*.tf",
    "gitlab.com/gitlab-org/configure/examples/gitlab-terraform-aws//*.tf",
    "gitlab.com/gitlab-org/configure/examples/gitlab-terraform-aws//**/*.tf"
  ]
}
```

#### Configuring S3 URLs

As a part of reading files from remote, you can also query all Terraform configuration files stored inside an S3 bucket (public or private) using the URL. For example:

- **Accessing a private bucket**

  In order to access your files in a private S3 bucket, you will need to configure your credentials. S3 takes various access configurations in the URL. For example:

  - `aws_access_key_id` - AWS access key.
  - `aws_access_key_secret` - AWS access key secret.
  - `aws_access_token` - AWS access token if this is being used.
  - `aws_profile` - Use this profile from local `~/.aws/config`. Takes priority over the other three. Also, make sure that `region` is configured in the config. If not set in the config, `region` will be fetched from the standard environment variable (i.e. `AWS_REGION`).

  **Note:** You can also set the above configuration using the standard AWS environment variables. If set, it will take configurations from there.

  You can authenticate your request by using the above credential option in the URL path. For example:

  ```hcl
  connection "terraform" {
    plugin = "terraform"

    paths = [
      "s3::https://bucket-2.s3.us-east-1.amazonaws.com//*.tf?aws_profile=<AWS_PROFILE>",
      "s3::https://bucket-2.s3.us-east-1.amazonaws.com/test_folder//*.tf?aws_profile=<AWS_PROFILE>",
      "s3::https://bucket-1.s3.us-east-1.amazonaws.com/test_folder//*.tf?aws_access_key_id=<AWS_ACCESS_KEY>&aws_access_key_secret=<AWS_ACCESS_KEY_SECRET>",
      "s3::https://bucket-3.s3.us-east-1.amazonaws.com/test_folder//*.tf?aws_access_key_id=<AWS_TEMPORARY_ACCESS_KEY>&aws_access_key_secret=<AWS_TEMPORARY_ACCESS_KEY_SECRET>&aws_access_token=<AWS_SESSION_TOKEN>"
    ]
  }
  ```

  **Note:**

  By default, users from the same account where the bucket is created can access the bucket objects if the user has basic S3 read access. If you are not the bucket owner and can't access a bucket, ask the owner to update the bucket policy with the required access. You will need basic read access for the bucket and object to run the query. You can refer to the sample bucket policy mentioned below:

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

  Also, ensure the credentials/profile passed in the URL are valid. Otherwise, the query will fail with an error code `403`.

- **Accessing a public bucket**

  Public access granted to buckets and objects allows all the users inside/outside the organization to access the bucket and objects stored inside it. You can give the public access to buckets and objects through access control lists (ACLs), bucket policies, or both. To identify a public bucket, you can use the below AWS CLI command:

  ```bash
  aws s3api get-bucket-policy-status --bucket <BUCKET_NAME>
  ```

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
