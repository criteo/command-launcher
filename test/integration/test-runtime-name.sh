#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

##
# test binary name becomes app name
##
echo "> test binary name is used as app name"
RESULT=$($CL_PATH version)
echo "$RESULT" | grep -q "^cl version"
if [ $? -ne 0 ]; then
  echo "KO - expected app name 'cl' from binary name"
  exit 1
else
  echo "OK"
fi

##
# test copied binary gets its own name
##
echo "> test copied binary gets its own name"
COPIED=$OUTPUT_DIR/myapp
cp $CL_PATH $COPIED
export MYAPP_HOME=$OUTPUT_DIR/myapp-home
mkdir -p $MYAPP_HOME

RESULT=$($COPIED version)
echo "$RESULT" | grep -q "^myapp version"
if [ $? -ne 0 ]; then
  echo "KO - expected app name 'myapp' from copied binary"
  rm -f $COPIED
  rm -rf $MYAPP_HOME
  exit 1
else
  echo "OK"
fi

##
# test symlink resolves to original name
##
echo "> test symlink resolves to original binary name"
LINK=$OUTPUT_DIR/myalias
ln -sf $CL_PATH $LINK

RESULT=$($LINK version)
echo "$RESULT" | grep -q "^cl version"
if [ $? -ne 0 ]; then
  echo "KO - symlink should resolve to original name 'cl'"
  rm -f $LINK $COPIED
  rm -rf $MYAPP_HOME
  exit 1
else
  echo "OK"
fi

rm -f $LINK

##
# test long name from config
##
echo "> test default long name from compiled-in value"
RESULT=$($COPIED)
echo "$RESULT" | grep -q "Command Launcher - A command launcher"
if [ $? -ne 0 ]; then
  echo "KO - expected compiled-in long name as default"
  rm -f $COPIED
  rm -rf $MYAPP_HOME
  exit 1
else
  echo "OK"
fi

echo "> test long name override from config"
$COPIED config app_long_name "My Custom App"

RESULT=$($COPIED)
echo "$RESULT" | grep -q "My Custom App - A command launcher"
if [ $? -ne 0 ]; then
  echo "KO - expected long name from config"
  rm -f $COPIED
  rm -rf $MYAPP_HOME
  exit 1
else
  echo "OK"
fi

##
# test original binary is unaffected by copy's config
##
echo "> test original binary unaffected by copy's config"
RESULT=$($CL_PATH)
echo "$RESULT" | grep -q "Command Launcher - A command launcher"
if [ $? -ne 0 ]; then
  echo "KO - original should still use compiled-in long name"
  rm -f $COPIED
  rm -rf $MYAPP_HOME
  exit 1
else
  echo "OK"
fi

# cleanup
rm -f $COPIED
rm -rf $MYAPP_HOME
