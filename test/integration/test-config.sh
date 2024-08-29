#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

##
# test config
##
echo "> test config"
RESULT=$($OUTPUT_DIR/cl config)
echo "$RESULT" | grep 'local_command_repository_dirname' | grep 'home' | grep -q 'current'
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - wrong config: local_command_repository_dirname"
  exit 1
fi

echo "> test get all config in json"
RESULT=$($OUTPUT_DIR/cl config --json)
VALUE=$(echo "$RESULT" | jq -r '.log_enabled')
if [ $VALUE == "false" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

echo "> test get one config in json"
RESULT=$($OUTPUT_DIR/cl config log_enabled --json)
VALUE=$(echo "$RESULT" | jq -r '.log_enabled')
if [ $VALUE == "false" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

echo "> test group_help_by_registry config exist, and default true"
RESULT=$($OUTPUT_DIR/cl config --json)
VALUE=$(echo "$RESULT" | jq -r '.group_help_by_registry')
if [ $VALUE == "true" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi
