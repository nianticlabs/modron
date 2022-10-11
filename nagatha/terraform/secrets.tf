resource "google_secret_manager_secret" "sendgrid_api_key_config" {
  secret_id = "sendgrid_api_key"

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

data "google_secret_manager_secret_version" "sendgrid_api_key" {
  secret  = google_secret_manager_secret.sendgrid_api_key_config.id
  version = "1"
}

data "google_iam_policy" "secret_accessors_policy" {
  binding {
    role = "roles/secretmanager.secretAccessor"
    members = [
      "serviceAccount:${google_service_account.nagatha_sa.email}",
    ]
  }
}

resource "google_secret_manager_secret_iam_policy" "prod_accesses_secret" {
  secret_id   = google_secret_manager_secret.sendgrid_api_key_config.secret_id
  policy_data = data.google_iam_policy.secret_accessors_policy.policy_data
}
