#!/bin/bash

# Test: When updating an existing package fails, the pause mechanism should work
# This test covers the update failure case (as opposed to new installation failure)

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up
rm -rf $CL_HOME/dropins
rm -rf $CL_HOME/current
rm -f $CL_HOME/config.json
mkdir -p $CL_HOME/dropins

# Create a local test remote repository
TEST_REMOTE_DIR=$CL_HOME/test-remote
mkdir -p $TEST_REMOTE_DIR

# Copy a valid package (version 1.0.0) from test assets
cp $SCRIPT_DIR/../packages-src/bonjour/bonjour-v1.pkg $TEST_REMOTE_DIR/bonjour-1.0.0.pkg

# Calculate checksum for valid package
CHECKSUM_V1=$(shasum -a 256 $TEST_REMOTE_DIR/bonjour-1.0.0.pkg | cut -d' ' -f1)

# Create index.json with version 1.0.0 (valid package)
cat > $TEST_REMOTE_DIR/index.json << EOF
[
  {
    "name": "bonjour",
    "version": "1.0.0",
    "checksum": "$CHECKSUM_V1",
    "startPartition": 0,
    "endPartition": 9
  }
]
EOF

# Configure the command launcher (updates disabled initially)
$CL_PATH config command_repository_base_url "file://$TEST_REMOTE_DIR" > /dev/null 2>&1

################
echo "> test install valid package first"

# Enable updates and run to install the package
$CL_PATH config command_update_enabled true > /dev/null 2>&1
echo "* running first command (installing valid package 1.0.0)"
RESULT=$($CL_PATH 2>&1)

# Check that the bonjour command exists (package was installed)
echo "$RESULT" | grep -q "bonjour.*print bonjour"
if [ $? -eq 0 ]; then
  echo "OK - bonjour command is available (package installed)"
else
  echo "KO - bonjour command should be available"
  echo "Output: $RESULT"
  exit 1
fi

# Verify package directory exists
if [ -d "$CL_HOME/current/bonjour" ]; then
  echo "OK - package directory exists"
else
  echo "KO - package directory should exist"
  exit 1
fi

################
echo "> test failed update creates pause file"

# Now create a broken version 2.0.0 in the remote
echo "this is not a valid zip file" > $TEST_REMOTE_DIR/bonjour-2.0.0.pkg

# Update index.json to have version 2.0.0 (broken package)
cat > $TEST_REMOTE_DIR/index.json << EOF
[
  {
    "name": "bonjour",
    "version": "2.0.0",
    "checksum": "0000000000000000000000000000000000000000000000000000000000000000",
    "startPartition": 0,
    "endPartition": 9
  }
]
EOF

# Run command - triggers automatic update which should fail
echo "* running second command (triggers update to broken 2.0.0, expecting failure)"
RESULT=$($CL_PATH 2>&1)

# Check that update was attempted
echo "$RESULT" | grep -q "upgrade package 'bonjour' from version 1.0.0 to version 2.0.0"
if [ $? -eq 0 ]; then
  echo "OK - update was attempted"
else
  echo "KO - should have attempted to update bonjour"
  echo "Output: $RESULT"
  exit 1
fi

# Check that update failed
echo "$RESULT" | grep -q "Cannot .* the package bonjour"
if [ $? -eq 0 ]; then
  echo "OK - update failed as expected"
else
  echo "KO - should have failed to update bonjour"
  echo "Output: $RESULT"
  exit 1
fi

# Check if pause succeeded
echo "$RESULT" | grep -q "has been paused due to installation failure"
if [ $? -eq 0 ]; then
  echo "OK - package was paused after update failure"
else
  echo "$RESULT" | grep -q "Failed to pause update for package"
  if [ $? -eq 0 ]; then
    echo "KO - BUG: pause was attempted but failed"
    exit 1
  else
    echo "KO - BUG: pause mechanism was not triggered at all for failed update"
    exit 1
  fi
fi

# Run again - should skip the paused package (not retry update)
echo "* running third command (should skip paused package)"
RESULT=$($CL_PATH 2>&1)

# If pause works correctly, it should not try to update again
echo "$RESULT" | grep -q "upgrade package 'bonjour'"
if [ $? -ne 0 ]; then
  echo "OK - paused package was skipped"
else
  echo "KO - BUG: paused package should have been skipped but update was retried"
  exit 1
fi

# Verify the original package is still working (version 1.0.0 should still be installed)
echo "* verifying original package still works"
RESULT=$($CL_PATH bonjour 2>&1)
echo "$RESULT" | grep -qi "bonjour"
if [ $? -eq 0 ]; then
  echo "OK - original package still works"
else
  echo "KO - original package should still be working"
  echo "Output: $RESULT"
  exit 1
fi
