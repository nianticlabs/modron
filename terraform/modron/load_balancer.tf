resource "google_compute_global_address" "modron" {
  address_type = "EXTERNAL"
  ip_version   = "IPV4"
  name         = "modron-${var.env}"
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_backend_service" "modron_grpc_web" {
  name                  = "modron-grpc-web-${var.env}-backend"
  port_name             = "http"
  protocol              = "HTTPS"
  session_affinity      = "NONE"
  timeout_sec           = 30
  load_balancing_scheme = "EXTERNAL"

  log_config {
    enable      = true
    sample_rate = 1
  }
  backend {
    group = google_compute_region_network_endpoint_group.grpc_web_neg.self_link
  }
  iap {
    enabled              = true
    oauth2_client_id     = google_iap_client.project_client.client_id
    oauth2_client_secret = google_iap_client.project_client.secret
  }
}

resource "google_compute_backend_service" "modron_ui" {
  name                  = "modron-ui-${var.env}-backend"
  port_name             = "http"
  protocol              = "HTTPS"
  session_affinity      = "NONE"
  timeout_sec           = 30
  load_balancing_scheme = "EXTERNAL"

  log_config {
    enable      = true
    sample_rate = 1
  }
  backend {
    group = google_compute_region_network_endpoint_group.ui_neg.self_link
  }
  iap {
    enabled              = true
    oauth2_client_id     = google_iap_client.project_client.client_id
    oauth2_client_secret = google_iap_client.project_client.secret
  }
}

resource "google_compute_global_forwarding_rule" "modron" {
  name                  = "modron-${var.env}"
  ip_address            = google_compute_global_address.modron.address
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "443-443"
  target                = google_compute_target_https_proxy.modron.id
}

resource "google_compute_global_forwarding_rule" "modron_redirect" {
  name                  = "modron-${var.env}-redirect"
  ip_address            = google_compute_global_address.modron.address
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "80-80"
  target                = google_compute_target_http_proxy.modron.id
}

resource "google_compute_target_https_proxy" "modron" {
  name             = "modron-${var.env}-target-proxy"
  quic_override    = "NONE"
  ssl_certificates = [google_compute_managed_ssl_certificate.modron.self_link]
  ssl_policy       = google_compute_ssl_policy.modern_TLS_policy.id
  url_map          = google_compute_url_map.modron.id
}

resource "google_compute_target_http_proxy" "modron" {
  name    = "modron-${var.env}-target-proxy-http"
  url_map = google_compute_url_map.modron_redirect.id
}

resource "google_compute_url_map" "modron_redirect" {
  name = "modron-${var.env}-redirect"
  default_url_redirect {
    strip_query            = false
    https_redirect         = true
    redirect_response_code = "FOUND"
  }
}

resource "google_compute_url_map" "modron" {
  name            = "modron-${var.env}"
  default_service = google_compute_backend_service.modron_ui.id

  host_rule {
    hosts        = [var.domain]
    path_matcher = "modron-${var.env}-matcher"
  }

  path_matcher {
    name            = "modron-${var.env}-matcher"
    default_service = google_compute_backend_service.modron_ui.id

    route_rules {
      priority = 1
      match_rules {
        prefix_match = "/api"
        header_matches {
          header_name = "content-type"
          // Includes grpc-web, grpc-web-text, and grpc-web+proto
          prefix_match = "application/grpc-web"
        }
      }
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_service.modron_grpc_web.id
    }
  }
}

resource "google_compute_managed_ssl_certificate" "modron" {
  name = "modron-${var.env}-cert"

  managed {
    domains = [var.domain]
  }

  depends_on = [
    google_project_service.compute_service
  ]
}
