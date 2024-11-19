resource "google_service_account" "deployer_SA" {
  account_id   = "gitlab-deployer"
  description  = "Used by Gitlab to deploy on Cloud Run."
  display_name = "gitlab-deployer"
}

data "google_iam_policy" "gitlab_deployer" {
  binding {
    role = "roles/iam.serviceAccountTokenCreator"
    members = [
      "serviceAccount:${var.gitlab_impersonator_service_account}",
    ]
  }
  count = var.gitlab_impersonator_service_account != "" ? 1 : 0
}

resource "google_service_account_iam_policy" "gitlab_deployer_iam_policy" {
  policy_data        = data.google_iam_policy.gitlab_deployer[0].policy_data
  service_account_id = google_service_account.deployer_SA.name
  count              = var.gitlab_impersonator_service_account != "" ? 1 : 0
}

resource "google_project_iam_member" "gitlab_cloud_build" {
  project = var.project
  role    = "roles/cloudbuild.builds.editor"
  member  = "serviceAccount:${google_service_account.deployer_SA.email}"
}

resource "google_project_iam_member" "gitlab_run_developer" {
  project = var.project
  role    = "roles/run.developer"
  member  = "serviceAccount:${google_service_account.deployer_SA.email}"
}

# This is required to build, according to Google it is compatible with the concept of least privilege -_-
# https://cloud.google.com/build/docs/securing-builds/store-manage-build-logs#viewing_build_logs
# TODO: Update to a custom log bucket and remove this permission.
resource "google_project_iam_member" "gitlab_cloud_build_storage" {
  project = var.project
  role    = "roles/viewer"
  member  = "serviceAccount:${google_service_account.deployer_SA.email}"
}
