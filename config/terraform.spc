connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files.
  # Wildcard based searches are supported.
  # For example:
  #  - "*" matches all files in a directory.
  #  - "**" matches all files in a directory, and all the sub-directories in it.
  #  - "./*" matches all files in current working directory.
  #  - "../*" matches all files in parent of current working directory.
  #  - "steampipe*" matches all files starting with "steampipe".
  # Exact file paths can have any name, i.e. `"/path/to/exact/custom.tf"`.
  # Default set to current working directory.
  paths = [ "./*.tf" ]
}
