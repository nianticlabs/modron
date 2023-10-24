resource "google_compute_global_address" "private_ip_address" {
  name          = "modron-${var.env}-db-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.cloud_run_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.cloud_run_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "google_sql_database_instance" "instance" {
  name             = "modron-${var.env}-${random_id.db_name_suffix.hex}"
  database_version = "POSTGRES_14"

  depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier              = "db-custom-8-30720"
    availability_type = "ZONAL"
    ip_configuration {
      # Set this to true if you need to connect to the database via the Cloud SQL Proxy.
      ipv4_enabled    = false
      private_network = google_compute_network.cloud_run_network.id
      require_ssl     = true

      # No need to add private IPs in that list. Modron connects via a private IP.
      # This is for administration purposes only.
      dynamic "authorized_networks" {
        for_each = module.vpn.cidr_blocks
        iterator = cidr

        content {
          name  = cidr.value.display_name
          value = cidr.value.cidr_block
        }
      }
    }
    maintenance_window {
      day  = 7
      hour = 1
    }
    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }
    database_flags {
      name = "max_connections"
      // 100 is the maximum we can do from cloud run.
      value = "100"
    }
    database_flags {
      name  = "log_temp_files"
      value = "0"
    }
    backup_configuration {
      enabled  = true
      location = "us"
    }
    insights_config {
      query_insights_enabled  = true
      query_plans_per_minute  = 5
      query_string_length     = var.env == "dev" ? 4000 : 1024
      record_application_tags = false
      record_client_address   = false
    }
  }

  deletion_protection = "true"
}

resource "google_sql_database" "modron_database" {
  name     = "modron${var.env}"
  instance = google_sql_database_instance.instance.name
}

resource "google_sql_user" "iam_user" {
  name     = "modron${var.env}runner"
  instance = google_sql_database_instance.instance.name
  # TODO: Move to cloud IAM (soon)
  # https://github.com/GoogleCloudPlatform/cloud-sql-proxy#-enable_iam_login
  type     = "BUILT_IN"
  password = random_password.sql_user_password.result
}

resource "google_project_iam_member" "sql_client_iam" {
  project = var.project
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.modron_runner.email}"
}

resource "random_password" "sql_user_password" {
  length           = 16
  special          = true
  min_lower        = 2
  min_numeric      = 2
  min_special      = 2
  min_upper        = 2
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "google_compute_instance" "jump_host_sql" {
  name         = "jump-host-sql"
  machine_type = "e2-standard-2"
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
    }
  }

  network_interface {
    network = google_compute_network.cloud_run_network.id
  }

  metadata_startup_script = "sudo apt -y install postgresql-client-14 && gcloud -q components install cloud_sql_proxy"

  service_account {
    # Google recommends custom service accounts that have cloud-platform scope and permissions granted via IAM Roles.
    email  = google_service_account.jump_host_runner.email
    scopes = ["cloud-platform"]
  }

  shielded_instance_config {
    enable_secure_boot = true
  }
}

data "google_iam_policy" "jump_host_accessors" {
  binding {
    role    = "roles/compute.instanceAdmin.v1"
    members = var.modron_admins
  }
}

resource "google_compute_instance_iam_policy" "jump_host_policy" {
  instance_name = google_compute_instance.jump_host_sql.id
  policy_data   = data.google_iam_policy.jump_host_accessors.policy_data
}
