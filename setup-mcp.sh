#!/bin/bash

# Setup script for MCP configuration for Cursor

# Directory containing this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Build the MCP server in stdio mode
echo "Building MCP server..."
go build -o "$SCRIPT_DIR/mcp-stdio" "$SCRIPT_DIR/cmd/stdio/main.go"
if [ $? -ne 0 ]; then
    echo "Error building MCP server"
    exit 1
fi

# Get absolute path to the executable
MCP_PATH=$(cd "$SCRIPT_DIR" && echo "$(pwd)/mcp-stdio")

# Generate MCP configuration file
echo "Generating MCP configuration..."
cat > "$SCRIPT_DIR/mcp.example.json" << EOL
{
  "mcp.executablePath": "${MCP_PATH}",
  "mcp.transport": "stdio",
  "mcp.logging": true,
  "mcp.capabilities": {
    "resources": true,
    "tools": true,
    "prompts": true
  }
}
EOL

echo "MCP configuration generated successfully!"
echo "Copy the configuration from $SCRIPT_DIR/mcp.example.json to your Cursor settings or create a .cursor folder with settings.json in your project root."
echo "To run the server manually: $MCP_PATH"
chmod +x "$SCRIPT_DIR/mcp-stdio"

# Optional: Create an example .cursor/settings.json file
mkdir -p "$SCRIPT_DIR/.cursor"
cat > "$SCRIPT_DIR/.cursor/settings.json" << EOL
{
  "mcp.executablePath": "${MCP_PATH}",
  "mcp.transport": "stdio",
  "mcp.logging": true
}
EOL

echo "Created example .cursor/settings.json for this project."
echo "When you open this project in Cursor, it should automatically use the MCP server." 