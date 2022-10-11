provider "google" {
  project      = var.project
  region       = substr(var.zone, 0, length(var.zone) - 2)
  access_token = data.google_service_account_access_token.sa.access_token
  zone         = var.zone
}

resource "google_compute_ssl_policy" "modern_TLS_policy" {
  min_tls_version = "TLS_1_2"
  name            = "modern-ssl-policy"
  profile         = "MODERN"
}
