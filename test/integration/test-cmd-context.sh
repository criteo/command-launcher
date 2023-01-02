#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

##
# test command context
##
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/bonjour $CL_HOME/dropins

echo "> test the command without LOG_LEVEL"
RESULT=$($OUTPUT_DIR/cl bonjour)
echo "$RESULT" | grep -q "bonjour!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

echo "> test set config"
RESULT=$($OUTPUT_DIR/cl config log_level debug)
RESULT=$($OUTPUT_DIR/cl config)
echo "$RESULT" | grep -q "log_level                               : debug"
if [ $? -eq 0 ]; then
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

echo "> test the command with LOG_LEVEL"
RESULT=$($OUTPUT_DIR/cl bonjour)
echo $RESULT | grep -q "bonjour! debug"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

echo "> test default checkFlags = false, no flag and arg environment should be injected"
RESULT=$($OUTPUT_DIR/cl bonjour --name Joe --language French world)
echo "$RESULT" | grep -q "Joe"
if [ $? -eq 0 ]; then
  echo "KO - no environment variable CL_FLAG_NAME should be found"
  exit 1
else
  echo "OK"
fi

echo "$RESULT" | grep -q "French"
if [ $? -eq 0 ]; then
  echo "KO - no environment variable CL_FLAG_LANGUAGE should be found"
  exit 1
else
  echo "OK"
fi

echo "$RESULT" | grep -q "world"
if [ $? -eq 0 ]; then
  echo "KO - no environment variable CL_ARG_1 should be found"
  exit 1
else
  echo "OK"
fi

echo "> test PACKAGE_DIR environment variable"
RESULT=$($CL_PATH bonjour)
echo "$RESULT" | grep "home"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have PACKAGE_DIR environment variable"
  exit 1
fi

