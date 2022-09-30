#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

###
# Test flag and arg environments (available when checkFlags: true)
###
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/flag-env $CL_HOME/dropins

echo "> test the flag and arg environment"
RESULT=$($OUTPUT_DIR/cl bonjour --name Joe --language French world)
echo $RESULT

echo $RESULT | grep -q "bonjour!"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of hello command: $RESULT"
  exit 1
fi

echo $RESULT | grep -q "Joe"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - no environment variable CL_FLAG_NAME found"
  exit 1
fi

echo $RESULT | grep -q "French"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - no environment variable CL_FLAG_LANGUAGE found"
  exit 1
fi

echo $RESULT | grep -q "world"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - no environment variable CL_ARG_1 found"
  exit 1
fi


