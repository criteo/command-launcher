#!/bin/sh

SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}
echo "integration test directory: $SCRIPT_DIR"

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
echo $RESULT
echo "$OUTPUT_DIR"
echo "$OUTPUT_DIR/home/current"
echo $RESULT | grep "local_command_repository_dirname : $OUTPUT_DIR" | grep 'home' | grep 'current'
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - wrong config: local_command_respository_dirname"
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
RESULT=$($OUTPUT_DIR/cl config command_repository_base_url https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo)
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
if [ "$RESULT" = "Hello World!" ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

##
# remove the output folder
##
echo "clean up"
rm -rf $OUTPUT_DIR
