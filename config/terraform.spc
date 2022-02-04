connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files.
  # Wildcard based searches are supported.
  # Exact file paths can have any name. Wildcard based matches must have an
  # extension of .tf (case insensitive). Default set to current working directory.
  paths = [ "./*.tf" ]
}
