resource "google_compute_network" "cloud_run_network" {
  auto_create_subnetworks = true
  mtu                     = 1460
  name                    = "cloud-run-network"
  routing_mode            = "REGIONAL"
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_region_network_endpoint_group" "grpc_web_neg" {
  name                  = "modron-grpc-web-${var.env}-endpoint"
  network_endpoint_type = "SERVERLESS"
  region                = substr(var.zone, 0, length(var.zone) - 2)
  cloud_run {
    service = google_cloud_run_service.grpc_web.name
  }
}

resource "google_compute_region_network_endpoint_group" "ui_neg" {
  name                  = "modron-ui-${var.env}-endpoint"
  network_endpoint_type = "SERVERLESS"
  region                = substr(var.zone, 0, length(var.zone) - 2)
  cloud_run {
    service = google_cloud_run_service.ui.name
  }
}
