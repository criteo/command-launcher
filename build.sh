#!/usr/bin/env sh

DEFAULT_VERSION=$(git rev-parse --abbrev-ref HEAD)-dev

VERSION=${1:-$DEFAULT_VERSION}
APP_NAME=${2:-cdt}
APP_LONG_NAME=${3:-Criteo Dev Toolkit}
RESIGN=${4:-}

go build -o $APP_NAME -ldflags="-X main.version=$VERSION -X main.buildNum=$(date +'%Y%m%d-%H%M%S') -X main.appName=$APP_NAME -X 'main.appLongName=$APP_LONG_NAME'"

if [ "$RESIGN" = "--resign" ] && [ "$(uname)" = "Darwin" ]; then
    codesign --force -s - "$APP_NAME"
fi
