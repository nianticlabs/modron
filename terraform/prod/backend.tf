# Store the state in a bucket. The bucket must already exist.
terraform {
  backend "gcs" {
    bucket                      = "modron-tfstate"
    prefix                      = "terraform/state"
    impersonate_service_account = "terraform-sa@<your-project>.iam.gserviceaccount.com"
  }
}
