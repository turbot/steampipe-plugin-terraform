connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files
  # Paths can be configured with a local directory, a remote Git repository URL, or an S3 bucket URL
  # Wildcard based searches are supported, including recursive searches
  # All paths are resolved relative to the current working directory (CWD)
  # Defaults to CWD
  paths = [ "*.tf" ]
}
