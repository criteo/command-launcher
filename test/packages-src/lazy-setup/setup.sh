#!/bin/sh

# the setup hook only receives the process environment, so locate the package dir from $0
echo "calling lazy setup"
echo run >> "$(dirname "$0")/setup.count"
