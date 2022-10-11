// Allocate global IP for the service and UI
resource "google_compute_global_address" "nagatha" {
  address_type = "EXTERNAL"
  ip_version   = "IPV4"
  name         = "nagatha"
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_backend_service" "nagatha" {
  name                  = "nagatha-backend"
  port_name             = "http"
  protocol              = "HTTP2"
  session_affinity      = "NONE"
  timeout_sec           = 30
  load_balancing_scheme = "EXTERNAL"

  log_config {
    enable      = true
    sample_rate = 1
  }
  backend {
    group = google_compute_region_network_endpoint_group.neg.self_link
  }
  iap {
    oauth2_client_id     = google_iap_client.project_client.client_id
    oauth2_client_secret = google_iap_client.project_client.secret
  }
}

resource "google_compute_global_forwarding_rule" "nagatha" {
  ip_address            = google_compute_global_address.nagatha.address
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  name                  = "nagatha"
  port_range            = "443-443"
  target                = google_compute_target_https_proxy.nagatha.id
}

resource "google_compute_target_https_proxy" "nagatha" {
  name             = "nagatha-target-proxy"
  quic_override    = "NONE"
  ssl_certificates = [google_compute_managed_ssl_certificate.nagatha.self_link]
  ssl_policy       = google_compute_ssl_policy.modern_TLS_policy.id
  url_map          = google_compute_url_map.nagatha.id
}

resource "google_compute_url_map" "nagatha" {
  name            = "nagatha"
  default_service = google_compute_backend_service.nagatha.id
}

resource "google_compute_managed_ssl_certificate" "nagatha" {
  name = "nagatha-cert"

  managed {
    domains = [var.domain]
  }

  depends_on = [
    google_project_service.compute_service
  ]
}
