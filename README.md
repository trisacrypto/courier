# Courier

A stand-alone service that allows the GDS to deliver TRISA certificates via a webhook
rather than email. The service accepts PCKS12 passwords and encrypted certificates from
TRISA as HTTP `POST` requests and stores the certificates and passwords in either
Google Secret Manager or on the local disk (other secret management backends such as
Vault or Postgres may be available in the future).

This tool is mostly used by TRISA Service Providers (TSPs) who have to handle many
TRISA certificate deliveries at a time. VASPs who want to automate certificate delivery
may also use this service.

## Storage

Courier can be configured to store PKCS12 passwords and x509 certificates in different backends. Currently available backends include:

1. **Local storage**: stored as gzip text files in a specified directory
2. **Google Secret Manager**: stored using Google Cloud Platform secrets

At least one storage backend must be configured for Courier to function properly. If there is another storage backend that you would like implemented for Courier, please [create an issue to request it](https://github.com/trisacrypto/courier/issues)!

## Deploying

Courier is intended to be set up and run in your local environment. **We strongly recommend that you ensure the webhook is TLS encrypted**. Once you have a courier service setup, you can update the GDS with webhook delivery instructions.

### Docker

The simplest way to run the courier service is to use the docker image
`trisa/courier:latest` and to configure it from the environment. This allows the
courier service to be easily run on a Kubernetes cluster.

### Build and Run

Alternatively you can build and run the courier executable on your own instance. Use Go to build and install the executable as follows:

```
$ go install github.com/trisacrypto/courier/cmd/courier
```

You can then run the server as follows:

```
$ courier serve
```

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