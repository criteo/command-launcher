#!/bin/bash

# availeble environment varibale
# CL_PATH: the path of the command launcher binary
# CL_HOME: the path of the command launcher home directory
# OUTPUT_DIR: the output folder
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up the dropin folder and local folder
rm -rf $CL_HOME/dropins
rm -rf $CL_HOME/current
rm -f $CL_HOME/config.json
mkdir -p $CL_HOME/dropins

# copy the example to the dropin folder for the test
cp -R $SCRIPT_DIR/../packages-src/bonjour $CL_HOME/dropins

# download remote package
RESULT=$($OUTPUT_DIR/cl config command_repository_base_url https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo)
RESULT=$($OUTPUT_DIR/cl)

################
echo "> test list all packages"
RESULT=$($CL_PATH package list)

echo "* should contain managed packages section"
echo "$RESULT" | grep -q "=== Managed Repository: Default ==="
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have Default Managed Repository section"
  exit 1
fi

echo "* should contain 'command-launcher-demo' package as local package"
echo "$RESULT" | grep -q "\- command-launcher-demo"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain local package 'command-launcher-demo'"
  exit 1
fi

echo "* should contain dropin packages section"
echo "$RESULT" | grep -q "=== Dropin Repository ==="
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have Dropin Repository section"
  exit 1
fi

echo "* should contain 'bonjour' package as dropin"
echo "$RESULT" | grep -q "\- bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain dropin package 'bonjour'"
  exit 1
fi

################
echo "> test list --dropin command"
RESULT=$($CL_PATH package list --dropin)

echo "* should contain dropin packages section"
echo "$RESULT" | grep -q "=== Dropin Repository ==="
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have Dropin Repository section"
  exit 1
fi

echo "* should contain 'bonjour' package as dropin"
echo "$RESULT" | grep -q "\- bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain dropin package 'bonjour'"
  exit 1
fi

echo "* should NOT contain local packages section"
echo "$RESULT" | grep -q "=== Local Repository ==="
if [ $? -ne 0 ]; then
  echo "OK"
else
  echo "KO - should NOT have Local Repository section"
  exit 1
fi

################
echo "> test list --local command"
RESULT=$($CL_PATH package list --local)

echo "* should contain local packages section"
echo "$RESULT" | grep -q "=== Managed Repository: Default ==="
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should have Local Repository section"
  exit 1
fi

echo "* should contain 'command-launcher-demo' package as local package"
echo "$RESULT" | grep -q "\- command-launcher-demo"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain local package 'command-launcher-demo'"
  exit 1
fi

echo "* should NOT contain dropin packages section"
echo "$RESULT" | grep -q "=== Dropin Repository ==="
if [ $? -ne 0 ]; then
  echo "OK"
else
  echo "KO - should NOT have Dropin Repository section"
  exit 1
fi

################
echo "> test list local --include-cmd"
RESULT=$($CL_PATH package list --local --include-cmd)

echo "* should contain package with version"
echo "$RESULT" | grep -q "Package: command-launcher-demo (v1.0.0)"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain package with version"
  exit 1
fi

echo "* should contain group"
echo "$RESULT" | grep -q "\* __no_group__"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain __no_group__"
  exit 1
fi

echo "* should contain command"
echo "$RESULT" | grep -q "\- hello"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain command"
  exit 1
fi

################
echo "> test list dropin --include-cmd"
RESULT=$($CL_PATH package list --dropin --include-cmd)

echo "* should contain package version"
echo "$RESULT" | grep -q "Package: bonjour (v1.0.0)"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain package version"
  exit 1
fi

echo "* should contain group"
echo "$RESULT" | grep -q "\* __no_group__"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain __no_group__"
  exit 1
fi

echo "* should contain command"
echo "$RESULT" | grep -q "\- bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain command"
  exit 1
fi

################
echo "> test list remote"
RESULT=$($CL_PATH package list --remote)

echo "* should contain remote package and version"
echo "$RESULT" | grep -q "\- command-launcher-demo                              1.0.0"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain remote package and version"
  exit 1
fi

