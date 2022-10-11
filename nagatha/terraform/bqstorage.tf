locals {
  dataset_labels = {
    env      = var.env
    billable = "true"
    owner    = replace(var.bq_owner[0], "@example.com", "_nianticlabs_com")
  }
  tables = [
    {
      table_id           = "exceptions",
      schema             = file("${path.module}/bqschemas/exceptions_schema.json"),
      time_partitioning  = null,
      range_partitioning = null,
      expiration_time    = null,
      clustering         = [],
      labels             = local.dataset_labels,
    },
    {
      table_id           = "notifications",
      schema             = file("${path.module}/bqschemas/notifications_schema.json"),
      time_partitioning  = null,
      range_partitioning = null,
      expiration_time    = null,
      clustering         = [],
      labels             = local.dataset_labels,
    },
  ]
  dataset_id = "nagatha"
}

module "bigquery" {
  source                     = "terraform-google-modules/bigquery/google"
  dataset_id                 = "${local.dataset_id}_bq"
  dataset_name               = "${local.dataset_id}_bq"
  description                = "Nagatha storage backend"
  project_id                 = var.project
  location                   = "US"
  delete_contents_on_destroy = "false"
  tables                     = local.tables
  dataset_labels             = local.dataset_labels
  access                     = []
}

resource "google_service_account" "bigquerysa" {
  account_id   = "nagatha-bigquerysa"
  display_name = "Nagatha BigQuery Service Account"
}

resource "google_bigquery_dataset_iam_policy" "bq_iam_policy_binding" {
  dataset_id  = "${local.dataset_id}_bq"
  policy_data = data.google_iam_policy.bq_iam_policy.policy_data
}

data "google_iam_policy" "bq_iam_policy" {
  binding {
    role = "roles/bigquery.dataOwner"
    members = [
      "group:nagatha-admins@example.com",
      "serviceAccount:${data.google_service_account.terraform_sa.email}"
    ]
  }
  binding {
    role = "roles/bigquery.dataEditor"
    members = [
      "serviceAccount:${google_service_account.nagatha_sa.email}"
    ]
  }
}

resource "google_project_iam_member" "bigquery_user" {
  project = var.project
  role    = "roles/bigquery.user"
  member  = "serviceAccount:${google_service_account.nagatha_sa.email}"
}
