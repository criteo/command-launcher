#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

##
# test yaml format manifest
##
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/yaml-manifest $CL_HOME/dropins

echo "> test YAML manifest without arguments in manifest"
RESULT=$($OUTPUT_DIR/cl bonjour1 world)
echo $RESULT | grep -q "bonjour! world"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of bonjour command: $RESULT"
  exit 1
fi

echo "> test YAML manifest with arguments in manifest"
RESULT=$($OUTPUT_DIR/cl bonjour2)
echo $RESULT | grep -q "bonjour! monde"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong output of bonjour command: $RESULT"
  exit 1
fi

echo "> test YAML manifest with long description"
RESULT=$($OUTPUT_DIR/cl help bonjour1)
echo $RESULT
echo $RESULT | grep -q "This is another line"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong format of the long description"
  exit 1
fi

echo "> test argsUsage and examples, when checkFlags=true, should have custom help message"
RESULT=$($OUTPUT_DIR/cl bonjour2 -h)
echo $RESULT
echo $RESULT | grep -q "bonjour2 name"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong format of custom help message"
  exit 1
fi

echo $RESULT | grep -q "# Print greeting message"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - wrong format of example message"
  exit 1
fi


