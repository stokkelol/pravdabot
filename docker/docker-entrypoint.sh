#!/usr/bin/env sh
set -ue

set -x ## debug trace

case $1 in
  test-integ|test-gocov)
    # -race not working on alpine: https://github.com/golang/go/issues/14481
    # CGO_ENABLED=1 go build -race -o study-manager-test .
    #gocov test . | gocov-xml > coverage.xml
    gocov test ./... | gocov report
    exit
    ;;
  linter)
    exec golangci-lint run ./...
    ;;
  air)
    shift
    cd ./app
    exec air "$@"
    ;;
  delve)
    shift
    cd ./app
    CGO_ENABLED=0 go build -x -gcflags="-N -l" -o pravdabot-delve .
    chmod +x ./notifications-delve
    dlv --listen=:40000 --headless=true --api-version=2 --accept-multiclient --log exec ./pravdabot-delve -- "$@"
    exit
    ;;
esac

exec "$@"