# PagerBot

Update your Slack user groups based on your PagerDuty Schedules.

Provided with API credentials and some configuration, PagerBot will
automatically update Slack user group membership and post a message to channels
you select informing everyone who's currently on the rotation.

PagerBot matches PagerDuty users to Slack users by their email addresses,
so your users must have the same email address in Slack as in PagerDuty.
PagerBot will log warnings for any users it finds in PagerDuty but not in Slack.

# Docker Image

[![](https://images.microbadger.com/badges/version/karlkfi/pagerbot.svg)](https://cloud.docker.com/repository/docker/karlkfi/pagerbot "Latest Image on DockerHub") [![](https://images.microbadger.com/badges/image/karlkfi/pagerbot.svg)](https://microbadger.com/images/karlkfi/pagerbot "Image Layers")

# Local Build

Use [goenv](https://github.com/syndbg/goenv) to install dependencies:

`goenv local`

Compile the `pagerbot` binary:

`go build`

You should have a nice `pagerbot` binary ready to go.

# Binary Releases

Cross-compiling can be done with [gox](https://github.com/mitchellh/gox):

```
# install gox
go get github.com/mitchellh/gox

# build binaries
gox -osarch "linux/amd64" -ldflags "-extldflags '-static'" -output "dist/{{.OS}}_{{.Arch}}/pagerbot"
```

You can also download pre-built binaries from the
[releases](https://github.com/karlkfi/pagerbot/releases) page.

# Slack Setup

1. [Create a Slack App](https://api.slack.com/apps)
2. Configure the App Scopes under `OAuth & Permissions`:
    - Send messages as PagerBot (chat:write:bot)
    - Post to specific channels in Slack (incoming-webhook)
    - Access basic information about the workspace’s User Groups (usergroups:read)
    - Change user’s User Groups (usergroups:write)
    - Access your workspace’s profile information (users:read)
    - View email addresses of people on this workspace (users:read.email)
3. Install the App and copy the `OAuth Access Token` (requires workspace admin)
4. Save the token `echo "SLACK_TOKEN=<token>" >> .secrets.env`

# PagerDuty Setup

1. [Create a read-only PagerDuty API Key](https://support.pagerduty.com/docs/using-the-api#section-generating-an-api-key) (requires account admin)
2. Save the key `echo "PAGERDUTY_KEY=<key>" >> .secrets.env`
3. Save the org name `echo "PAGERDUTY_ORG=<org>" >> .secrets.env`

# Config

A basic configuration file will look like

```yaml
api_keys:
  slack: "$SLACK_TOKEN" # Slack OAuth Access Token
  pagerduty:
    org: "$PAGERDUTY_ORG" # PagerDuty subdomain
    key: "$PAGERDUTY_KEY" # PagerDuty API key

groups:
- name: firefighter # name of the Slack user group to update
  schedules: # one or more PagerDuty schedule IDs
  - PAAAAAA
  - PBBBBBB
  update_message: # optional update message (%s is a comma delimited members list)
    message: ":fire_engine: @paas-oncall shift change: %s :fire_engine:"
    channels: # one or more channels to post the message to
    - paas
```

This config specifies the use of environment variables which pagerbot will
interpolate at runtime, allowing you to inject secrets and env vars.

Specify the config when launching pagerbot:

```
./pagerbot --config /path/to/config.yml --env-file /path/to/.secrets.env
```

# Deploy

It's recommended to run PagerBot using a package manager or container platform.

## Docker Image

The included Dockerfile uses a multi-stage build to compile pagebot and package
it in a small alpine-based Docker image for use in production:

```
# build and tag image
docker build -t karlkfi/pagerbot:latest .

# run locally in docker
docker run --env-file .secrets.env karlkfi/pagerbot:latest

# push to a docker image registry
docker push karlkfi/pagerbot:latest
```

## Kubernetes

Once a Docker image is built, it can be deployed to Kubernetes using
encrypted secret management:

```
# interpolate and base64 encode the secrets into a k8s secret
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: pagerbot-secrets
type: Opaque
data:
  slack-token: "$(grep 'SLACK_TOKEN' .secrets.env | cut -d'=' -f2 | tr -d '\n' | base64)"
  pagerduty-key: "$(grep 'PAGERDUTY_KEY' .secrets.env | cut -d'=' -f2 | tr -d '\n' | base64)"
  pagerduty-org: "$(grep 'PAGERDUTY_ORG' .secrets.env | cut -d'=' -f2 | tr -d '\n' | base64)"
EOF

# create pagerbot deployment
kubectl apply -f deployments/kubernetes/deployment.yml
```
