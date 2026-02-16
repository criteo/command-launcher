#!/usr/bin/env sh

DEBUG=0

# Default values
DEFAULT_APP_NAME="cdt"
DEFAULT_APP_LONG_NAME="Criteo Dev Toolkit"

default_version() {
    echo "$(git rev-parse --abbrev-ref HEAD)-dev"
}

default_build_number() {
    date +'%Y%m%d-%H%M%S'
}

debug() {
    if [ $DEBUG -eq 1 ]; then
        echo "$1"
    fi
}

# Function to show usage
show_usage() {
    program=$(basename "$0")
    cat <<EOF
Usage:
    $program [OPTIONS]
    OR
    $program VERSION APP_NAME APP_LONG_NAME

Options:
    -v, --version VERSION      Set version
    -n, --name NAME            Set app name
    -l, --long-name LONG_NAME  Set app long name
    -o, --output OUTPUT        Set output file name (default: APP_NAME)
    -b, --build-number NUMBER  Set build number (default: timestamp)
    -d, --debug                Enable debug output
    -h, --help                 Show this help message

Examples:
    $program                                   # Use all defaults
    $program 1.0.0                             # Set version only (legacy)
    $program 1.0.0 myapp 'My App'              # Set all three (legacy)
    $program -v 1.0.0                          # Set version only
    $program -n myapp                          # Set app name only
    $program -l 'My App'                       # Set app long name only
    $program -v 1.0.0 -n myapp -l 'My App'    # Set all three
    $program -b 42                             # Set build number
EOF
}

validate_argument() {
    if [ -z "$2" ] || [ "${2#-}" = "$2" ]; then
        return 0
    else
        echo "Error: $1 requires a value" >&2
        exit 1
    fi
}

# Parse options
VERSION=
APP_NAME=
APP_LONG_NAME=
OUTPUT=
BUILD_NUMBER=
POSITIONAL_COUNT=0

while [ $# -gt 0 ]; do
    case "$1" in
        -v|--version)
            validate_argument "$1" "$2"
            VERSION="$2"
            shift 2
            ;;
        -n|--name)
            validate_argument "$1" "$2"
            APP_NAME="$2"
            shift 2
            ;;
        -l|--long-name)
            validate_argument "$1" "$2"
            APP_LONG_NAME="$2"
            shift 2
            ;;
        -o|--output)
            validate_argument "$1" "$2"
            OUTPUT="$2"
            shift 2
            ;;
        -b|--build-number)
            validate_argument "$1" "$2"
            BUILD_NUMBER="$2"
            shift 2
            ;;
        -d|--debug)
            DEBUG=1
            shift 1
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        -*)
            echo "Error: Unknown option $1" >&2
            show_usage
            exit 1
            ;;
        *)
            # Collect positional arguments for legacy mode
            POSITIONAL_COUNT=$((POSITIONAL_COUNT + 1))
            case $POSITIONAL_COUNT in
                1) VERSION="$1" ;;
                2) APP_NAME="$1" ;;
                3) APP_LONG_NAME="$1" ;;
                *)
                    echo "Error: Too many positional arguments" >&2
                    show_usage
                    exit 1
                    ;;
            esac
            shift
            ;;
    esac
done

# Derive the long name from the app name if not set
if [ -z "$APP_LONG_NAME" ] && [ -n "$APP_NAME" ]; then
    APP_LONG_NAME="$APP_NAME Command Launcher"
fi

# Set defaults for missing arguments
[ -z "$VERSION" ] && VERSION=$(default_version)
[ -z "$APP_NAME" ] && APP_NAME=$DEFAULT_APP_NAME
[ -z "$APP_LONG_NAME" ] && APP_LONG_NAME=$DEFAULT_APP_LONG_NAME
[ -z "$OUTPUT" ] && OUTPUT=$APP_NAME
[ -z "$BUILD_NUMBER" ] && BUILD_NUMBER=$(default_build_number)

LD_FLAGS=""
LD_FLAGS="$LD_FLAGS -X main.version=${VERSION}"
LD_FLAGS="$LD_FLAGS -X main.buildNum=${BUILD_NUMBER}"
LD_FLAGS="$LD_FLAGS -X main.appName=${APP_NAME}"
LD_FLAGS="$LD_FLAGS -X 'main.appLongName=${APP_LONG_NAME}'"

debug "Building $APP_NAME"
debug "Version: $VERSION"
debug "Build number: $BUILD_NUMBER"
debug "App name: $APP_NAME"
debug "App long name: $APP_LONG_NAME"
debug "Output: $OUTPUT"

go build -o "$OUTPUT" -ldflags="${LD_FLAGS}"
