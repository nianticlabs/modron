resource "google_compute_network" "cloud_run_network" {
  auto_create_subnetworks = true
  mtu                     = 1460
  name                    = "cloud-run-network"
  routing_mode            = "REGIONAL"
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_region_network_endpoint_group" "neg" {
  name                  = "nagatha-endpoint"
  network_endpoint_type = "SERVERLESS"
  region                = "us-central1"
  cloud_run {
    service = google_cloud_run_service.nagatha.name
  }
}
