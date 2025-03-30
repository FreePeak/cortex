# Cortex - Flexible MCP Server Platform

## Project Overview

The Cortex project provides a flexible Model Context Protocol (MCP) server platform that supports dynamic tool registration and multiple communication protocols. The platform is designed with a clean separation between the core server infrastructure and the specific tool implementations.

## Key Features

### Plugin Architecture

- **Provider Interface**: A standardized interface for tool providers
- **Registry System**: Dynamic registration and discovery of tools
- **Base Provider**: A foundation for building custom tool providers
- **Flexible Server**: Support for both stdio and HTTP protocols

### Decoupled Design

- **Clean Separation**: Core platform is separate from specific tool implementations
- **External Tool Providers**: Tools can be implemented outside the core
- **Dynamic Loading**: Tools can be registered at runtime
- **Versioned Interfaces**: Support for evolving tool interfaces

### Communication Support

- **Stdio Protocol**: Command-line interface via standard input/output
- **HTTP Protocol**: Web-based interface via Server-Sent Events (SSE)
- **Unified Message Handling**: Consistent API across protocols
- **Graceful Shutdown**: Proper handling of termination signals

### Security Features

- **Input Validation**: Validation of all tool parameters
- **Error Handling**: Consistent error reporting and logging
- **Provider Isolation**: Tools run in isolated provider contexts
- **Session Management**: Client session tracking and management

## Implementation

The implementation includes the following components:

### Core Platform (`pkg/plugin`, `pkg/server`)

- **Plugin Interface**: Defines the contract for tool providers
- **Plugin Registry**: Manages provider registration and discovery
- **Base Provider**: Provides a foundation for building providers
- **Flexible Server**: Coordinates tools and communication

### Example Providers (`examples/providers`)

- **Weather Provider**: Demonstrates integration with a weather service
- **Database Provider**: Shows implementation of database operations

### Example Server (`cmd/flexible-server`)

- **Multi-Protocol Server**: Supports both stdio and HTTP
- **Dynamic Tool Integration**: Demonstrates provider registration
- **Command-Line Options**: Configuration for different operation modes
- **Error Handling**: Robust error handling and logging

## Usage

### Running the Server

```bash
# Run with stdio protocol
./run-flexible-stdio.sh

# Run with HTTP protocol
./run-flexible-http.sh
```

### Creating Custom Providers

See the documentation in `pkg/plugin/README.md` and examples in `examples/providers/` for details on how to create custom tool providers.

## Documentation

The project includes comprehensive documentation:

- **READMEs**: Overview and usage instructions in each directory
- **Code Comments**: Detailed comments for all public interfaces
- **Example Code**: Working examples of providers and server
- **Command-Line Help**: Documentation for command-line options

## Future Extensions

The platform design allows for several future extensions:

- **Dynamic Loading**: Support for loading providers from shared libraries
- **Remote Providers**: Distributed providers running on different machines
- **Authentication**: Provider and tool authentication system
- **Monitoring**: Metrics and monitoring for tool executions
- **Admin Interface**: Web-based administration console

## Conclusion

The Cortex platform provides a flexible and extensible foundation for building MCP servers that can interact with various backend services through a standardized protocol. The clean separation between the core platform and specific tool implementations allows for independent evolution and ensures maintainability and scalability. 