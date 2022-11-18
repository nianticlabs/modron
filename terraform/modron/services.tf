resource "google_project_service" "apikeys_service" {
  service = "apikeys.googleapis.com"
}
resource "google_project_service" "bigquery_service" {
  service = "bigquery.googleapis.com"
}
resource "google_project_service" "cloudasset_service" {
  service = "cloudasset.googleapis.com"
}
resource "google_project_service" "cloudidentity_service" {
  service = "cloudidentity.googleapis.com"
}
resource "google_project_service" "cloud_resource_manager_service" {
  service = "cloudresourcemanager.googleapis.com"
}
resource "google_project_service" "cloudbuild_service" {
  service = "cloudbuild.googleapis.com"
}
resource "google_project_service" "compute_service" {
  service = "compute.googleapis.com"
}
resource "google_project_service" "container_service" {
  service = "container.googleapis.com"
}
resource "google_project_service" "iam_service" {
  service = "iam.googleapis.com"
}
resource "google_project_service" "iap_service" {
  service = "iap.googleapis.com"
}
resource "google_project_service" "run_service" {
  service = "run.googleapis.com"
}
resource "google_project_service" "serviceusage_service" {
  service = "serviceusage.googleapis.com"
}
resource "google_project_service" "stackdriver_service" {
  service = "stackdriver.googleapis.com"
}
resource "google_project_service" "spanner_service" {
  service = "spanner.googleapis.com"
}
