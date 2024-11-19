resource "google_secret_manager_secret" "sql_connect_string_config" {
  secret_id = "sql_connect_string"

  replication {
    user_managed {
      replicas {
        location = local.region
      }
    }
  }
}

resource "google_secret_manager_secret_version" "sql_connect_string" {
  secret = google_secret_manager_secret.sql_connect_string_config.id

  # We use SQL proxy in cloud run, this is fine to disable SSL here as the proxy runs on the same instance and authenticates the instance.
  secret_data = "host=/cloudsql/${google_sql_database_instance.instance.connection_name} user=${google_sql_user.iam_user.name} dbname=modron${var.env} sslmode=disable password=${random_password.sql_user_password.result}"
}

data "google_iam_policy" "secret_accessors_policy" {
  binding {
    role = "roles/secretmanager.secretAccessor"
    members = [
      "serviceAccount:${google_service_account.modron_runner.email}"
    ]
  }
}

resource "google_secret_manager_secret_iam_policy" "prod_accesses_secret" {
  secret_id   = google_secret_manager_secret.sql_connect_string_config.secret_id
  policy_data = data.google_iam_policy.secret_accessors_policy.policy_data
}
