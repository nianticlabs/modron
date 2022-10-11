resource "google_iap_brand" "project_brand" {
  support_email     = "lds@example.com"
  application_title = "Nagatha"
}

resource "google_iap_client" "project_client" {
  display_name = "Nagatha"
  brand        = google_iap_brand.project_brand.name
}

data "google_iam_policy" "iap_web_users" {
  binding {
    role = "roles/iap.httpsResourceAccessor"
    members = [
      "group:nagatha-users@example.com",
    ]
  }
}

resource "google_iap_web_iam_policy" "users" {
  policy_data = data.google_iam_policy.iap_web_users.policy_data
}
