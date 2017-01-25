# Rancher Service Updater

[![CircleCI](https://img.shields.io/circleci/project/github/objectpartners/rancher-service-updater.svg)](https://circleci.com/gh/objectpartners/rancher-service-updater/tree/master)
[![GitHub Release](https://img.shields.io/github/release/objectpartners/rancher-service-updater.svg)](https://github.com/objectpartners/rancher-service-updater/releases)

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
* 
