# Modron 

## Build modron and push to Google Cloud Registry

```
gcloud --project modron-dev builds submit . --tag gcr.io/modron-dev/modron:dev --timeout=900
```

This applies the label `dev` on the image you're pushing.
This image is expected to run on modron-dev environment.

Deploy to cloud run dev:

```
gcloud --project modron-dev run deploy modron-dev --platform=managed --image=gcr.io/modron-dev/modron:dev --region=us-central1 --service-account=$DEV_RUNNER_SA_NAME
```

## Debug

To debug RPC issues, set the two following environment variables:

```
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

## Update libraries

```
CYPRESS_CACHE_FOLDER=/tmp npm install
```

Note: Cypress tries to write to /root/.cache which doesn't work. This is why we need to set the environment variable.