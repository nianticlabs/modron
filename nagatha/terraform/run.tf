resource "google_cloud_run_service" "nagatha" {
  name     = "nagatha"
  location = "us-central1"

  # We need this to avoid naming collision with deployments.
  autogenerate_revision_name = true

  template {
    spec {
      service_account_name = google_service_account.nagatha_sa.email
      timeout_seconds      = 300
      containers {
        image = "gcr.io/${var.project}/nagatha:dev"
        ports {
          container_port = 8080
          name           = "http1"
        }
        env {
          name  = "EXCEPTION_TABLE_ID"
          value = "nagatha_bq.exceptions"
        }
        env {
          name  = "EMAIL_SENDER_ADDRESS"
          value = var.email_sender_address
        }
        env {
          name  = "GCP_PROJECT_ID"
          value = var.project
        }
        env {
          name  = "NOTIFICATION_TABLE_ID"
          value = "nagatha_bq.notifications"
        }
        env {
          name  = "NOTIFY_TRIGGER_SUBSCRIPTION"
          value = google_pubsub_subscription.notify_all_sub.name
        }
        env {
          name = "SENDGRID_API_KEY"
          value_from {
            secret_key_ref {
              key  = "latest"
              name = split("/", resource.google_secret_manager_secret.sendgrid_api_key_config.name)[3]
            }
          }
        }
        resources {
          limits = {
            cpu    = "1000m"
            memory = "256Mi"
          }
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = "1"
        "autoscaling.knative.dev/maxScale" = "1"
      }
    }
  }

  metadata {
    annotations = {
      "client.knative.dev/user-image"  = "gcr.io/${var.project}/nagatha:dev"
      "run.googleapis.com/client-name" = "terraform"
      "run.googleapis.com/ingress"     = "internal-and-cloud-load-balancing"
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

resource "google_project_iam_member" "log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.nagatha_sa.email}"
}

resource "google_service_account" "nagatha_sa" {
  account_id   = "nagatha"
  description  = "Nagatha account"
  display_name = "nagatha"
}
