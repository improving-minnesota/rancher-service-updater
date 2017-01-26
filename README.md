# Rancher Service Updater

[![CircleCI](https://img.shields.io/circleci/project/github/objectpartners/rancher-service-updater.svg?maxAge=0)](https://circleci.com/gh/objectpartners/rancher-service-updater/tree/master)
[![GitHub Release](https://img.shields.io/github/release/objectpartners/rancher-service-updater.svg?maxAge=0)](https://github.com/objectpartners/rancher-service-updater/releases)
[![Apache License 2.0](https://img.shields.io/badge/license-Apache_License_2.0-blue.svg)](https://github.com/objectpartners/rancher-service-updater/blob/master/LICENSE)

An Inversion of Control (IOC) service used to notify Rancher of new container 
images and execute updates to those services based on container labels.

## Configuring

* `AUTOUPDATE_ENABLE_LABEL` [`autoupdate.enable`] - Specifies the container label to query for automatic update enabling.
* `AUTOUPDATE_ENVIRONMENT_NAMES` [`[".*"]`] - An array  of regex patterns to match Rancher environment names against. Environment name must match a pattern for auto-updating to occur in that environment.
* `AUTOUPDATE_HTTP_PORT` [`8080`] - The port that the service updater listens on.
* `AUTOUPDATE_SLACK_WEBHOOK_URL` - The webhook URL to use for sending Slack notifications. If not specified, Slack messaging is disabled.
* `AUTOUPDATE_SLACK_BOT_NAME` - The bot name to send as for Slack messages.
* `CATTLE_ACCESS_KEY` - The API access key for Rancher.
* `CATTLE_SECRET_KEY` - The API secret key for Rancher.
* `CATTLE_URL` - The Rancher server URL.

## Enabling Auto-Updating for Rancher Environment

The Rancher Service Updater configures itself to only check certain environments for services to automatically update.
This is done in conjunction with the API restrictions based on the credentials passed to the container.
That is, if an environment specific key is used, then the updater will only ever try to update service within that environment.

If however, a global/user key is used, then the updater can update services in any environment that that key has access to.
By default the updater is configured to allow updating of services in any environment.
This is set via the `AUTOUDPATE_ENVIRONMENT_NAMES` environment variable which should contain a comma (`,`) separated list of regex patterns.
The default value is `.*`.

In conjunction with a global key, this property can be used to restrict the environments.
For example, assume a cluster with 3 environments: `dev`, `qa`, and `production`.
The updater is configured with a global key that has access to all 3 environments.
To restrict automatic updates to only the `dev` environment, set `AUTOUPDATE_ENVIRONMENT_NAMES=dev`.
To restrict automatic updates to both the `dev` and `qa` environments, set `AUTOUPDATE_ENVIRONMENT_NAMES=dev,qa`


## Enabling Auto-Updating for a Rancher Service

Configure a service to be automatically updated by adding a container label of `autoupdate.enabled=true`.
Alternatively, the label to check can be specified by setting the `AUTOUPDATE_ENABLE_LABEL` environment variable

### Determining if update is required

| Currently Deployed Version | Newly Published Version | Update? |
|:---------------------------|:------------------------|:--------|
| 1.0                        | 2.0                     | `true`  |
| 2.0                        | 1.3                     | `false` |
| 1.0                        | latest                  | `true`  |
| latest                     | latest                  | `true`  |
| latest                     | 1.0                     | `false` |

## Running Service Updater on Rancher

The Rancher Service Updater relies upon the standard environment variables for providing 
connection information to create a Rancher client instance.

If running this application as a container on a Rancher cluster itself, you can 
provide the following service labels to automatically provision the API 
credentials to the application.

* `io.rancher.container.create_agent=true`
* `io.rancher.container.agent.role=environment`

_NOTE:_ When using this method, the provisioned API credentials are tied to the
Rancher environment in which the service is deployed. Thus, it will only have access
to update services within that same environment.

_NOTE:_ If it is intended to auto-update services in multiple Rancher environments,
then the API configuration **must** be provided via environment variables.

## Triggering an upgrade.

Send a post to `/upgrade` with the following JSON payload:

```
{
  "docker_image": "docker:",
  "confirm": true,
  "start_first": false,
  "timeout": 30
}
```

* `docker_image` - the image path that is now available. Optionally start with `docker:`
* `confirm` - if the service upgrade should be confirmed/finished if successful
* `start_first` - Optional. Default of `false`. If true, then sets new services to be started before terminated old services.
* `timeout` - Optional. Timeout in seconds. Default of 30. Timeout for waiting for service upgrade to complete if `confirm = true`.

## Security

This service provides not mechanism for authentication/authorization. It is the responsibility of the user to properly secure 
this service such that unauthorized access is not available.
