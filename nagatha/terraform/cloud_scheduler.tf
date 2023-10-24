# The AppEngine application is required for the cloud scheduler.
resource "google_app_engine_application" "ae_app" {
  # Appengine uses different location.
  location_id = trimsuffix(var.region, "1")
}


resource "google_cloud_scheduler_job" "job" {
  name        = "notify-all"
  description = "Pubsub call to trigger notifications."
  schedule    = "42 7 * * *" # Every day
  time_zone   = "America/New_York"
  # This should be enough, the call is asynchronous and returns as soon as the project list has been sent to pubsub.

  paused = false

  retry_config {
    retry_count          = 1
    min_backoff_duration = "3600s" # Wait 1h before retry
  }

  pubsub_target {
    topic_name = google_pubsub_topic.notify_all_trigger.id
    data       = base64encode("non-empty-message")
  }

  depends_on = [
    google_project_service.cloudscheduler_service,
    google_app_engine_application.ae_app
  ]
}

# The topic to trigger the project list
resource "google_pubsub_topic" "notify_all_trigger" {
  name = "notify-all-trigger"
}

# Allow the publishing role to push to the asset feed.
# data "google_iam_policy" "notify_all_publisher_policy" {
#   binding {
#     role = "roles/pubsub.publisher"
#     members = [
#       "serviceAccount:${google_service_account.nagatha_sa.email}"
#     ]
#   }
# }

# resource "google_pubsub_topic_iam_policy" "notify_all_trigger_publisher" {
#   topic       = google_pubsub_topic.notify_all_trigger.id
#   policy_data = data.google_iam_policy.notify_all_publisher_policy.policy_data
# }

resource "google_pubsub_subscription" "notify_all_sub" {
  topic = google_pubsub_topic.notify_all_trigger.id
  name  = "notify-all"
}

data "google_iam_policy" "notify_all_subscriber_policy" {
  binding {
    role = "roles/pubsub.subscriber"
    members = [
      "serviceAccount:${google_service_account.nagatha_sa.email}"
    ]
  }
  binding {
    role = "roles/pubsub.viewer"
    members = [
      "serviceAccount:${google_service_account.nagatha_sa.email}"
    ]
  }
}

resource "google_pubsub_subscription_iam_policy" "notify_all_trigger_subscriber" {
  subscription = google_pubsub_subscription.notify_all_sub.id
  policy_data  = data.google_iam_policy.notify_all_subscriber_policy.policy_data
}
