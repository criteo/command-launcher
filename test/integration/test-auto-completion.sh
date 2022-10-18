#!/bin/bash

# availeble environment varibale
# CL_PATH: the path of the command launcher binary
# CL_HOME: the path of the command launcher home directory
# OUTPUT_DIR: the output folder
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up the dropin folder
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins

# copy the example to the dropin folder for the test
cp -R $SCRIPT_DIR/../packages-src/bonjour $CL_HOME/dropins

echo "> test registred command auto completion"
RESULT=$($CL_PATH __complete bon)

# check result or exit code, or any thing relevant to the test
echo $RESULT | grep -q "print bonjour from command launcher"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete registred command"
  exit 1
fi

echo "> test package update command auto completion"
RESULT=$($CL_PATH __complete package update " ")
echo $RESULT | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete package names in package update command"
  exit 1
fi

echo "> test package delete command auto completion"
RESULT=$($CL_PATH __complete package delete " ")
echo $RESULT | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete package names in package delete command"
  exit 1
fi

echo "> test package install command auto completion"
RESULT=$($CL_PATH __complete package install " ")
echo $RESULT | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete package names in package install command"
  exit 1
fi



