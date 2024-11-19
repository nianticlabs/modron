resource "google_cloud_run_v2_service" "grpc_web" {
  name     = "modron-grpc-web-${var.env}"
  location = local.region


  template {
    scaling {
      max_instance_count = 1
      min_instance_count = 1
    }

    service_account = google_service_account.modron_runner.email
    timeout         = "300s"
    containers {
      name       = "modron"
      depends_on = ["collector"]
      image      = "${local.region}-docker.pkg.dev/${var.project}/modron/modron:${var.env}"
      ports {
        container_port = 8080
        name           = "http1"
      }
      startup_probe {
        http_get {
          path = "/healthz"
          port = 8080
        }
        initial_delay_seconds = 10
        period_seconds        = 3
        failure_threshold     = 5 * 20 # 5 minutes (20 times 3s intervals = 1 minute)
      }
      resources {
        cpu_idle = false
        limits = {
          cpu    = "4000m"
          memory = "4Gi"
        }
      }
      env {
        name  = "ADDITIONAL_ADMIN_ROLES"
        value = join(",", var.additional_admin_roles)
      }
      env {
        name  = "ADMIN_GROUPS"
        value = join(",", [for g in var.modron_admins : split(":", g)[1]])
      }
      env {
        name  = "ALLOWED_SCC_CATEGORIES"
        value = join(",", var.allowed_scc_categories)
      }
      env {
        name = "DB_MAX_CONNECTIONS"
        # Max is 100 for Cloud SQL, but we may need some connections for other purposes.
        value = 90
      }
      env {
        name  = "ENVIRONMENT"
        value = "production"
      }
      env {
        name  = "IMPACT_MAP"
        value = jsonencode(var.impact_map)
      }
      env {
        name  = "LABEL_TO_EMAIL_REGEXP"
        value = var.label_to_email_regexp
      }
      env {
        name  = "LABEL_TO_EMAIL_SUBSTITUTION"
        value = var.label_to_email_substitution
      }
      env {
        name  = "LISTEN_ADDR"
        value = "0.0.0.0"
      }
      env {
        name  = "LOG_LEVEL"
        value = var.env == "dev" ? "debug" : "warning"
      }
      env {
        name  = "NOTIFICATION_INTERVAL_DURATION"
        value = "720h" // 30d
      }
      env {
        name  = "NOTIFICATION_SERVICE"
        value = var.notification_system
      }
      env {
        name  = "NOTIFICATION_SERVICE_CLIENT_ID"
        value = var.notification_system_client_id
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
        name  = "RULE_CONFIGS"
        value = var.rule_configs
      }
      env {
        name  = "SELF_URL"
        value = "https://${var.domain}"
      }
      env {
        name  = "SQL_BACKEND_DRIVER"
        value = "postgres"
      }
      env {
        name = "SQL_CONNECT_STRING"
        value_source {
          secret_key_ref {
            secret  = split("/", resource.google_secret_manager_secret.sql_connect_string_config.name)[3]
            version = "latest"
          }
        }
      }
      env {
        name  = "STORAGE"
        value = "sql"
      }
      env {
        name  = "TAG_CUSTOMER_DATA"
        value = "${var.org_id}/customer_data"
      }
      env {
        name  = "TAG_EMPLOYEE_DATA"
        value = "${var.org_id}/employee_data"
      }
      env {
        name  = "TAG_ENVIRONMENT"
        value = "${var.org_id}/environment"
      }

      volume_mounts {
        mount_path = "/cloudsql"
        name       = "cloudsql"
      }
    }

    containers {
      name  = "collector"
      image = "${var.docker_registry}/otel/opentelemetry-collector-contrib:0.111.0"
      startup_probe {
        http_get {
          path = "/"
          port = 13133
        }
      }

      volume_mounts {
        mount_path = "/etc/otelcol-contrib"
        name       = "otel-config"
      }
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [
          google_sql_database_instance.instance.connection_name
        ]
      }
    }

    volumes {
      name = "otel-config"
      gcs {
        bucket = google_storage_bucket.otel_config.name
      }
    }
  }

  ingress = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
  lifecycle {
    ignore_changes = [
      annotations["run.googleapis.com/operation-id"],
      annotations["run.googleapis.com/client-name"],
      annotations["run.googleapis.com/client-version"],
    ]
  }
  depends_on = [
    google_project_service.run_service
  ]
}

resource "google_cloud_run_v2_service" "ui" {
  name     = "modron-ui"
  location = local.region

  template {
    service_account = google_service_account.modron_runner.email
    timeout         = "300s"
    scaling {
      max_instance_count = 1
      min_instance_count = 1
    }
    containers {
      image = "${local.region}-docker.pkg.dev/${var.project}/modron/modron-ui:${var.env}"
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
  ingress = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
  lifecycle {
    ignore_changes = [
      annotations["run.googleapis.com/operation-id"],
      annotations["run.googleapis.com/client-name"],
      annotations["run.googleapis.com/client-version"],
    ]
  }
  depends_on = [
    google_project_service.run_service
  ]
}


data "google_iam_policy" "cloud_run_invokers" {
  binding {
    role    = "roles/run.invoker"
    members = concat(var.modron_users, var.modron_admins)
  }
}

resource "google_cloud_run_service_iam_policy" "cloud_run_ui_invokers" {
  service     = google_cloud_run_v2_service.ui.name
  location    = google_cloud_run_v2_service.ui.location
  policy_data = data.google_iam_policy.cloud_run_invokers.policy_data
}

resource "google_cloud_run_service_iam_policy" "cloud_run_backend_invokers" {
  service     = google_cloud_run_v2_service.grpc_web.name
  location    = google_cloud_run_v2_service.grpc_web.location
  policy_data = data.google_iam_policy.cloud_run_invokers.policy_data
}
