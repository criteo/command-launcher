#!/bin/bash

# required environment variable
# CL_PATH
# CL_HOME
# OUTPUT_DIR

SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

###
# Setup workspace project structure
###
WORKSPACE_DIR=$OUTPUT_DIR/workspace-project
mkdir -p $WORKSPACE_DIR/src
cp -R $SCRIPT_DIR/../packages-src/workspace-tool $WORKSPACE_DIR/workspace-tool

# Create .cl-packages file (binary name is "cl" in integration tests)
echo "workspace-tool" > $WORKSPACE_DIR/.cl-packages

# Enable workspace packages with short consent life for testing
$CL_PATH config ENABLE_WORKSPACE_PACKAGES true
$CL_PATH config user_consent_life 3s

# Wait for any stale consent from previous test runs to expire
sleep 4

###
# Test: workspace command appears in help (loaded without consent)
###
echo "> test workspace command appears in help"
RESULT=$(cd $WORKSPACE_DIR/src && $CL_PATH --help)
echo "$RESULT" | grep -q "ws-hello"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - workspace command ws-hello should appear in help"
  exit 1
fi

###
# Test: workspace command appears in autocompletion
###
echo "> test workspace command appears in autocompletion"
RESULT=$(cd $WORKSPACE_DIR/src && $CL_PATH __complete ws-hel)
echo "$RESULT" | grep -q "ws-hello"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - workspace command ws-hello should appear in autocompletion"
  exit 1
fi

###
# Test: workspace consent - user accepts
###
echo "> test workspace consent - user accepts"
RESULT=$(cd $WORKSPACE_DIR/src && echo 'y' | $CL_PATH ws-hello 2>&1)
echo "$RESULT"
echo "$RESULT" | grep -q "This command is provided by workspace"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show workspace consent prompt"
  exit 1
fi

echo "$RESULT" | grep -q "hello from workspace"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should execute workspace command when accepted"
  exit 1
fi

###
# Test: workspace consent remembered - no prompt on second run
###
echo "> test workspace consent remembered - no prompt on second run"
RESULT=$(cd $WORKSPACE_DIR/src && $CL_PATH ws-hello 2>&1)
echo "$RESULT"
echo "$RESULT" | grep -q "Do you trust"
if [ $? -eq 0 ]; then
  echo "KO - should NOT prompt for consent again"
  exit 1
else
  echo "OK"
fi

echo "$RESULT" | grep -q "hello from workspace"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should execute workspace command without re-prompting"
  exit 1
fi

###
# Test: workspace consent - user refuses (use a separate workspace)
# Wait for any stale consent/denial from previous test runs to expire
###
echo "> test workspace consent - user refuses"
DENY_WORKSPACE_DIR=$OUTPUT_DIR/deny-workspace
mkdir -p $DENY_WORKSPACE_DIR/src
cp -R $SCRIPT_DIR/../packages-src/workspace-tool $DENY_WORKSPACE_DIR/workspace-tool
echo "workspace-tool" > $DENY_WORKSPACE_DIR/.cl-packages

sleep 4

RESULT=$(cd $DENY_WORKSPACE_DIR/src && echo 'n' | $CL_PATH ws-hello 2>&1)
echo "$RESULT"
echo "$RESULT" | grep -q "This command is provided by workspace"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show workspace consent prompt"
  exit 1
fi

echo "$RESULT" | grep -q "execution denied"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show execution denied message"
  exit 1
fi

echo "$RESULT" | grep -q "hello from workspace"
if [ $? -eq 0 ]; then
  echo "KO - should NOT execute workspace command when refused"
  exit 1
else
  echo "OK"
fi

###
# Test: after denial, workspace command no longer appears
###
echo "> test denied workspace command no longer appears"
RESULT=$(cd $DENY_WORKSPACE_DIR/src && $CL_PATH --help)
echo "$RESULT" | grep -q "ws-hello"
if [ $? -eq 0 ]; then
  echo "KO - workspace command should NOT appear after denial"
  exit 1
else
  echo "OK"
fi

###
# Test: workspace command not visible when feature disabled
###
echo "> test workspace command not visible when feature disabled"
$CL_PATH config ENABLE_WORKSPACE_PACKAGES false
RESULT=$(cd $WORKSPACE_DIR/src && $CL_PATH --help)
echo "$RESULT" | grep -q "ws-hello"
if [ $? -eq 0 ]; then
  echo "KO - workspace command should NOT appear when feature disabled"
  exit 1
else
  echo "OK"
fi

###
# Test: workspace command not visible outside workspace
###
echo "> test workspace command not visible outside workspace"
$CL_PATH config ENABLE_WORKSPACE_PACKAGES true
RESULT=$(cd $OUTPUT_DIR && $CL_PATH --help)
echo "$RESULT" | grep -q "ws-hello"
if [ $? -eq 0 ]; then
  echo "KO - workspace command should NOT appear outside workspace"
  exit 1
else
  echo "OK"
fi

###
# Test: .cl-packages with parent traversal rejected
###
echo "> test parent traversal rejected"
SAFE_DIR=$OUTPUT_DIR/safe-project
mkdir -p $SAFE_DIR
echo "../workspace-project/workspace-tool" > $SAFE_DIR/.cl-packages
RESULT=$(cd $SAFE_DIR && $CL_PATH --help)
echo "$RESULT" | grep -q "ws-hello"
if [ $? -eq 0 ]; then
  echo "KO - parent traversal paths should be rejected"
  exit 1
else
  echo "OK"
fi
