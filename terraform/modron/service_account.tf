resource "google_service_account" "modron_runner" {
  account_id   = "modron-${var.env}-runner"
  description  = "Modron ${var.env} runner"
  display_name = "modron-${var.env}-runner"
}
