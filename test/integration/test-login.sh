#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR

SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

# First copy the dropin packages for test
rm -rf $CL_HOME/dropins
mkdir -p $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/login $CL_HOME/dropins


###
# login without -u
###
echo "Checking login without -u flag"
RESULT=$(echo | $CL_PATH login -p test-password) # The echo simulates an empty input from the user
echo -e "Actual:\n$RESULT"
EXPECTED_USERNAME_PROMPT="Please enter your user name [$(whoami)]: "
TEST_DESCRIPTION="should prompt '$EXPECTED_USERNAME_PROMPT'"
if ! grep -qF "$EXPECTED_USERNAME_PROMPT" <(echo "$RESULT"); then
    echo "KO - $TEST_DESCRIPTION"
    exit 1
else
    echo "OK - $TEST_DESCRIPTION"
fi

RESULT=$($CL_PATH print-credentials)
echo -e "Actual:\n$RESULT"
EXPECTED_USERNAME="CL_USERNAME: $(whoami)"
TEST_DESCRIPTION="should have username $(whoami)"
if ! grep -qF "$EXPECTED_USERNAME" <(echo "$RESULT"); then
    echo "KO - $TEST_DESCRIPTION"
    exit 1
else
    echo "OK - $TEST_DESCRIPTION"
fi


###
# login with -u
###
echo "Checking login with -u flag"
RESULT=$($CL_PATH login -u test-user -p test-password)
echo -e "Actual:\n$RESULT"
TEST_DESCRIPTION=" No output is expected"
if [[ "$RESULT" =~ ^[[:space:]]*$ ]]; then
    echo "OK - $TEST_DESCRIPTION"
else
    echo "KO - $TEST_DESCRIPTION"
    exit 1
fi

RESULT=$($CL_PATH print-credentials)
echo -e "Actual:\n$RESULT"
EXPECTED_USERNAME="CL_USERNAME: test-user"
TEST_DESCRIPTION="should have username test-user"
if ! grep -qF "$EXPECTED_USERNAME" <(echo "$RESULT"); then
    echo "KO - $TEST_DESCRIPTION"
    exit 1
else
    echo "OK - $TEST_DESCRIPTION"
fi
