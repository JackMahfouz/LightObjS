#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Starting LightObjS Installation..."

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Go is not installed. Please install Go before running this script."
    exit 1
fi

# Get the directory of the script and change to the project root
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
cd "$SCRIPT_DIR"

echo "Running Go installation utility (install.go)..."
go run install.go

echo "LightObjS Installation Complete!"
echo "You can now start the application by running: go run main.go"
