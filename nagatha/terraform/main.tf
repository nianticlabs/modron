provider "google" {
  project      = var.project
  region       = var.region
  access_token = data.google_service_account_access_token.sa.access_token
}

resource "google_compute_ssl_policy" "modern_TLS_policy" {
  min_tls_version = "TLS_1_2"
  name            = "modern-ssl-policy"
  profile         = "MODERN"
}

# Store the state in a bucket. The bucket must already exist.
terraform {
  backend "gcs" {
    bucket                      = "nagatha-tfstate"
    prefix                      = "terraform/state"
    impersonate_service_account = "terraform-sa@<your-project>.iam.gserviceaccount.com"
  }
}
