# Nagatha

Nagatha is a tool used to notify people about issues to fix. 
People can be notified multiple times until the issues actually get fixed.

## How does it work?

1.   Nagatha gets a list of notification as input from different sources inside your organization
1.   Notifications are deduplicated (do not set the same notification twice to the same person)
1.   Notifications are aggregated (send only one big list of notifications to each recipient)
1.   Cloud scheduler triggers the notify all daily to send all pending notifications

Each system can decide how often people should get notified by setting the notification interval in the request.
Nagatha will check upon the notification creation request if a notification has already been created in the given interval.

## Local run

```
PROJECT=<your-project>
GOOGLE_APPLICATION_CREDENTIALS=$HOME/.config/gcloud/application_default_credentials.json PORT=8080 GCP_PROJECT_ID=$PROJECT EXCEPTION_TABLE_ID=nagatha_bq.exceptions NOTIFY_TRIGGER_SUBSCRIPTION=notify-all NOTIFICATION_TABLE_ID=nagatha_bq.notifications EMAIL_SENDER_ADDRESS=test@example.com go run . --logtostderr
```

With Docker:

```
docker build -t gcr.io/nagatha/nagatha:test .
docker run -e PORT=8080 -e GCP_PROJECT_ID=nagatha  -e EXCEPTION_TABLE_ID=nagatha_bq.exceptions -e NOTIFY_TRIGGER_SUBSCRIPTION=notify-all -e NOTIFICATION_TABLE_ID=nagatha_bq.notifications -e EMAIL_SENDER_ADDRESS=test@example.com -p8080:8080 gcr.io/nagatha/nagatha:test
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
gcloud --project nagatha builds submit . --tag gcr.io/$PROJECT/nagatha:dev --timeout=900
```

## At the org level

* Create a `nagatha-users` group

## Build the proto

Install go dependencies if needed:

```
go install golang.org/x/tools/gopls \
               mvdan.cc/gofumpt \
               github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
               github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
               google.golang.org/protobuf/cmd/protoc-gen-go \
               google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

```
/usr/local/protoc/bin/protoc --plugin=$HOME/go/bin/protoc-gen-go  --plugin=$HOME/go/bin/protoc-gen-go-grpc  -I=proto --go_out=proto/.   --grpc-gateway_out=logtostderr=true:./proto --go-grpc_out=proto/ proto/nagatha.proto
```
