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

# run command launcher
echo "> integration test - template"
RESULT=$($CL_PATH bonjour)

# check result or exit code, or any thing relevant to the test
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should return 0 when command succeeds"
  exit 1
fi





