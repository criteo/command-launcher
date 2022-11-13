#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR

SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

###
# test consent
###
# First copy the dropin packages for test
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/login $CL_HOME/dropins

$CL_PATH config user_consent_life 3s
$CL_PATH login -u test-user -p test-password

echo "> test consent disabled"
$CL_PATH config enable_user_consent false
RESULT=$($CL_PATH bonjour-consent)
echo "$RESULT"
echo "$RESULT" | grep -q "test-user"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should pass USERNAME resource to command"
  exit 1
fi

echo "$RESULT" | grep -q "test-password"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should pass PASSWORD resource to command"
  exit 1
fi

echo "> test consent enabled - user refused"
$CL_PATH config enable_user_consent true
RESULT=$(echo 'n' | $CL_PATH bonjour-consent)
echo "$RESULT"
echo "$RESULT" | grep -q "authorize the access?"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should request authorization"
  exit 1
fi

echo "$RESULT" | grep -q "test-user"
if [ $? -eq 0 ]; then
  echo "KO - should NOT pass USERNAME resource to command"
  exit 1
else
  echo "OK"
fi

echo "$RESULT" | grep -q "test-password"
if [ $? -eq 0 ]; then
  echo "KO - should NOT pass PASSWORD resource to command"
  exit 1
else
  echo "OK"
fi

echo "> test consent enabled - user authorized"
$CL_PATH config enable_user_consent true
RESULT=$(echo 'y' | $CL_PATH bonjour-consent)
echo "$RESULT"
echo "$RESULT" | grep -q "authorize the access?"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should request authorization"
  exit 1
fi

echo "$RESULT" | grep -q "test-user"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should pass USERNAME resource to command"
  exit 1
fi

echo "$RESULT" | grep -q "test-password"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should pass PASSWORD resource to command"
  exit 1
fi

echo "> test consent authorized - should not request authorization again"
$CL_PATH config enable_user_consent true
RESULT=$(echo 'y' | $CL_PATH bonjour-consent)
echo "$RESULT"
echo "$RESULT" | grep -q "authorize the access?"
if [ $? -eq 0 ]; then
  echo "KO - should NOT request authorization again"
  exit 1
else
  echo "OK"
fi

echo "> test consent authorized - should request authorization once expired"
sleep 5
$CL_PATH config enable_user_consent true
RESULT=$(echo 'y' | $CL_PATH bonjour-consent)
echo "$RESULT"
echo "$RESULT" | grep -q "authorize the access?"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should request authorization again once expired"
  exit 1
fi


