# Modron 

## Build modron and push to Google Cloud Registry

```
gcloud builds submit . --tag gcr.io/modron-dev/modron:dev --timeout=900
```

This applies the label `dev` on the image you're pushing.
This image is expected to run on modron-dev environment.

Deploy to cloud run dev:

```
DEV_RUNNER_SA_NAME=$PROJECT-runner@$PROJECT.iam.gserviceaccount.com
gcloud run deploy modron-grpc-web-dev --platform=managed --image=gcr.io/modron-dev/modron:dev --region=us-central1 --service-account=$DEV_RUNNER_SA_NAME
gcloud run services update-traffic modron-ui --to-revisions LATEST=100 --region=us-central1
```

## Debug

To debug RPC issues, set the two following environment variables:

```
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

## Update libraries

```
CYPRESS_CACHE_FOLDER=/tmp npm upgrade
CYPRESS_CACHE_FOLDER=/tmp npm install
```

Note: Cypress tries to write to /root/.cache which doesn't work. This is why we need to set the environment variable.
