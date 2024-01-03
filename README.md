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
