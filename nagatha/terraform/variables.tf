variable "project" {
  description = "Project ID to deploy this infrastructure to."
  validation {
    error_message = "The project ID must be a unique string of 6 to 30 lowercase letters, digits, or hyphens. It must start with a letter, and cannot have a trailing hyphen."
    condition     = length(var.project) > 5 && length(var.project) < 31 && regex("[a-z][a-z0-9-]*[^-]", var.project) == var.project
  }
}

variable "org_id" {
  description = "Organization ID to deploy this infrastructure to."
  default     = "0123456789"
  validation {
    error_message = "The org ID must be a number. It must not start with a 0 (match the regex '[1-9][0-9]*')."
    condition     = regex("[1-9][0-9]*", var.org_id) == var.org_id
  }
}

variable "region" {
  description = "Region to deploy all of this in. https://cloud.google.com/compute/docs/regions-zones"
}

variable "project_admins" {
  description = "People that can impersonate the terraform account and manage the project."
  type        = list(string)
  validation {
    error_message = "The project_admins variable must be a list of emails (@your_domain.com)"
    condition     = length(var.project_admins) > 0 && can([for o in var.project_admins : regex("(.*@.*)", o)])
  }
}

variable "bq_owner" {
  description = "Owner of the BigQuery backend"
  type        = list(string)
  validation {
    error_message = "The bq_owner variable must contain an @ (@your_domain.com or @project.iam.gserviceaccount.com)"
    condition     = length(var.bq_owner) > 0 && can([for o in var.bq_owner : regex("(.*@.*)", o)])
  }
}

variable "env" {
  description = "Environment to be deployed to"
}


variable "domain" {
  description = "Domain of nagatha."
}

variable "email_sender_address" {
  description = "Email sending the emails for Nagatha."
  validation {
    error_message = "The email_sender_address variable must contain an @ (@your_domain.com)"
    condition     = length(var.email_sender_address) > 0 && can(regex("(.*@.*)", var.email_sender_address))
  }
}
