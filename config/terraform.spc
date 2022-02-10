connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files
  # All paths are resolved relative to the current working directory (CWD)
  # Wildcard based searches are supported, including recursive searches

  # For example:
  #  - "*.tf" matches all Terraform configuration files in the CWD
  #  - "**/*.tf" matches all Terraform configuration files in the CWD and all sub-directories
  #  - "../*.tf" matches all Terraform configuration files in the CWD's parent directory
  #  - "steampipe*.tf" matches all Terraform configuration files starting with "steampipe" in the current CWD
  #  - "/path/to/dir/*.tf" matches all Terraform configuration files in a specific directory
  #  - "/path/to/dir/main.tf" matches a specific file

  # If paths includes "*", all files (including non-Terraform configuration files) in
  # the current CWD will be matched, which may cause errors if incompatible filetypes exist

  # Defaults to CWD
  paths = [ "*.tf" ]
}
