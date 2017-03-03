#!/bin/bash

set -e

MATRIX=(
  ""
)

show() {
  echo "$@" >&2
  eval "$@"
}

fatal() {
  echo "$@" >&2
  exit 1
}

PKGS="$(go list ./... | grep -v /vendor/)"

i=0

for pkg in $PKGS; do
  for opts in "${MATRIX[@]}"; do
    show go test -timeout 30s -v -coverprofile=test-$i.cov $opts $pkg \
      || fatal "go test errored"
    let i++ || true
  done
done

if [[ ! $LATEST_GO ]]; then
  echo "Skipping linting and coverage report"
  exit 0
fi

for opts in "${MATRIX[@]}"; do
  show go vet -v $opts $PKGS || fatal "go vet errored"
done

GOFILES=$(find * -name '*.go' -not -path 'vendor/*')

FMT_FILES="$(gofmt -s -l $GOFILES)"
if [[ -n $FMT_FILES ]]; then
  fatal "Run 'gofmt -s -w' on these files:\n$FMT_FILES"
fi

IMP_FILES="$(goimports -local github.com/TykTechnologies -l $GOFILES)"
if [[ -n $IMP_FILES ]]; then
  fatal "Run 'goimports -local github.com/TykTechnologies -w' on these files:\n$IMP_FILES"
fi
