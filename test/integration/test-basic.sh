#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR

SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}


RESULT=$($OUTPUT_DIR/cl)

##
# test application name
##
echo "> test application name"
echo "$RESULT" | grep -q "Command Launcher - A command launcher"
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

echo "$RESULT" | grep -q "hello"
if [ $? -ne 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - hello command shouldn't exist"
  exit 1
fi

##
# test help message
##

# clean up the dropin folder
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins

# copy the example to the dropin folder for the test
cp -R $SCRIPT_DIR/../packages-src/bonjour $CL_HOME/dropins

echo "> test group help message"
RESULT=$($CL_PATH)
echo "$RESULT" | grep -q "Commands from 'dropin' registry"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should group help message by registry"
  exit 1
fi

echo "> test help message without group"
# set group_help_by_registry to false
$CL_PATH config group_help_by_registry false
# run command launcher to show the help message
RESULT=$($CL_PATH)
echo "$RESULT" | grep -q "Commands from 'dropin' registry"
if [ $? -eq 0 ]; then
  echo "KO - should group help message by registry"
  exit 1
else
  echo "OK"
fi
