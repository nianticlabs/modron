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
  region                = local.region
  cloud_run {
    service = google_cloud_run_v2_service.grpc_web.name
  }
}

resource "google_compute_region_network_endpoint_group" "ui_neg" {
  name                  = "modron-ui-${var.env}-endpoint"
  network_endpoint_type = "SERVERLESS"
  region                = local.region
  cloud_run {
    service = google_cloud_run_v2_service.ui.name
  }
}


resource "google_vpc_access_connector" "connector" {
  name          = "cloud-run-vpc-connector"
  network       = google_compute_network.cloud_run_network.name
  ip_cidr_range = "10.42.0.0/28"
}

# This is required to install packages on the SQL jump host
resource "google_compute_router" "router" {
  name    = "sql-jump-host"
  region  = local.region
  network = google_compute_network.cloud_run_network.id

  bgp {
    asn = 64514
  }
}

resource "google_compute_router_nat" "nat" {
  name                               = "sql-jump-host"
  router                             = google_compute_router.router.name
  region                             = google_compute_router.router.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}
