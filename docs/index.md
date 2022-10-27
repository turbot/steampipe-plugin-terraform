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
  # All paths are resolved relative to the current working directory (CWD)
  # Wildcard based searches are supported, including recursive searches
  # Defaults to CWD
  paths = [ "*.tf" ]
}
```

### Setting up paths

The argument `paths` in the config is a list of directory paths, GitHub URLs, GilLab URLs, BitBucket URLs or a S3 URL to search for Terraform files. Paths may [include wildcards](https://pkg.go.dev/path/filepath#Match) and also support `**` for recursive matching. Defaults to the current working directory. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "*.tf",
    "~/*.tf",
    "github.com/turbot/polygoat//*.tf",
    "github.com/turbot/polygoat//testing_frameworks/steampipe_mod_benchmark//*.tf",
    "git::https://github.com/turbot/steampipe-plugin-alicloud.git//alicloud-test/tests/alicloud_account//*.tf",
    "bitbucket.org/YourTeamOrUser/YourRepository//YourFolder//*.tf",
    "gitlab.com/YourProject/YourRepository//YourFolder//*.tf",
    "s3::https://bucket.s3.ap-southeast-1.amazonaws.com/terraform_examples//**/*.tf"
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

**NOTE:** If paths includes `*`, all files (including non-Terraform configuration files) in the CWD will be matched, which may cause errors if incompatible file types exist.

#### Configuring GitHub/GitLab/BitBucket URLs

You can define a list of GitHub URL as input to search for terraform files from a variety of protocols. For example:

- `github.com/turbot/polygoat//*.tf` matches all top-level Terraform configuration files in the specified github repository.
- `github.com/turbot/polygoat//**/*tf` matches all Terraform configuration files in the specified github repository and all sub-directories.
- `github.com/turbot/polygoat?ref=fix_7677//**/*tf` matches all Terraform configuration files in the specific tag of github repository.
- `git::https://github.com/turbot/steampipe-plugin-alicloud.git//alicloud-test/tests/alicloud_account//*.tf` matches all Terraform configuration files in the given HTTP URL using the Git protocol.

If you want to download only a specific subdirectory from a downloaded directory, you can specify a subdirectory after a double-slash (`//`).

- `github.com/turbot/polygoat//testing_frameworks/steampipe_mod_benchmark//*.tf` matches all Terraform configuration files in the specific folder of a github repo.

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "github.com/turbot/polygoat//*.tf",
    "github.com/turbot/polygoat//testing_frameworks/steampipe_mod_benchmark//*.tf",
    "git::https://github.com/turbot/steampipe-plugin-alicloud.git//alicloud-test/tests/alicloud_account//*.tf"
  ]
}
```

Similarly, you can also define a list of GitLab and BitBucket URLs to search for the Terraform configuration files. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "bitbucket.org/YourTeamOrUser/YourRepository//YourFolder//*.tf",
    "bitbucket.org/YourTeamOrUser/YourRepository//**/*.tf",
    "gitlab.com/YourProject/YourRepository//YourFolder//*.tf",
    "gitlab.com/YourProject/YourRepository//**/*.tf"
  ]
}
```

#### Configuring S3 URLs

As a part of reading files from remote, you can also query all Terraform configuration files stored inside a S3 bucket (public or private) using the URL. For example:

##### Accessing a private bucket

Using this plugin, you can query the files inside a private S3 bucket. S3 takes various access configurations in the URL. For example:

- `aws_access_key_id` - AWS access key.
- `aws_access_key_secret` - AWS access key secret.
- `aws_access_token` - AWS access token if this is being used.
- `aws_profile` - Use this profile from local `~/.aws/config`. Takes priority over the other three.

You can use any of the above credential option in the URL path to authenticate your request. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "<bucket-name>.s3.amazonaws.com/<YOUR_FOLDER>?aws_access_key_id=<AWS_ACCESS_KEY>&aws_access_key_secret=<AWS_ACCESS_KEY_SECRET>&region=<region-code>//*.tf",
    "<bucket-name>.s3.amazonaws.com/<YOUR_FOLDER>?aws_profile=<AWS_PROFILE>&region=<region-code>//*.tf",
    "<bucket-name>.s3-<region-code>.amazonaws.com/<YOUR_FOLDER>?aws_profile=<AWS_PROFILE>//**/*.tf",
    "s3::<bucket-name>.s3.amazonaws.com/<YOUR_FOLDER>/<YOUR_FILE>?aws_profile=<AWS_PROFILE>&region=<region-code>"
  ]
}
```

**NOTE:** Make sure the credentials passed is valid or have proper access to list the bucket and list objects inside it, otherwise, the query will fail with an error code `403`.

##### Accessing a public bucket

You can query any public S3 bucket directly using the URL without passing any credentials. For example:

```hcl
connection "terraform" {
  plugin = "terraform"

  paths = [
    "s3::<bucket_name>.s3-<region-code>.amazonaws.com/<YOUR_FOLDER>//*.tf",
    "s3::https://<bucket_name>.s3.<region-code>.amazonaws.com/<YOUR_FOLDER>//**/*.tf",
    "s3::https://<bucket_name>.s3.<region-code>.amazonaws.com/<YOUR_FILE>"
  ]
}
```

## Get involved

- Open source: https://github.com/turbot/steampipe-plugin-terraform
- Community: [Slack Channel](https://steampipe.io/community/join)
