variable "project" {
  description = "Project ID to deploy this infrastructure to."
  validation {
    error_message = "The project ID must be a unique string of 6 to 30 lowercase letters, digits, or hyphens. It must start with a letter, and cannot have a trailing hyphen."
    condition     = length(var.project) > 5 && length(var.project) < 31 && regex("[a-z][a-z0-9-]*[^-]", var.project) == var.project
  }
}

variable "org_id" {
  description = "Organization ID to deploy this infrastructure to."
  validation {
    error_message = "The org ID must be a number. It must not start with a 0 (match the regex '[1-9][0-9]*')."
    condition     = regex("[1-9][0-9]*", var.org_id) == var.org_id
  }
}

variable "org_suffix" {
  description = "User email suffix of your organization."
  validation {
    error_message = "Org suffix must match: '^@[^@]*$'"
    condition = regex("^@[^@]*$", var.org_suffix) == var.org_suffix
  }
}

variable "zone" {
  description = "Zone to deploy all of this in. https://cloud.google.com/compute/docs/regions-zones"
}

variable "domain" {
  description = "DNS domain of modron."
}

variable "env" {
  description = "Environment type of the database."
}

variable "dataset_id" {
  description = "(optional) Name of the dataset to be created. Will default to modron"
  type        = string
  default     = "modron"
}

variable "project_admins" {
  description = "People that can impersonate the terraform account and manage the project."
  type        = list(string)
  validation {
    error_message = "The project_admins variable must be a list of emails (@your_domain.com)"
    condition     = length(var.project_admins) > 0 && can([for o in var.project_admins : regex("(.+@.+)", o)])
  }
}

variable "modron_admins" {
  description = "User groups that are admin of modron and see all the results"
  type        = list(string)
}

variable "modron_users" {
  description = "List of group or users that will have access to the modron UI. The content will still be showed depending on the users' access inside the organisation."
  type        = list(string)
}
