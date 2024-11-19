resource "google_artifact_registry_repository" "registry" {
  location      = local.region
  repository_id = "modron"
  description   = "Modron Docker images"
  format        = "DOCKER"
}

# writer is not enough: GitLab needs to be able to delete tags
# otherwise the pipeline will fail with IAM_PERMISSION_DENIED when trying to replace the :dev / :prod tags
data "google_iam_policy" "modron_repository_editor_policy" {
  binding {
    role = "organizations/0123456789/roles/ArtifactRegistryDockerEditor"
    members = [
      "serviceAccount:${google_service_account.deployer_SA.email}"
    ]
  }
}

resource "google_artifact_registry_repository_iam_policy" "modron_repository_write_policy" {
  location    = google_artifact_registry_repository.registry.location
  repository  = google_artifact_registry_repository.registry.name
  policy_data = data.google_iam_policy.modron_repository_editor_policy.policy_data
}
