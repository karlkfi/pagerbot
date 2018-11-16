# PagerBot

Update your Slack user groups based on your PagerDuty Schedules.

Provided with API credentials and some configuration, PagerBot will automatically update Slack user group membership and post a message to channels you select informing everyone who's currently on the rotation.

# Build

We use [goenv](https://github.com/syndbg/goenv) so:

`goenv local`

Then build

`go build`

You should have a nice `pagerbot` binary ready to go. You can also download prebuild binaries from the [releases](https://github.com/karlkfi/pagerbot/releases) page.

# Slack Setup

1. [Create a Slack App](https://api.slack.com/apps)
2. Configure the App Scopes under `OAuth & Permissions`
3. Install the App and copy the `OAuth Access Token` (requires workspace admin)
4. Write the token to `echo "SLACK_TOKEN=<token>" >> .ci-runner.env`

# PagerDuty Setup

1. [Create a read-only PagerDuty API Key](https://support.pagerduty.com/docs/using-the-api#section-generating-an-api-key) (requires account admin)
2. Write the key `echo "PAGERDUTY_KEY=<key>" >> .ci-runner.env`
3. Write the org name `echo "PAGERDUTY_ORG=<org>" >> .ci-runner.env`

# Config

A basic configuration file will look like

```yaml
api_keys:
  slack: "abcd123"
  pagerduty:
    org: "songkick"
    key: "qwerty567"

groups:
  - name: firefighter
    schedules:
      - PAAAAAA
      - PBBBBBB
    update_message:
      message: ":fire_engine: Your firefighters are %s :fire_engine:"
      channels:
        - general
  - name: fielder
    schedules:
      - PCCCCCC
    update_message:
      message: "Your :baseball: TechOps @Fielder :baseball: this week is %s"
      channels:
        - team-engineering
```

The configuration should be fairly straightforward, under API keys provide your Slack and Pagerduty keys. Under groups configure the Slack groups you'd like to update. Schedules is a list of PagerDuty schedule IDs, update_message is the message you'd like to post, and the channels you'd like to post them in.

Once done, you can run PagerBot with `./pagerbot --config /path/to/config.yml`

It's recommended to run PagerBot under Upstart or some other process manager.

N.B. PagerBot matches PagerDuty users to Slack users by their email addresses, so your users must have the same email address in Slack as in PagerDuty. PagerBot will log warnings for any users it finds in PagerDuty but not in Slack.

# Deploy

```
sudo docker build -t gcr.io/cruise-gcr-dev/karlkfi/pagerbot:latest .
sudo docker run --env-file .ci-runner.env gcr.io/cruise-gcr-dev/karlkfi/pagerbot:latest

kubeenv <namespace>

kubectl create -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: pagerbot-secrets
type: Opaque
data:
  slack-token: <base64-secret>
  pagerduty-key: <base64-secret>
  pagerduty-org: <base64-secret>
EOF

kubectl create -f pagerbot-kubernetes.yml

```
