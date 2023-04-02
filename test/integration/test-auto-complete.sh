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
echo "$RESULT" | grep -q "print bonjour from command launcher"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete registred command"
  exit 1
fi

echo "> test static argument auto-complete"
RESULT=$($CL_PATH __complete bonjour "")
echo "$RESULT" | grep -q "Joe"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete static arguments"
  exit 1
fi

echo "> test dynamic argument auto-complete"
RESULT=$($CL_PATH __complete greeting saybonjour "")
echo "$RESULT" | grep -q "John"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete dynamic arguments"
  exit 1
fi
echo "$RESULT" | grep -q "Kate"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete dynamic arguments"
  exit 1
fi

echo "> test flag name auto-complete"
RESULT=$($CL_PATH __complete bonjour -)
echo "$RESULT" | grep -q "\-\-lang"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete flag names"
  exit 1
fi
echo "$RESULT" | grep -q "\-l"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete short flag names"
  exit 1
fi

echo "> test flag static value auto-complete"
RESULT=$($CL_PATH __complete bonjour --lang "")
echo "$RESULT" | grep -q "fr"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete static flag values"
  exit 1
fi
echo "$RESULT" | grep -q "jp"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete static flag values"
  exit 1
fi

echo "> test flag dynamic value auto-complete"
RESULT=$($CL_PATH __complete bonjour bo --lang fr --name "")
echo "$RESULT" | grep -q "bo --lang fr"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should pass original arguments to the completion command"
  exit 1
fi
echo "$RESULT" | grep -q "John"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete dynamic flag values"
  exit 1
fi
echo "$RESULT" | grep -q "Kate"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete dynamic flag values"
  exit 1
fi

echo "> test delete command auto completion"
RESULT=$($CL_PATH __complete package delete " ")
echo "$RESULT" | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should auto-complete package names in package delete command"
  exit 1
fi

