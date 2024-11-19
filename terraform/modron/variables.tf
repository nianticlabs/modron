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
    condition     = regex("^@[^@]*$", var.org_suffix) == var.org_suffix
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

variable "notification_system" {
  description = "Notification system to use for modron."
  type        = string
  validation {
    condition     = length(var.notification_system) > 0
    error_message = "The notification_system URL is required"
  }
}

variable "notification_system_client_id" {
  description = "Notification system client id."
  type        = string
  validation {
    condition     = length(var.notification_system_client_id) > 0
    error_message = "The notification_system_client_id is required"
  }
}

variable "gitlab_impersonator_service_account" {
  description = "The service account email that will impersonate the GitLab service account"
  type        = string
  default     = ""
  validation {
    error_message = "This must be a valid GCP service account email."
    condition     = can(regex(".*@.*\\.iam\\.gserviceaccount\\.com", var.gitlab_impersonator_service_account)) || length(var.gitlab_impersonator_service_account) == 0
  }
}

variable "docker_registry" {
  description = "Docker registry to use for the public images"
  validation {
    error_message = "The docker registry must be a valid URL"
    condition     = length(var.docker_registry) > 0
  }
}

variable "impact_map" {
  type        = map(string)
  description = "A map of environments to impact (e.g: prod -> IMPACT_HIGH) that will be used for calculating the risk score"
  default = {
    "prod"    = "IMPACT_HIGH"
    "staging" = "IMPACT_MEDIUM"
    "dev"     = "IMPACT_LOW"
  }
}

variable "additional_admin_roles" {
  type        = list(string)
  default     = []
  description = "A list of additional roles that are considered admin in GCP, for example ['organizations/11111/roles/MyOrgOwner']"
}

variable "label_to_email_regexp" {
  description = "Regexp to be used to convert the label contents of contact1,contact2 to an email"
  type        = string
  default     = "(.*)_(.*?)_(.*?)$"
}

variable "label_to_email_substitution" {
  description = "Substitution to be used to convert the label contents of contact1,contact2 to an email"
  type        = string
  default     = "$1@$2.$3"
}

variable "allowed_scc_categories" {
  description = "List of allowed Security Command Center categories to create observations.\nThese categories are the Finding.category (https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings#Finding), often refered to as \"API equivalent\" in the GCP console.\nFor example, if you want to allow the category \"GKE_RUNTIME_OS_VULNERABILITY\" you should add it to this list.\n\nA list with some possible findings can be found on https://cloud.google.com/chronicle/docs/ingestion/default-parsers/collect-security-command-center-findings."
  type        = list(string)
  default     = []
}

variable "rule_configs" {
  description = "A JSON map of the rules to their configuration"
  validation {
    error_message = "The rule_configs must be a valid JSON map"
    condition     = can(jsondecode(var.rule_configs))
  }
}
