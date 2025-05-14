# Cortex Embedding Implementation Plan

This document provides a detailed plan for implementing embedding support in Cortex. Each task is designed to be small and independently verifiable, following the principles of test-driven development.

## Task 1: Create Embeddable Interface

**Description**: Define an `Embeddable` interface that provides the core integration points for embedding Cortex in other applications.

**Steps**:
1. Create the interface in `pkg/server/embeddable.go`
2. Define methods for creating handlers, adding tools, and accessing server information
3. Write tests for the interface
4. Update the MCPServer to implement this interface

**Acceptance Criteria**:
- Interface defined with clear method signatures
- Tests pass for the interface implementation
- Documentation provided for the interface

## Task 2: Implement HTTP Handler Adapter

**Description**: Create an adapter that exposes Cortex functionality as an HTTP handler that can be used with any Go HTTP server.

**Steps**:
1. Create `pkg/server/http_adapter.go`
2. Implement `ToHTTPHandler()` method that converts an Embeddable to an http.Handler
3. Implement SSE (Server-Sent Events) support
4. Write tests for the handler adapter
5. Create example usage in the documentation

**Acceptance Criteria**:
- HTTP handler properly handles MCP requests
- SSE events are properly emitted
- Tests pass for the handler implementation
- Documentation provides clear usage examples

## Task 3: Create Standard Middleware Adapters

**Description**: Implement middleware adapters for common Go web frameworks to simplify integration.

**Steps**:
1. Create `pkg/server/middleware.go`
2. Implement adapter for standard net/http middleware
3. Write tests for the middleware adapter
4. Document usage examples

**Acceptance Criteria**:
- Middleware adapter correctly passes requests to Cortex
- Tests pass for the middleware implementation
- Documentation provides clear usage examples

## Task 4: Implement PocketBase Plugin

**Description**: Create a plugin for PocketBase that exposes Cortex functionality.

**Steps**:
1. Create `pkg/integration/pocketbase/plugin.go`
2. Implement the PocketBase Plugin interface
3. Create a factory function for creating the plugin
4. Write tests for the plugin
5. Document how to use the plugin

**Acceptance Criteria**:
- Plugin correctly registers with PocketBase
- Cortex tools and resources are accessible through PocketBase
- Tests pass for the plugin implementation
- Documentation provides clear usage instructions

## Task 5: Add Basic Embedding Examples

**Description**: Create examples of embedding Cortex in different server types.

**Steps**:
1. Create example for embedding in standard Go HTTP server
2. Create example for embedding in a PocketBase application
3. Ensure examples are well-documented
4. Test examples to verify functionality

**Acceptance Criteria**:
- Examples work as expected
- Documentation clearly explains the examples
- Examples are simple enough to serve as starting points

## Task 6: Implement Authentication Support

**Description**: Add support for using the host application's authentication in Cortex.

**Steps**:
1. Create `pkg/server/auth.go`
2. Implement authentication adapter interface
3. Create default implementation that uses HTTP headers
4. Write tests for authentication support
5. Document authentication integration

**Acceptance Criteria**:
- Authentication adapter works with host application credentials
- Tests pass for authentication implementation
- Documentation provides clear integration examples

## Task 7: Create Comprehensive Documentation

**Description**: Create comprehensive documentation for embedding Cortex.

**Steps**:
1. Update main README to mention embedding support
2. Create detailed guide for PocketBase integration
3. Add API documentation for embedding interfaces
4. Include troubleshooting section
5. Create tutorial for embedding in a simple application

**Acceptance Criteria**:
- Documentation is clear and comprehensive
- Examples demonstrate real-world use cases
- API documentation is complete and accurate

## Implementation Order

These tasks should be implemented in the following order:

1. Task 1: Create Embeddable Interface
2. Task 2: Implement HTTP Handler Adapter
3. Task 3: Create Standard Middleware Adapters
4. Task 5: Add Basic Embedding Examples
5. Task 4: Implement PocketBase Plugin
6. Task 6: Implement Authentication Support
7. Task 7: Create Comprehensive Documentation

This ordering allows each task to build on the previous ones, with early tasks providing the foundation for later ones. 