resource "google_storage_bucket" "otel_config" {
  name                        = "${var.project}-otel-config"
  location                    = local.region
  uniform_bucket_level_access = true
}

data "google_iam_policy" "otel_config" {
  binding {
    role = "roles/storage.objectViewer"
    members = [
      "serviceAccount:${google_service_account.modron_runner.email}",
    ]
  }

  binding {
    role = "roles/storage.admin"
    members = concat(
      var.project_admins,
      [
        "serviceAccount:${data.google_service_account.terraform_sa.email}",
      ]
    )
  }
}

resource "google_storage_bucket_iam_policy" "otel_config" {
  bucket      = google_storage_bucket.otel_config.name
  policy_data = data.google_iam_policy.otel_config.policy_data
}

resource "google_storage_bucket_object" "otel_config" {
  bucket  = google_storage_bucket.otel_config.name
  name    = "config.yaml"
  content = file("${path.module}/otel/config.yaml")
}