echo "* should contain remote section"
echo "$RESULT" | grep -q "=== Remote Registry: Default ==="
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain remote section"
  exit 1
fi

################
echo "> test inspect dropin package"
RESULT=$($CL_PATH package inspect bonjour)

echo "* should contain package name"
echo "$RESULT" | grep -q "Package: bonjour"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain package name"
  exit 1
fi

echo "* should contain version"
echo "$RESULT" | grep -q "Version:.*1.0.0"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain version"
  exit 1
fi

echo "* should show full name"
echo "$RESULT" | grep -q "Full Name:.*bonjour@dropin"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show full name bonjour@dropin"
  exit 1
fi

echo "* should show source as dropin"
echo "$RESULT" | grep -q "Source:.*dropin"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show source as dropin"
  exit 1
fi

echo "* should show managed as false"
echo "$RESULT" | grep -q "Managed:.*false"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show managed as false"
  exit 1
fi

echo "* should contain commands section"
echo "$RESULT" | grep -q "Commands:"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain commands section"
  exit 1
fi

################
echo "> test inspect managed package"
RESULT=$($CL_PATH package inspect command-launcher-demo)

echo "* should contain package name"
echo "$RESULT" | grep -q "Package: command-launcher-demo"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain package name"
  exit 1
fi

echo "* should show full name"
echo "$RESULT" | grep -q "Full Name:.*command-launcher-demo@default"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show full name command-launcher-demo@default"
  exit 1
fi

echo "* should show managed as true"
echo "$RESULT" | grep -q "Managed:.*true"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show managed as true"
  exit 1
fi

echo "* should show sync policy"
echo "$RESULT" | grep -q "Sync:"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should show sync policy"
  exit 1
fi

################
echo "> test inspect nonexistent package"
RESULT=$($CL_PATH package inspect nonexistent-pkg 2>&1)
EXIT_CODE=$?

echo "* should fail with error"
if [ $EXIT_CODE -ne 0 ]; then
  echo "OK"
else
  echo "KO - should fail for nonexistent package"
  exit 1
fi

echo "* should contain error message"
echo "$RESULT" | grep -q "no package named nonexistent-pkg found"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain error message"
  exit 1
fi

################
echo "> test install git package"
RESULT=$($CL_PATH package install --git https://github.com/criteo/command-launcher-package-example)
RESULT=$($CL_PATH package list --dropin --include-cmd)

echo "* should contain package from git repo"
echo "$RESULT" | grep -q "Package: command-launcher-example-package (v0.0.1)"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain package from git repo"
  exit 1
fi

echo "* should contain group command from git repo"
echo "$RESULT" | grep -q "\* cola-example"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain group command from git repo"
  exit 1
fi

echo "* should contain greeting command from git repo"
echo "$RESULT" | grep -q "\- greeting"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain greeting command from git repo"
  exit 1
fi

################
echo "> test delete dropin package"
RESULT=$($CL_PATH package delete command-launcher-example-package)
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - delete should exit 0"
  exit 1
fi

echo "* should NOT contain package from git repo"
RESULT=$($CL_PATH package list --dropin --include-cmd)
echo "$RESULT" | grep -q "Package: command-launcher-example-package (v0.0.1)"
if [ $? -ne 0 ]; then
  echo "OK"
else
  echo "KO - should NOT contain package from git repo"
  exit 1
fi

################
echo "> test install file package"
RESULT=$($CL_PATH package install --file https://github.com/criteo/command-launcher/raw/main/test/remote-repo/command-launcher-demo-2.0.0.pkg)

echo "* should contain 2.0.0 demo package"
RESULT=$($CL_PATH package list --dropin --include-cmd)
echo "$RESULT" | grep -q "Package: command-launcher-demo (v2.0.0)"
if [ $? -eq 0 ]; then
  echo "OK"
else
  echo "KO - should contain 2.0.0 demo package"
  exit 1
fi

###############
echo "> test setup package"
RESULT=$($CL_PATH package setup command-launcher-example-package)
echo "$RESULT" | grep -q "no setup hook found"
if [ $? -ne 0 ]; then
  echo "OK"
else
  echo "KO - should prompt no setup hook found"
  exit 1
fi
