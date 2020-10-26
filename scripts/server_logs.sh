#!/usr/bin/env bash

LOG_STREAM_NAME=$(AWS_PAGER="" aws logs describe-log-streams \
  --region us-west-2 \
  --log-group-name /aws/lambda/othelgoServer \
  --order-by LastEventTime \
  --descending \
  --no-paginate \
  --query 'logStreams[0].logStreamName' \
  --output text)

AWS_PAGER="" aws logs get-log-events \
  --region us-west-2 \
  --log-group-name /aws/lambda/othelgoServer \
  --log-stream-name "$LOG_STREAM_NAME" \
  --query events \
  --output json |
  jq --raw-output '.[] | .timestamp |= (. / 1000 | strftime("%H:%M:%S")) | "\(.timestamp) \(.message)"'
