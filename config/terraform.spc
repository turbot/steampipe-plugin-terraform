connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files
  # All paths are resolved relative to the current working directory (CWD)
  # Wildcard based searches are supported, including recursive searches

  # For example:
  #  - "*.tf" matches all Terraform configuration files in the CWD
  #  - "**/*.tf" matches all Terraform configuration files in the CWD and all sub-directories
  #  - "../*.tf" matches all Terraform configuration files in the CWD's parent directory
  #  - "steampipe*.tf" matches all Terraform configuration files starting with "steampipe" in the CWD
  #  - "/path/to/dir/*.tf" matches all Terraform configuration files in a specific directory
  #  - "/path/to/dir/main.tf" matches a specific file

    # current folder only
    #   ".",
    # current folder recursive
    #   "./**",
    # current folder recursive, implicit
    #   "**",
    # home dir only, tf files
    #   "~/*.tf",
    # home recursive tf files
    #   "~/**/*.tf",
    # TODO file::examples
    # NOTE: file@::~/ will not work
    # git hub repo, all top level files
    #   "github.com/turbot/polygoat",
    # git hub repo, recursive tf files
    #   "github.com/turbot/polygoat//**/*tf",
    # specific tag of git repo, recursive tf files
    #   "github.com/turbot/polygoat?ref=fix_7677//**/*.tf",
    # git repo, depth 1, recursive tf files
    #   "github.com/turbot/polygoat?depth=1//**/*.tf",
    # specific folder of a github repo, tf files
    #   "github.com/turbot/polygoat//testing_frameworks/steampipe_mod_benchmark//*.tf",
    # force git protocol, list top level tf files
    #   "git::github.com/turbot/polygoat//*.tf",
    # s3 bucket, tf files
    #   "s3::https://s3.amazonaws.com/bucket/terraform_examples//**/*.tf",

  # If paths includes "*", all files (including non-Terraform configuration files) in
  # the CWD will be matched, which may cause errors if incompatible file types exist

  # Defaults to CWD


        },
}

2 arrays? paths, remote_paths
can glob library return root? -> check for dir existence