resource "google_service_account" "modron_runner" {
  account_id   = "modron-${var.env}-runner"
  description  = "Modron ${var.env} runner"
  display_name = "modron-${var.env}-runner"
}

resource "google_service_account" "jump_host_runner" {
  account_id   = "modron-${var.env}-sql-jumphost"
  display_name = "modron-${var.env}-sql-jumphost"
}

resource "google_project_iam_member" "jump_host_log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.jump_host_runner.email}"
}
