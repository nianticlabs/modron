resource "google_project_iam_binding" "cloud_sql_admins" {
  project = var.project
  role    = "roles/cloudsql.admin"
  members = var.project_admins
}

resource "google_project_iam_binding" "cloud_trace_agent" {
  project = var.project
  role    = "roles/cloudtrace.agent"
  members = [
    "serviceAccount:${google_service_account.modron_runner.email}",
  ]
}

resource "google_project_iam_binding" "monitoring_writer" {
  project = var.project
  role    = "roles/monitoring.metricWriter"
  members = [
    "serviceAccount:${google_service_account.modron_runner.email}",
  ]
}