# Changelog

## v1.0.0

### Structure

- Support Resource Group Hierarchy

### Observations

- Add [Risk Score](docs/RISK_SCORE.md) to observations, calculated from the severity of the observation (as defined in the rule) and the impact of the observation (detected from the environment)
- Collectors can now collect observations

### Stats

- Improved stats view
- Improved export to CSV

### GCP

- Add support for [Security Command Center (SCC)](https://cloud.google.com/security-command-center/docs/concepts-security-command-center-overview)
- Start collecting Kubernetes resources

### Storage

- Use [GORM](https://gorm.io/) for both the PSQL and SQLite storage backends
- Use SQLite for the in-memory database for testing

### Performance

- Increase performance overall by optimizing the DB queries, parallelizing the scans, and reducing the number of external calls
- Introduce rate limiting for the collectors

### Observability

- Use [logrus](https://github.com/sirupsen/logrus) with structured logging for GCP Logging (Stackdriver)
- Add support for OpenTelemetry
 - Add an otel-collector to receive traces and metrics
 - Send traces to [Google Cloud Trace](https://cloud.google.com/trace)
 - Send metrics to [Google Cloud Monitoring](https://cloud.google.com/monitoring)

### UI

- Completely rework the UI with an improved design
- Show observations as a table, sorted by Risk Score by default
- Add a detailed view dialog for the observations

### Misc

- Use [`go-arg`](https://github.com/alexflint/go-arg) for the CLI arguments / environment variables
- Switch to [buf](https://buf.build/) for the protobuf generation
- Bug fixes
- Upgrade to Go 1.23
- Rules now support external configuration

## v0.2

- Moved to go 1.19 
- Added automated runs for scans 
- Fixed issue where last reported observation would still appear even if newer scans reported no observations 
- Fixed group member ship resolution when checking for accesses to GCP projects

## v0.1

- Initial public release
