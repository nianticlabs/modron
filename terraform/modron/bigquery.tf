module "bigquery" {
  source                     = "terraform-google-modules/bigquery/google"
  dataset_id                 = "${var.dataset_id}_bq"
  dataset_name               = "${var.dataset_id}_bq"
  description                = "Modron storage backend"
  project_id                 = var.project
  location                   = "US"
  delete_contents_on_destroy = false
  tables                     = local.tables
  dataset_labels             = local.dataset_labels
  access                     = []
}

data "google_iam_policy" "bq_editor" {
  binding {
    role    = "roles/bigquery.dataOwner"
    members = concat(var.project_admins, ["serviceAccount:${data.google_service_account.terraform_sa.email}"])
  }
  binding {
    role = "roles/bigquery.dataEditor"
    members = [
      "serviceAccount:${google_service_account.modron_runner.email}"
    ]
  }
}

resource "google_bigquery_dataset_iam_policy" "editor" {
  dataset_id  = "${var.dataset_id}_bq"
  policy_data = data.google_iam_policy.bq_editor.policy_data
  depends_on = [
    module.bigquery
  ]
}

resource "google_project_iam_member" "bigquery_user" {
  project = var.project
  role    = "roles/bigquery.user"
  member  = "serviceAccount:${google_service_account.modron_runner.email}"
}
