#!/bin/bash

SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}
echo "integration test directory: $SCRIPT_DIR"

#BRANCH_NAME=$(git branch --show-current)
EXAMPLE_BRANCH_NAME=main

# create output folder
OUTPUT_DIR=$SCRIPT_DIR/output
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

##
# build the binary
##
cd $SCRIPT_DIR
go build -o $OUTPUT_DIR/cl -ldflags='-X main.version=integration-test -X main.buildNum=1 -X main.appName=cl -X "main.appLongName=Command Launcher"' $SCRIPT_DIR/../main.go

# specify the app home
export CL_HOME=$OUTPUT_DIR/home


RESULT=$($OUTPUT_DIR/cl)

##
# test application name
##
echo "> test application name"
echo $RESULT | grep -q "Command Launcher - A command launcher"
if [ $? -ne 0 ]; then
  echo "KO - wrong application name"
  exit 1
else
  echo "OK"
fi

##
# test home folder & loacl repository
##
echo "> test home folder & local repository"

if [ -d "$OUTPUT_DIR/home/current" ]; then
  # ok
  echo "OK"
else
  echo "KO - local repository should exist"
  exit 1
fi

##
# test command list
##
echo "> test default command list"

echo $RESULT | grep -q "hello"
if [ $? -ne 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - hello command shouldn't exist"
  exit 1
fi


##
# test config
##
echo "> test config"
RESULT=$($OUTPUT_DIR/cl config)
echo $RESULT | grep 'local_command_repository_dirname' | grep 'home' | grep -q 'current'
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - wrong config: local_command_repository_dirname"
  exit 1
fi

echo "> test get all config in json"
RESULT=$($OUTPUT_DIR/cl config --json)
VALUE=$(echo $RESULT | jq -r '.log_enabled')
if [ $VALUE == "false" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

echo "> test get one config in json"
RESULT=$($OUTPUT_DIR/cl config log_enabled --json)
VALUE=$(echo $RESULT | jq -r '.log_enabled')
if [ $VALUE == "false" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

##
# test exit code
##
# First copy the dropin packages for test
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/packages-src/exit-code $CL_HOME/dropins

echo "> test exit code - success case"
RESULT=$($OUTPUT_DIR/cl exit0)
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should return 0 when command succeeds"
  exit 1
fi

echo "> test exit code - failure case"
RESULT=$($OUTPUT_DIR/cl exit1)
if [ $? -eq 1 ]; then
  echo "OK"
else
  echo "KO - should return non-0 when command fails"
  exit 1
fi

##
# test set config
##
echo "> test set config"
RESULT=$($OUTPUT_DIR/cl config log_level debug)
RESULT=$($OUTPUT_DIR/cl config)
echo $RESULT | grep -q "log_level : debug"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - failed to set config: log_level"
  exit 1
fi

echo "> test get single config"
RESULT=$($OUTPUT_DIR/cl config log_level)
if [ "$RESULT" = "debug" ]; then
  echo "OK"
else
  echo "KO - failed to get config: log_level"
  exit 1
fi

##
# test remote command
##
echo "> test download remote command"
RESULT=$($OUTPUT_DIR/cl config command_repository_base_url https://raw.githubusercontent.com/criteo/command-launcher/${EXAMPLE_BRANCH_NAME}/examples/remote-repo)
RESULT=$($OUTPUT_DIR/cl)
echo $RESULT | grep -q "hello"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - hello command should exist"
  exit 1
fi

echo "> test run remote command"
RESULT=$($OUTPUT_DIR/cl hello)
echo $RESULT | grep -q "Hello World!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

echo "> test remote config"
export CL_REMOTE_CONFIG_URL=https://raw.githubusercontent.com/criteo/command-launcher/${EXAMPLE_BRANCH_NAME}/examples/remote-config/remote_config.json
RESULT=$($OUTPUT_DIR/cl config)
echo $RESULT | grep -q "test/remote-repo"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - remote config didn't set correctly"
  exit 1
fi

echo "> test update command"
RESULT=$($OUTPUT_DIR/cl update --package)
echo $RESULT | grep "upgrade command 'command-launcher-demo' from version 1.0.0 to version 2.0.0"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - failed to run update command"
  exit 1
fi

echo "> test update command updates bonjour package"
RESULT=$($OUTPUT_DIR/cl)
echo $RESULT | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - bonjour command should exist"
  exit 1
fi

echo "> test bonjour command from remote config"
RESULT=$($OUTPUT_DIR/cl bonjour)
echo $RESULT | grep -q "bonjour!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of bonjour command: $RESULT"
  exit 1
fi

echo "> test downloaded package specified from a remote config"
RESULT=$($OUTPUT_DIR/cl hello)
echo $RESULT | grep -q "Hello World v2!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

echo "> test the LOG_LEVEL command"
export CL_LOG_LEVEL=Debug
RESULT=$($OUTPUT_DIR/cl hello)
echo $RESULT | grep -q "Debug"
if [ $? -eq 1 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi
UNSET CL_LOG_LEVEL

##
# remove the output folder
##
echo "clean up"
rm -rf $OUTPUT_DIR
