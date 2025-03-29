#!/bin/bash

# Script to update import paths from github.com/FreePeak/cortex to github.com/FreePeak/cortex

# Run from project root
find . -name "*.go" -type f -exec sed -i '' 's|github.com/FreePeak/cortex|github.com/FreePeak/cortex|g' {} \;

echo "Import paths updated in all Go files" 