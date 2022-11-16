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
