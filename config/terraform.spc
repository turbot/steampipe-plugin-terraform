connection "terraform" {
  plugin = "terraform"

  # Paths is a list of locations to search for Terraform configuration files.
  # Wildcards are supported per https://golang.org/pkg/path/filepath/#Match
  # Exact file paths can have any name. Wildcard based matches must have an
  # extension of .csv (case insensitive).
  # paths = [ "/path/to/dir/*", "/path/to/exact/custom.tf" ]
}
