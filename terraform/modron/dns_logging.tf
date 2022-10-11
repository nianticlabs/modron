resource "google_project_service" "dns_service" {
  service = "dns.googleapis.com"
}

resource "google_dns_policy" "dns-logging" {
  name = "dns-logging"

  enable_logging = true

  networks {
    network_url = google_compute_network.cloud_run_network.id
  }
  depends_on = [
    google_project_service.dns_service
  ]
}
