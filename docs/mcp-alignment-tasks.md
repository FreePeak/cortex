# MCP 2025-03-26 Alignment Tasks

This document outlines the tasks needed to align Cortex with the latest Model Context Protocol (MCP) specification (2025-03-26). The recent MCP update introduces several significant changes that require implementation in our codebase.

## Major Changes in MCP 2025-03-26

The latest MCP specification includes the following key updates:

1. **Authorization Framework**: A comprehensive authorization framework based on OAuth 2.1
2. **Streamable HTTP Transport**: Replacement of HTTP+SSE transport with a more flexible Streamable HTTP transport
3. **JSON-RPC Batching**: Support for batched requests via JSON-RPC
4. **Tool Annotations**: Comprehensive tool annotations for better describing tool behavior

## Implementation Tasks

### 1. OAuth 2.1 Authorization Framework

- [x] **Task 1.1**: Define OAuth 2.1 authentication interfaces in `pkg/server/auth.go`
- [x] **Task 1.2**: Implement OAuth 2.1 token validation and verification
- [x] **Task 1.3**: Create middleware for OAuth token validation
- [x] **Task 1.4**: Add scope-based permission system for tool access
- [x] **Task 1.5**: Update PocketBase integration to support OAuth 2.1
- [ ] **Task 1.6**: Create documentation for authorization setup and configuration
- [ ] **Task 1.7**: Implement tests for OAuth authentication flows

### 2. Streamable HTTP Transport

- [ ] **Task 2.1**: Define interface for streamable HTTP transport in `pkg/server/transport.go`
- [ ] **Task 2.2**: Implement streamable response writer
- [ ] **Task 2.3**: Update existing SSE implementation to use new transport
- [ ] **Task 2.4**: Add support for streaming binary data (for audio support)
- [ ] **Task 2.5**: Create adapters for different streaming protocols
- [ ] **Task 2.6**: Update client notification system to use new transport
- [ ] **Task 2.7**: Implement tests for streamable transport
- [ ] **Task 2.8**: Update documentation with new transport details

### 3. JSON-RPC Batching

- [ ] **Task 3.1**: Add batch request handler in `pkg/server/jsonrpc.go`
- [ ] **Task 3.2**: Implement concurrent execution of batched requests
- [ ] **Task 3.3**: Add result collation for batch responses
- [ ] **Task 3.4**: Create error handling for partial batch failures
- [ ] **Task 3.5**: Add batch size limits and validation
- [ ] **Task 3.6**: Update HTTP handlers to support batch endpoints
- [ ] **Task 3.7**: Implement tests for batch processing
- [ ] **Task 3.8**: Update documentation with batching examples

### 4. Tool Annotations

- [ ] **Task 4.1**: Extend tool definition schema in `pkg/tools` to include annotations
- [ ] **Task 4.2**: Add read-only/destructive operation flags
- [ ] **Task 4.3**: Implement permission level requirements based on annotations
- [ ] **Task 4.4**: Update tool registration to include annotation metadata
- [ ] **Task 4.5**: Add validation for tool annotations
- [ ] **Task 4.6**: Update integration examples to demonstrate annotations
- [ ] **Task 4.7**: Create tests for tool annotation handling
- [ ] **Task 4.8**: Update documentation with annotation guidelines

### 5. Other Schema Updates

- [ ] **Task 5.1**: Add `message` field to `ProgressNotification` in notification system
- [ ] **Task 5.2**: Implement audio data support in content types
- [ ] **Task 5.3**: Add `completions` capability flag for argument autocompletion
- [ ] **Task 5.4**: Update schema validation to match latest MCP specification
- [ ] **Task 5.5**: Update all example tools to use the new schema features
- [ ] **Task 5.6**: Create tests for new schema elements
- [ ] **Task 5.7**: Update documentation with new schema examples

## Migration Strategy

To ensure a smooth transition to the new MCP specification:

1. **Backward Compatibility**: Maintain support for the previous specification (2024-11-05) during transition
2. **Phased Implementation**: Implement changes in the following order:
   - Tool Annotations (lowest impact)
   - Schema Updates
   - JSON-RPC Batching
   - Streamable HTTP Transport
   - OAuth 2.1 Authorization Framework (highest impact)
3. **Version Flagging**: Add version headers to allow clients to request specific protocol versions
4. **Documentation Updates**: Keep documentation in sync with implementation progress

## Testing Approach

For each implemented task:

1. Write unit tests before implementation (TDD approach)
2. Create integration tests that verify compatibility with the specification
3. Implement example clients that exercise the new functionality
4. Verify backward compatibility with existing clients

## Timeline

The estimated completion timeline for aligning with MCP 2025-03-26:

- Phase 1 (Tool Annotations & Schema Updates): 2 weeks
- Phase 2 (JSON-RPC Batching): 1 week
- Phase 3 (Streamable HTTP Transport): 2 weeks
- Phase 4 (OAuth 2.1 Framework): 3 weeks

Total estimated time: 8 weeks

## Resources

- [MCP 2025-03-26 Specification](https://modelcontextprotocol.io/specification/2025-03-26/)
- [MCP Changelog](https://modelcontextprotocol.io/specification/2025-03-26/changelog)
- [OAuth 2.1 Specification](https://oauth.net/2.1/) 