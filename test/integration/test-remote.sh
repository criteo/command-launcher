#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

EXAMPLE_BRANCH_NAME=main
##
# test remote command
##
echo "> test download remote command"
RESULT=$($OUTPUT_DIR/cl config command_repository_base_url https://raw.githubusercontent.com/criteo/command-launcher/${EXAMPLE_BRANCH_NAME}/examples/remote-repo)
RESULT=$($OUTPUT_DIR/cl)

echo "$RESULT"

echo "$RESULT" | grep -q "hello"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - hello command should exist"
  exit 1
fi

echo "> test run remote command"
RESULT=$($OUTPUT_DIR/cl hello)
echo "$RESULT" | grep -q "Hello World!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

echo "> test remote config"
export CL_REMOTE_CONFIG_URL=https://raw.githubusercontent.com/criteo/command-launcher/${EXAMPLE_BRANCH_NAME}/examples/remote-config/remote_config.json
RESULT=$($OUTPUT_DIR/cl config)
echo "$RESULT" | grep -q "test/remote-repo"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - remote config didn't set correctly"
  exit 1
fi

echo "> test update command"
RESULT=$($OUTPUT_DIR/cl update --package)
echo "$RESULT" | grep "upgrade package 'command-launcher-demo' from version 1.0.0 to version 2.0.0"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - failed to run update command"
  exit 1
fi

echo "> test update command updates bonjour package"
RESULT=$($OUTPUT_DIR/cl)
echo "$RESULT" | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - bonjour command should exist"
  exit 1
fi

echo "> test bonjour command from remote config"
RESULT=$($OUTPUT_DIR/cl bonjour)
echo "$RESULT" | grep -q "bonjour!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of bonjour command: $RESULT"
  exit 1
fi

echo "> test downloaded package specified from a remote config"
RESULT=$($OUTPUT_DIR/cl hello)
echo "$RESULT" | grep -q "Hello World v2!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi
