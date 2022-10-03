#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

##
# test exit code
##
# First copy the dropin packages for test
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/exit-code $CL_HOME/dropins

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

