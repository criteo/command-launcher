#!/bin/bash

# availeble environment varibale
# CL_PATH: the path of the command launcher binary
# CL_HOME: the path of the command launcher home directory
# OUTPUT_DIR: the output folder
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up the dropin folder
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins

echo "> test download default remote command"
RESULT=$($OUTPUT_DIR/cl config command_repository_base_url https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo)
RESULT=$($OUTPUT_DIR/cl)

echo "* should have hello command installed"
echo "$RESULT" | grep -q "hello"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - hello command should exist"
  exit 1
fi

echo "* should contain default remote registry"
RESULT=$($CL_PATH remote list)
echo "$RESULT" | grep -q "default : https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should contain default remote registry"
  exit 1
fi


echo "> test add extra remote registry"
RESULT=$($CL_PATH remote add extra1 https://raw.githubusercontent.com/criteo/command-launcher/main/test/remote-repo)
RESULT=$($CL_PATH remote list)

echo "* should contain default remote registry"
echo "$RESULT" | grep -q "default : https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should contain default remote registry"
  exit 1
fi

echo "* should contain extra remote registry"
echo "$RESULT" | grep -q "extra1 : https://raw.githubusercontent.com/criteo/command-launcher/main/test/remote-repo"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should contain extra remote registry"
  exit 1
fi

echo "* should contain extra command: 'bonjour'"
RESULT=$($CL_PATH)
echo "$RESULT" | grep -q "bonjour"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should contain extra command 'bonjour'"
  exit 1
fi

echo "* should contain auto-renamed command: 'hello@@command-launcher-demo@extra1'"
echo "$RESULT" | grep -q "hello@@command-launcher-demo@extra1"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should contain auto-renamed command 'hello@@command-launcher-demo@extra1'"
  exit 1
fi

echo "* should be able to run 'hello@@command-launcher-demo@extra1'"
RESULT=$($CL_PATH hello@@command-launcher-demo@extra1)
echo "$RESULT" | grep -q "Hello World v2!"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should successfully run command 'hello@@command-launcher-demo@extra1'"
  exit 1
fi

echo "* should be able to run 'hello'"
RESULT=$($CL_PATH hello)
echo "$RESULT" | grep -q "Hello World!"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should successfully run command 'hello'"
  exit 1
fi

echo "> test delete extra remote registry"
RESULT=$($CL_PATH remote delete extra1)
RESULT=$($CL_PATH remote list)
echo "$RESULT" | grep -q "default : https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo"
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - should contain default remote registry"
  exit 1
fi

echo "* should NOT contain default remote registry"
echo "$RESULT" | grep -q "extra1 : https://raw.githubusercontent.com/criteo/command-launcher/main/test/remote-repo"
if [ $? -eq 0 ]; then
  echo "KO - should NOT contain extra remote registry"
  exit 1
else
  echo "OK"
fi

echo "* should NOT contain extra command"
RESULT=$($CL_PATH)
echo "$RESULT" | grep -q "bonjour"
if [ $? -eq 0 ]; then
  echo "KO - should NOT contain extra command 'bonjour'"
  exit 1
else
  echo "OK"
fi


