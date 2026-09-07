#!/bin/bash

# Test: ENABLE_LAZY_SETUP config (run the package __setup__ hook before a command) + .setup marker
#
# availeble environment varibale
# CL_PATH: the path of the command launcher binary
# CL_HOME: the path of the command launcher home directory
# OUTPUT_DIR: the output folder
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up the dropin folder and the configuration
rm -rf $CL_HOME/dropins
rm -f $CL_HOME/config.json
mkdir -p $CL_HOME/dropins

# copy the test packages to the dropin folder
cp -R $SCRIPT_DIR/../packages-src/lazy-setup $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/lazy-setup-fail $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/lazy-setup-nohook $CL_HOME/dropins
cp -R $SCRIPT_DIR/../packages-src/bonjour $CL_HOME/dropins

PKG=$CL_HOME/dropins/lazy-setup
PKG_FAIL=$CL_HOME/dropins/lazy-setup-fail
PKG_NOHOOK=$CL_HOME/dropins/lazy-setup-nohook

# helpers
ok() { echo "OK"; }
ko() { echo "KO - $1"; exit 1; }
# number of lines in the setup counter file (0 when absent)
setup_count() {
  if [ -f "$PKG/setup.count" ]; then
    wc -l < "$PKG/setup.count" | tr -d ' '
  else
    echo 0
  fi
}
expect_count() {
  COUNT=$(setup_count)
  if [ "$COUNT" == "$1" ]; then ok; else ko "setup hook should have run $1 time(s), got $COUNT"; fi
}

################
echo "> test lazy setup is disabled by default"
RESULT=$($CL_PATH lazy-hello)
EXIT_CODE=$?

echo "* command should succeed"
[ $EXIT_CODE -eq 0 ] && ok || ko "lazy-hello should exit 0"

echo "* command output should be printed"
echo "$RESULT" | grep -q "hello lazy" && ok || ko "should print hello lazy"

echo "* setup hook should NOT be called when enable_lazy_setup is false"
echo "$RESULT" | grep -q "calling lazy setup" && ko "setup hook should not be called" || ok

echo "* no setup marker should be written"
[ ! -f "$PKG/.setup" ] && ok || ko ".setup should not exist"
expect_count 0

################
echo "> test enable lazy setup"
$CL_PATH config enable_lazy_setup true
RESULT=$($CL_PATH config enable_lazy_setup)
echo "$RESULT" | grep -q "true" && ok || ko "enable_lazy_setup should be true"

################
echo "> test first command of the package runs setup"
RESULT=$($CL_PATH lazy-hello)
EXIT_CODE=$?

echo "* command should succeed"
[ $EXIT_CODE -eq 0 ] && ok || ko "lazy-hello should exit 0"

echo "* setup hook should be called before the command"
echo "$RESULT" | grep -q "calling lazy setup" && ok || ko "setup hook should be called"

echo "* command output should be printed"
echo "$RESULT" | grep -q "hello lazy" && ok || ko "should print hello lazy"

echo "* setup marker should be written in the package folder"
[ -f "$PKG/.setup" ] && ok || ko ".setup should exist"

echo "* setup marker should record the package version"
grep -q '"packageVersion": "1.0.0"' "$PKG/.setup" && ok || ko ".setup should contain the package version"

echo "* setup hook should have run once"
expect_count 1

################
echo "> test second command of the package does not re-run setup"
RESULT=$($CL_PATH lazy-hello)

echo "* setup hook should NOT be called again"
echo "$RESULT" | grep -q "calling lazy setup" && ko "setup hook should not be called again" || ok

echo "* command output should be printed"
echo "$RESULT" | grep -q "hello lazy" && ok || ko "should print hello lazy"

echo "* setup hook should still have run once"
expect_count 1

################
echo "> test removing the marker re-runs setup"
rm -f "$PKG/.setup"
RESULT=$($CL_PATH lazy-hello)

