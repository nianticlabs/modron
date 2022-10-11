# Nagatha

Nagatha is a tool used to notify people about issues to fix. 
People can be notified multiple times until the issues actually get fixed.

## Local run

```
PROJECT=<your-project>
GOOGLE_APPLICATION_CREDENTIALS=$HOME/.config/gcloud/application_default_credentials.json PORT=8080 GCP_PROJECT_ID=$PROJECT EXCEPTION_TABLE_ID=nagatha_bq.exceptions NOTIFICATION_TABLE_ID=nagatha_bq.notifications EMAIL_SENDER_ADDRESS=test@example.com go run . --logtostderr
```

With Docker:

```
docker build -t gcr.io/nagatha/nagatha:test .
docker run -e PORT=8080 -e GCP_PROJECT_ID=nagatha  -e EXCEPTION_TABLE_ID=nagatha_bq.exceptions -e NOTIFICATION_TABLE_ID=nagatha_bq.notifications -e EMAIL_SENDER_ADDRESS=test@example.com -p8080:8080 gcr.io/nagatha/nagatha:test
```

Get a GRPC UI:

```
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
grpcui -plaintext -proto proto/nagatha.proto localhost:8080
```

Trigger NotifyAll with curl

```
PROJECT=<your-project>
grpcurl -H "Authorization:bearer $(gcloud auth print-identity-token --impersonate-service-account=caller@$PROJECT.iam.gserviceaccount.com --audiences app-id.apps.googleusercontent.com --include-email test@example.com)" -H "Content-Type:application/json" -vv -proto "proto/nagatha.proto" -d "" nagatha.example.com:443 Nagatha.NotifyAll
```

## Build

```
gcloud --project nagatha builds submit . --tag gcr.io/nagatha/nagatha:dev --timeout=900
```

## At the org level

* Create a `nagatha-users` group
