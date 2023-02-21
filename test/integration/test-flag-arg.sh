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
echo "$RESULT" | grep -q "bonjour!"
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

# test COLA environment variable
echo "> test COLA environment variable"
echo $RESULT | grep -q "cola flag: Joe"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - no environment variable COLA_FLAG_NAME found"
  exit 1
fi

echo $RESULT | grep -q "cola flag: French"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - no environment variable COLA_FLAG_LANGUAGE found"
  exit 1
fi

echo $RESULT | grep -q "cola arg: world"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - no environment variable COLA_ARG_1 found"
  exit 1
fi


echo "> test required flags error"
RESULT=$($CL_PATH nihao World 2>&1)
echo $RESULT
echo "$RESULT" | grep -q "Error: required flag(s) \"isolated-required\", \"name\" not set"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have error"
  exit 1
fi

echo "> test exclusive flags error"
RESULT=$($CL_PATH nihao --name Joe --language fr World --isolated-required value --json --text 2>&1)
echo "$RESULT" | grep -q "Error: if any flags in the group \[text json\] are set none of the others can be; \[json text\] were all set"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have error"
  exit 1
fi

echo "> test group flags error"
RESULT=$($CL_PATH nihao --name Joe World --isolated-required value  2>&1)
echo "$RESULT" | grep -q "Error: if any flags in the group \[name language\] are set they must all be set; missing \[language\]"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have error"
  exit 1
fi

echo "> test flags"
RESULT=$($CL_PATH nihao --name Joe --language fr World --json --isolated-required value 2>&1)
echo "$RESULT" | grep -q "\-\-json \-\-language fr \-\-name Joe World"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should pass original arguments to the scripts"
  exit 1
fi


echo "> test flags - show contain number of arguments"
echo "$RESULT" | grep -q "cola nargs: 1"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - show contain number of arguments"
  exit 1
fi
