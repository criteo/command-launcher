#!/bin/bash
set -eu

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REMOTE_DIR=$CL_HOME/test-update-stderr-remote
mkdir -p "$REMOTE_DIR"
cp "$SCRIPT_DIR/../packages-src/bonjour/bonjour-v1.pkg" "$REMOTE_DIR/bonjour-1.0.0.pkg"

cat > "$REMOTE_DIR/index.json" <<JSON
[{"name":"bonjour","version":"1.0.0","startPartition":0,"endPartition":9}]
JSON
"$CL_PATH" config command_repository_base_url "file://$REMOTE_DIR"
"$CL_PATH" config self_update_enabled false
"$CL_PATH" config command_update_enabled false
"$CL_PATH" bonjour > "$CL_HOME/expected-stdout"
"$CL_PATH" config command_update_enabled true

# A different advertised version triggers an update using the valid fixture.
cat > "$REMOTE_DIR/index.json" <<JSON
[
 {"name":"bonjour","version":"1.0.0","startPartition":0,"endPartition":9},
 {"name":"bonjour","version":"2.0.0","url":"file://$REMOTE_DIR/bonjour-1.0.0.pkg","startPartition":0,"endPartition":9}
]
JSON
"$CL_PATH" bonjour > "$CL_HOME/stdout" 2> "$CL_HOME/stderr"
cmp "$CL_HOME/expected-stdout" "$CL_HOME/stdout"
grep -q "Some commands require update" "$CL_HOME/stderr"
grep -q "upgrade package 'bonjour'" "$CL_HOME/stderr"
grep -q "installed successfully" "$CL_HOME/stderr"
grep -q "Update done!" "$CL_HOME/stderr"

# A download failure must also leave command results intact.
rm "$REMOTE_DIR/bonjour-1.0.0.pkg"
"$CL_PATH" bonjour > "$CL_HOME/stdout" 2> "$CL_HOME/stderr"
cmp "$CL_HOME/expected-stdout" "$CL_HOME/stdout"
grep -q "Cannot get the package bonjour" "$CL_HOME/stderr"
echo "OK - successful and failed updates keep command stdout clean"