echo "* setup hook should be called again"
echo "$RESULT" | grep -q "calling lazy setup" && ok || ko "setup hook should be called after marker removal"

echo "* setup marker should be written again"
[ -f "$PKG/.setup" ] && ok || ko ".setup should exist"
expect_count 2

################
echo "> test failing lazy setup blocks the command"
RESULT=$($CL_PATH lazy-fail 2>&1)
EXIT_CODE=$?

echo "* command should fail"
[ $EXIT_CODE -ne 0 ] && ok || ko "lazy-fail should exit with a non-zero code"

echo "* error should mention the lazy setup failure"
echo "$RESULT" | grep -q "lazy setup failed" && ok || ko "should report lazy setup failure"

echo "* the command itself should NOT run"
echo "$RESULT" | grep -q "should not run" && ko "command should not run when setup fails" || ok

echo "* no setup marker should be written"
[ ! -f "$PKG_FAIL/.setup" ] && ok || ko ".setup should not exist after a failed setup"

################
echo "> test manual package setup writes the marker"
rm -f "$PKG/.setup"
RESULT=$($CL_PATH package setup lazy-setup)
EXIT_CODE=$?

echo "* package setup should succeed"
[ $EXIT_CODE -eq 0 ] && ok || ko "package setup lazy-setup should exit 0"

echo "* setup marker should be written"
[ -f "$PKG/.setup" ] && ok || ko ".setup should exist after package setup"
expect_count 3

echo "* the command should not re-run setup after a manual setup"
RESULT=$($CL_PATH lazy-hello)
echo "$RESULT" | grep -q "calling lazy setup" && ko "setup hook should not be called" || ok
expect_count 3

echo "* package setup of a failing hook should fail"
RESULT=$($CL_PATH package setup lazy-setup-fail 2>&1)
[ $? -ne 0 ] && ok || ko "package setup lazy-setup-fail should exit with a non-zero code"

################
echo "> test command in a package without setup hook"
RESULT=$($CL_PATH lazy-nohook)
EXIT_CODE=$?

echo "* command should succeed"
[ $EXIT_CODE -eq 0 ] && ok || ko "lazy-nohook should exit 0"

echo "* command output should be printed"
echo "$RESULT" | grep -q "hello nohook" && ok || ko "should print hello nohook"

echo "* no setup marker should be written for a package without hook"
[ ! -f "$PKG_NOHOOK/.setup" ] && ok || ko ".setup should not exist"

################
echo "> test package inspect shows the setup state"
RESULT=$($CL_PATH package inspect lazy-setup)
echo "* should show setup done"
echo "$RESULT" | grep -q "Setup:.*done" && ok || ko "lazy-setup should show Setup: done"

RESULT=$($CL_PATH package inspect lazy-setup-fail)
echo "* should show setup pending"
echo "$RESULT" | grep -q "Setup:.*pending" && ok || ko "lazy-setup-fail should show Setup: pending"

RESULT=$($CL_PATH package inspect bonjour)
echo "* should show setup n/a for a package without hook"
echo "$RESULT" | grep -q "Setup:.*n/a" && ok || ko "bonjour should show Setup: n/a"

################
echo "> test lazy setup is independent from the install-time setup hook config"
$CL_PATH config enable_package_setup_hook false
rm -f "$PKG/.setup"
RESULT=$($CL_PATH lazy-hello)

echo "* setup hook should be called even with enable_package_setup_hook false"
echo "$RESULT" | grep -q "calling lazy setup" && ok || ko "lazy setup should not depend on enable_package_setup_hook"
expect_count 4

################
echo "> test disabling lazy setup again"
$CL_PATH config enable_lazy_setup false
rm -f "$PKG/.setup"
RESULT=$($CL_PATH lazy-hello)

echo "* setup hook should NOT be called"
echo "$RESULT" | grep -q "calling lazy setup" && ko "setup hook should not be called when disabled" || ok
expect_count 4
