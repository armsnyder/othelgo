#!/usr/bin/env bash

# This script runs the go benchmark on the current code, then checks out the previous commit and
# runs the benchmark a second time, then finally compares the two.

git diff-files --quiet || {
  echo "You must stash or commit changes before running this script"
  exit 1
}

HEAD_REF=$(git rev-parse --abbrev-ref HEAD)
HEAD_SHA=$(git rev-parse --short HEAD)
PREV_HEAD_SHA=$(git rev-parse --short HEAD^)

mkdir -p perf || exit 1

echo "Running benchmark on current version..."

go test -bench=. -count=5 ./... >>"perf/$HEAD_SHA.txt" || exit 1

git checkout "$PREV_HEAD_SHA" || exit 1

trap 'git checkout $HEAD_REF' EXIT

echo "Running benchmark on previous version..."

go test -bench=. -count=5 ./... >>"perf/$PREV_HEAD_SHA.txt" || exit 1

git checkout "$HEAD_REF" || exit 1

echo 'If the difference is statistically significant, a delta will be shown. Otherwise the delta will be "~":'

benchstat "perf/$PREV_HEAD_SHA.txt" "perf/$HEAD_SHA.txt"
