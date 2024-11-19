# Findings

## API_KEY_WITH_OVERBROAD_SCOPE
The API key is granting access to too many different scopes, or is not limited at all in what actions it allows. 
A malicious actor in possession of this key would be able to make a lot of damages to your infrastructure.

### Recommendation
Limit the scope of the API key to the smallest set of actions required by the user of this key to run properly. 
A list of scope is available in the [Google documentation](https://developers.google.com/identity/protocols/oauth2/scopes).

## BUCKET_IS_PUBLIC
A public bucket means that the content of this bucket is accessible to anybody on the internet. 
Make sure that the content of this bucket is actually intended to be public.  

> [!WARNING]  
> Do not assume that files with a cryptic name will never be found. These files will eventually be found. If files should stay private, then they should be hosted in a private bucket.

### Recommendation
Make this bucket private  
  
OR  
  
Make sure that the content of this bucket is intended to be public

## CLUSTER_NODES_HAVE_PUBLIC_IPS
On GCP, this means that you have a public cluster. There is no reason to have a public cluster today. Services that should be publicly accessible should be exposed using a load balancer.

For Airflow and Dataflow clusters, there is an option to set when starting the flows to use private cluster.

### Recommendation

> [!NOTE]  
> There is no way to transform a public cluster into a private one.
 
1. Create a new private cluster matching the specifications of the existing one
2. Migrate your workloads to the new cluster
3. Delete the old public cluster.

## CROSS_PROJECT_PERMISSIONS
The resource is controlled by an account defined in another project. 
This circumvents the isolation provided by a project.

### Recommendation
Use only accounts defined in the project to grant write and admin access to a resource.

## DATABASE_ALLOWS_UNENCRYPTED_CONNECTIONS
All connections to a database should use an encrypted connection. No clear text communication between a workload and a database should be allowed.

### Recommendation
Configure your database to allow only encrypted connections

## DATABASE_AUTHORIZED_NETWORKS_NOT_SET
Anyone can connect to this database without limitations.

### Recommendation
Add a list of IP or IP networks from which you expect connections and allow only connections from these networks.
This is also valid if your database is only available to internal IPs.

## EXPORTED_KEY_EXPIRY_TOO_LONG
An exported key has been around for too long.

Exported keys are immutable credentials that grant whoever has access to them a time unbounded access to our infrastructure. As people come and go, it is recommended to regularly rotate credentials to reduce the risk associated with leaks and malicious activity.

### Recommendation

- Rotate these credentials by deleting the existing one and creating a new one
- Create a process, possibly automated, to rotate credentials in the future and run this process regularly (every 3-6 months)

## EXPORTED_KEY_WITH_ADMIN_PRIVILEGES
An exported key in that resource group grants administrative privileges. Read EXPORTED_KEY_EXPIRY_TOO_LONG and add to the risk the fact that these credentials grant access to deployments, databases and possibly user-data.

### Recommendation

1. Remove the admin privileges of that service account or create a new service account with limited privileges
2. Rotate the key after this has been done

## HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE
A human user or a group has one of the following permissions at the project level:

- Owner 
- Editor
- Viewer
- Security Admin

### Recommendation

Use less privileged roles for humans. The principle of least privilege should be applied to all users.

## LOAD_BALANCER_MIN_TLS_VERSION_TOO_OLD
The load balancer supports a deprecated TLS version. 
The TLS version was deprecated because it supports broken cryptographic primitives.

### Recommendation
Define an SSL Policy at the project level or for the load balancer specifically that with a minimum TLS version 
of 1.2 and a MODERN or RESTRICTED profile.
See [defining an SSL policy](https://cloud.google.com/load-balancing/docs/ssl-policies-concepts#defining_an_ssl_policy) for more information.

## LOAD_BALANCER_USER_MANAGED_CERTIFICATE
A load balancer has been found with a user generated certificate. 

User generated certificates have multiple risks:
- The cryptographic material associated with that load balancer must be manually managed, leaving the door open to:
    - Certificate expiry if the certificates are not renewed in time 
    - Credentials leakage if anybody that has access to the cryptographic material is compromised or bad intended

- The generation of the private key has usually been done on a private machine where 
  - the Pseudo Random Number Generator has not been verified 
  - the entropy might have been too low

This weakens the encryption of the communication between the clients and the load balancer

### Recommendation
Migrate to a Google Managed certificate.

## KUBERNETES_VULNERABILITY_SCANNING_DISABLED

The GKE cluster is not using the [workload vulnerability scanning feature](https://cloud.google.com/kubernetes-engine/docs/concepts/about-workload-vulnerability-scanning). 
This means that container image vulnerabilities aren’t surfaced, or are only partially surfaced.

Follow the steps in [Automatically scan workloads for known vulnerabilities](https://cloud.google.com/kubernetes-engine/docs/how-to/security-posture-vulnerability-scanning)
to enable it.  Modron also suggests a command to run based on the name of your cluster and the project it is in.

## MASTER_AUTHORIZED_NETWORKS_NOT_SET
The administration interface of your Kubernetes cluster is available anyone who has access to a Google IP, or anybody on the internet as anybody can create a VM on GCP.

### Recommendation

Restrict the access to the Kubernetes API to a list of IP or IP networks from which you expect connections.

## OUTDATED_KUBERNETES_VERSION

The version of Kubernetes running in that cluster is not supported anymore.
Running outdated software is the first source of compromise. 
Running up-to-date software is the first barrier of defence against known vulnerabilities.

### Recommendation

- Update your Kubernetes cluster to a supported version 
- Onboard into release channel to benefit from automated updates in the future.

## PRIVATE_GOOGLE_ACCESS_DISABLED
The mentioned network contains some subnets that can have a preferred routing to the Google APIs without going through the Internet. It is recommended to use this routing pattern for security and latency reasons.

### Recommendation
Enable Private Google Access on all your subnetworks.

## SERVICE_ACCOUNT_TOO_HIGH_PRIVILEGES
This service account has too high privileges. In general we try to avoid granting too high privileges to service accounts.
Often time, permissions are granted at the project level where they should be granted at a more granular level. 
For instance, Service Account Token Creator at the project level allows for privilege escalation as it allows the 
service account that has this permission to get a token for any other service account in that project.  
  
Granting this permissions at the service account level only allows that service account to get a token 
for another specific service account.

### Recommendation
Limit the permissions of that service account to the strict minimum set of permissions required for this service 
account to run the tasks it is running.

## UNUSED_EXPORTED_CREDENTIALS
An exported credential is still valid but has not been used in a while.

### Recommendation
It is recommended to delete unused credentials and regenerate a new set when they are needed to prevent leaking 
of credentials and unauthorised access to our infrastructure.
