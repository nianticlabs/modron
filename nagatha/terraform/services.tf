resource "google_project_service" "appengine_service" {
  # Required by cloud scheduler
  service = "appengine.googleapis.com"
}
resource "google_project_service" "cloud_resource_manager_service" {
  service = "cloudresourcemanager.googleapis.com"
}
resource "google_project_service" "cloudbuild_service" {
  service = "cloudbuild.googleapis.com"
}
resource "google_project_service" "cloudscheduler_service" {
  service = "cloudscheduler.googleapis.com"
}
resource "google_project_service" "compute_service" {
  service = "compute.googleapis.com"
}
resource "google_project_service" "iam_service" {
  service = "iam.googleapis.com"
}
resource "google_project_service" "stackdriver_service" {
  service = "stackdriver.googleapis.com"
}
resource "google_project_service" "iap_service" {
  service = "iap.googleapis.com"
}
resource "google_project_service" "bigquery_service" {
  service = "bigquery.googleapis.com"
}
resource "google_project_service" "run_service" {
  service = "run.googleapis.com"
}
resource "google_project_service" "secretmanager_service" {
  service = "secretmanager.googleapis.com"
}
