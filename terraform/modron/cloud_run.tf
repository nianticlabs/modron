resource "google_cloud_run_service" "grpc_web" {
  name     = "modron-grpc-web-${var.env}"
  location = substr(var.zone, 0, length(var.zone) - 2)

  # We need this to avoid naming collision with CI/CD deployments.
  autogenerate_revision_name = true

  template {
    spec {
      service_account_name = google_service_account.modron_runner.email
      timeout_seconds      = 1800
      containers {
        image = "gcr.io/${var.project}/modron:${var.env}"
        ports {
          container_port = 8080
          name           = "http1"
        }
        resources {
          limits = {
            cpu    = "4000m"
            memory = "4Gi"
          }
        }
        env {
          name  = "ADMIN_GROUPS"
          value = join(",", [for g in var.modron_admins : split(":", g)[1]])
        }
        env {
          name = "DB_MAX_CONNECTIONS"
          # Max is 100 for Cloud SQL, but we may need some connections for other purposes.
          value = 90
        }
        env {
          name  = "ENVIRONMENT"
          value = "PRODUCTION"
        }
        env {
          name  = "GCP_PROJECT_ID"
          value = var.project
        }
        env {
          name  = "GLOG_logtostderr"
          value = 1
        }
        env {
          name  = "GLOG_v"
          value = var.env == "dev" ? 10 : 1
        }
        env {
          name  = "NOTIFICATION_INTERVAL_DURATION"
          value = "720h" // 30d
        }
        env {
          name  = "NOTIFICATION_SERVICE"
          value = "https://nagatha.example.com:443"
        }
        env {
          name  = "OBSERVATION_TABLE_ID"
          value = "observations"
        }
        env {
          name  = "OPERATION_TABLE_ID"
          value = "operations"
        }
        env {
          name  = "ORG_ID"
          value = var.org_id
        }
        env {
          name  = "ORG_SUFFIX"
          value = var.org_suffix
        }
        env {
          name  = "RESOURCE_TABLE_ID"
          value = "resources"
        }
        env {
          name  = "SQL_BACKEND_DRIVER"
          value = "postgres"
        }
        env {
          name = "SQL_CONNECT_STRING"
          value_from {
            secret_key_ref {
              key  = "latest"
              name = split("/", resource.google_secret_manager_secret.sql_connect_string_config.name)[3]
            }
          }
        }
        env {
          name  = "STORAGE"
          value = "SQL"
        }
      }
    }
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale"        = "1"
        "autoscaling.knative.dev/minScale"        = "1"
        "client.knative.dev/user-image"           = "gcr.io/${var.project}/modron:${var.env}"
        "run.googleapis.com/cloudsql-instances"   = google_sql_database_instance.instance.connection_name
        "run.googleapis.com/cpu-throttling"       = "false"
        "run.googleapis.com/vpc-access-connector" = google_vpc_access_connector.connector.id
        "run.googleapis.com/vpc-access-egress"    = "private-ranges-only"
      }
      labels = {
        "run.googleapis.com/startupProbeType" = "Default"
      }
    }
  }

  metadata {
    annotations = {
      "client.knative.dev/user-image" = "gcr.io/${var.project}/modron:${var.env}"
      "run.googleapis.com/ingress"    = "internal-and-cloud-load-balancing"
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
  lifecycle {
    ignore_changes = [
      metadata[0].annotations["run.googleapis.com/operation-id"],
      metadata[0].annotations["run.googleapis.com/client-name"],
      metadata[0].annotations["run.googleapis.com/client-version"],
      template[0].metadata[0].annotations["run.googleapis.com/operation-id"],
      template[0].metadata[0].annotations["run.googleapis.com/client-name"],
      template[0].metadata[0].annotations["run.googleapis.com/client-version"],
    ]
  }
  depends_on = [
    google_project_service.run_service
  ]
}

resource "google_cloud_run_service" "ui" {
  name     = "modron-ui"
  location = substr(var.zone, 0, length(var.zone) - 2)

  # We need this to avoid naming collision with CI/CD deployments.
  autogenerate_revision_name = true

  template {
    spec {
      service_account_name = google_service_account.modron_runner.email
      timeout_seconds      = 300
      containers {
        image = "gcr.io/${var.project}/modron-ui:${var.env}"
        ports {
          container_port = 8080
          name           = "http1"
        }
        resources {
          limits = {
            cpu    = "4000m"
            memory = "4Gi"
          }
        }
        env {
          name  = "DIST_PATH"
          value = "./ui"
        }
      }
    }
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "1"
        "autoscaling.knative.dev/minScale" = "1"
        "client.knative.dev/user-image"    = "gcr.io/${var.project}/modron-ui:${var.env}"
      }
      labels = {
        "run.googleapis.com/startupProbeType" = "Default"
      }
    }
  }

  metadata {
    annotations = {
      "client.knative.dev/user-image" = "gcr.io/${var.project}/modron-ui:${var.env}"
      "run.googleapis.com/ingress"    = "internal-and-cloud-load-balancing"
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
  lifecycle {
    ignore_changes = [
      metadata[0].annotations["run.googleapis.com/operation-id"],
      metadata[0].annotations["run.googleapis.com/client-name"],
      metadata[0].annotations["run.googleapis.com/client-version"],
      template[0].metadata[0].annotations["run.googleapis.com/operation-id"],
      template[0].metadata[0].annotations["run.googleapis.com/client-name"],
      template[0].metadata[0].annotations["run.googleapis.com/client-version"],
    ]
  }
  depends_on = [
    google_project_service.run_service
  ]
}

resource "google_project_iam_member" "runner_log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.modron_runner.email}"
}

data "google_iam_policy" "cloud_run_invokers" {
  binding {
    role    = "roles/run.invoker"
    members = concat(var.modron_users, var.modron_admins)
  }
}

resource "google_cloud_run_service_iam_policy" "cloud_run_ui_invokers" {
  service     = google_cloud_run_service.ui.name
  location    = google_cloud_run_service.ui.location
  policy_data = data.google_iam_policy.cloud_run_invokers.policy_data
}

resource "google_cloud_run_service_iam_policy" "cloud_run_backend_invokers" {
  service     = google_cloud_run_service.grpc_web.name
  location    = google_cloud_run_service.grpc_web.location
  policy_data = data.google_iam_policy.cloud_run_invokers.policy_data
}
