# Service account impersonation.
# This allows authorized users to impersonate the terraform account to apply
# terraform configs.
provider "google" {
  alias   = "tokengen"
  project = var.project
}

locals {
  terraform_sa_email = "terraform-sa@${var.project}.iam.gserviceaccount.com"
}

data "google_service_account" "terraform_sa" {
  account_id = local.terraform_sa_email
}

data "google_service_account_access_token" "sa" {
  provider               = google.tokengen
  target_service_account = local.terraform_sa_email
  # This lifetime needs to be long enough to recreate clusters which takes around 15 minutes.
  lifetime = "1200s"
  scopes = [
    "https://www.googleapis.com/auth/cloud-platform",
  ]
}

data "google_iam_policy" "terraform_iam_policy" {
  provider = google.tokengen
  binding {
    role    = "roles/iam.serviceAccountTokenCreator"
    members = var.project_admins
  }

  binding {
    role    = "roles/iam.serviceAccountAdmin"
    members = var.project_admins
  }
}

resource "google_service_account_iam_policy" "admin-account-iam" {
  service_account_id = data.google_service_account.terraform_sa.name
  policy_data        = data.google_iam_policy.terraform_iam_policy.policy_data
}
