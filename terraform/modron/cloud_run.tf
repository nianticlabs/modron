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
          name  = "GCP_PROJECT_ID"
          value = var.project
        }
        env {
          name  = "DATASET_ID"
          value = "modron_bq"
        }
        env {
          name  = "RESOURCE_TABLE_ID"
          value = "resources"
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
          name  = "ENVIRONMENT"
          value = "PRODUCTION"
        }
        env {
          name  = "NOTIFICATION_SERVICE"
          value = "nagatha.example.com:443"
        }
        env {
          name  = "ORG_ID"
          value = var.org_id
        }
        env {
          name  = "ORG_SUFFIX"
          value = var.org_suffix
        }
      }
    }
    metadata {
      annotations = {
        "client.knative.dev/user-image"    = "gcr.io/${var.project}/modron:${var.env}"
        "autoscaling.knative.dev/minScale" = "1"
        "autoscaling.knative.dev/maxScale" = "1"
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
        "client.knative.dev/user-image"    = "gcr.io/${var.project}/modron-ui:${var.env}"
        "autoscaling.knative.dev/minScale" = "1"
        "autoscaling.knative.dev/maxScale" = "1"
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
