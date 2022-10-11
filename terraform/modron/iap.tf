resource "google_iap_brand" "project_brand" {
  support_email     = "modron-support${var.org_suffix}"
  application_title = "Modron ${title(var.env)}"
}

resource "google_iap_client" "project_client" {
  display_name = "Modron ${title(var.env)}"
  brand        = google_iap_brand.project_brand.name
}

data "google_iam_policy" "iap_web_users" {
  binding {
    role    = "roles/iap.httpsResourceAccessor"
    members = concat(var.modron_users, var.modron_admins)
  }
}

resource "google_iap_web_iam_policy" "users" {
  policy_data = data.google_iam_policy.iap_web_users.policy_data
}
