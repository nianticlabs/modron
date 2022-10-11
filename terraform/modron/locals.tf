locals {
  dataset_labels = {
    env      = var.env
    billable = "true"
    owner    = "lds"
  }

  # Table definitions needed for BigQuery
  tables = [
    {
      table_id           = "resources",
      schema             = file("${path.module}/bigquery_schema_resource.json"),
      time_partitioning  = null,
      range_partitioning = null,
      expiration_time    = null
      clustering         = [],
      labels             = local.dataset_labels,
    },
    {
      table_id           = "observations",
      schema             = file("${path.module}/bigquery_schema_observation.json"),
      time_partitioning  = null,
      range_partitioning = null,
      expiration_time    = null
      clustering         = [],
      labels             = local.dataset_labels,
    },
    {
      table_id           = "operations",
      schema             = file("${path.module}/bigquery_schema_operation.json"),
      time_partitioning  = null,
      range_partitioning = null,
      expiration_time    = null
      clustering         = [],
      labels             = local.dataset_labels,
    }
  ]
}
