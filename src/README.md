# Modron 

## Build modron and push to Google Cloud Registry

```bash
gcloud builds submit . --tag us-central1.docker.pkg.dev/$PROJECT_ID/modron/modron:dev --timeout=900
```

This applies the label `dev` on the image you're pushing.
This image is expected to run on modron-dev environment.

Deploy to cloud run dev:

```bash
DEV_RUNNER_SA_NAME=$PROJECT_ID-runner@$PROJECT_ID.iam.gserviceaccount.com
REGION=us-central1
gcloud run deploy \
  modron-grpc-web-dev \
  --platform=managed \
  --image="$REGION.docker.pkg.dev/$PROJECT_ID/modron/modron:dev" \
  --region="$REGION" \
  --service-account="$DEV_RUNNER_SA_NAME"
gcloud run services update-traffic modron-ui --to-revisions LATEST=100 --region="$REGION"
```

## Debug

To debug RPC issues, set the two following environment variables:

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

## Update libraries

```bash
export CYPRESS_CACHE_FOLDER=/tmp
npm upgrade
npm install
```

Note: Cypress tries to write to `/root/.cache` which doesn't work. This is why we need to set the environment variable.
