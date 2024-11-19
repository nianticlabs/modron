# gcp_service_agents

GCP publishes a list of "Service Agents" on their [documentation pages](https://cloud.google.com/iam/docs/service-agents).  
Unfortunately this list is not in a machine readable format. This little helper scrapes that page and provides a list
of project IDs (e.g: `service-PROJECT_NUMBER@gcp-sa-aiplatform-cc.iam.gserviceaccount.com` -> `gcp-sa-aiplatform-cc`)
that Google provides, and thus that are considered "secure" to be used in IAM policies.

## Usage

```bash
go run ./ -o out.json
jq -r '.projects[] | "\"" + . + "\"" + ": {},"' out.json | clipcopy
```

Then paste the content of your clipboard into the `constants/gcp_sa_projects.go` file.