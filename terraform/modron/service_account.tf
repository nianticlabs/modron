resource "google_service_account" "modron_runner" {
  account_id   = "modron-${var.env}-runner"
  description  = "Modron ${var.env} runner"
  display_name = "modron-${var.env}-runner"
}

locals {
  service_account_sa_users = [for v in compact([
    google_service_account.deployer_SA.email,
    var.gitlab_impersonator_service_account,
  ]) : "serviceAccount:${v}"]
}

resource "google_service_account_iam_binding" "modron_runner_user" {
  service_account_id = google_service_account.modron_runner.name
  role               = "roles/iam.serviceAccountUser"
  members            = concat(local.service_account_sa_users, var.project_admins)
}

resource "google_project_iam_member" "runner_log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.modron_runner.email}"
}

resource "google_project_iam_member" "project_monitoring" {
  project = var.project
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.modron_runner.email}"
}

resource "google_project_iam_member" "sql_client_iam" {
  project = var.project
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.modron_runner.email}"
}
############

resource "google_service_account" "jump_host_runner" {
  account_id   = "modron-${var.env}-sql-jumphost"
  display_name = "modron-${var.env}-sql-jumphost"
}

resource "google_service_account_iam_binding" "jump_host_runner_user" {
  service_account_id = google_service_account.jump_host_runner.name
  role               = "roles/iam.serviceAccountUser"
  members            = var.project_admins
}

resource "google_project_iam_member" "jump_host_log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.jump_host_runner.email}"
}
