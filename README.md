# Courier

A stand-alone service that allows the GDS to deliver TRISA certificates via a webhook
rather than email. The service accepts PCKS12 passwords and encrypted certificates from
TRISA as HTTP `POST` requests and stores the certificates and passwords in either
Google Secret Manager or on the local disk (other secret management backends such as
Vault or Postgres may be available in the future).

This tool is mostly used by TRISA Service Providers (TSPs) who have to handle many
TRISA certificate deliveries at a time. VASPs who want to automate certificate delivery
may also use this service.

## Deploying with Docker

The simplest way to run the courier service is to use the docker image
`trisa/courier:latest` and to configure it from the environment. This allows the
courier service to be easily run on a Kubernetes cluster.

### Configuration

This application is configured via the environment. The following environment
variables can be used:

| KEY                                    | TYPE         | DEFAULT | DESCRIPTION                                                         |
|----------------------------------------|--------------|---------|---------------------------------------------------------------------|
| COURIER_MAINTENANCE                    | Boolean      | FALSE   | starts the server in maintenance mode                               |
| COURIER_BIND_ADDR                      | String       | :8842   | ip address and port of server                                       |
| COURIER_MODE                           | String       | release | either debug or release                                             |
| COURIER_LOG_LEVEL                      | LevelDecoder | info    | verbosity of logging: trace, debug, info, warn, error, fatal, panic |
| COURIER_CONSOLE_LOG                    | Boolean      | FALSE   | set for human readable logs (otherwise json logs)                   |
| COURIER_MTLS_INSECURE                  | Boolean      | TRUE    | set to false to enable TLS configuration                            |
| COURIER_MTLS_CERT_PATH                 | String       |         | the certificate chain and private key of the server                 |
| COURIER_MTLS_POOL_PATH                 | String       |         | the cert pool to validate clients for mTLS                          |
| COURIER_LOCAL_STORAGE_ENABLED          | Boolean      | FALSE   | set to true to enable local storage                                 |
| COURIER_LOCAL_STORAGE_PATH             | String       |         | path to the directory to store certs and passwords                  |
| COURIER_GCP_SECRET_MANAGER_ENABLED     | Boolean      | FALSE   | set to true to enable GCP secret manager                            |
| COURIER_GCP_SECRET_MANAGER_CREDENTIALS | String       |         | path to json file with gcp service account credentials              |
| COURIER_GCP_SECRET_MANAGER_PROJECT     | String       |         | name of gcp project to use with secret manager                      |