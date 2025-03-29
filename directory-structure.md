# Directory Structure

The golang-mcp-server-sdk is organized following clean architecture principles and GoLang best practices.

## Directory Overview

```
golang-mcp-server-sdk/
├── bin/                    # Compiled binaries
├── cmd/                    # Example MCP server applications
│   ├── echo-sse-server/    # SSE-based example server
│   ├── echo-stdio-server/  # StdIO-based example server
│   └── multi-protocol-server/  # Server that supports multiple transport methods
├── examples/               # Example code snippets and use cases
├── internal/               # Private implementation details (not exposed to users)
│   ├── builder/            # Internal builder implementation
│   ├── domain/             # Core domain models and interfaces
│   ├── infrastructure/     # Implementation of core interfaces
│   ├── interfaces/         # Transport adapters (stdio, rest)
│   └── usecases/           # Business logic and use cases
└── pkg/                    # Public API (exposed to users)
    ├── builder/            # Public builder pattern for server construction
    ├── server/             # Public server implementation
    ├── tools/              # Utilities for creating MCP tools
    └── types/              # Shared types and interfaces
```

## Public API (pkg/)

The `pkg/` directory contains all publicly exposed APIs that users of the SDK should interact with:

- **pkg/builder/**: Builder pattern implementation for creating MCP servers.
- **pkg/server/**: Core server implementation with support for different transports.
- **pkg/tools/**: Helper functions for creating and configuring MCP tools.
- **pkg/types/**: Core data structures and interfaces shared across the SDK.

## Private Implementation (internal/)

The `internal/` directory contains implementation details that are not part of the public API:

- **internal/domain/**: Core domain models, interfaces, and business logic.
- **internal/infrastructure/**: Implementation of domain interfaces.
- **internal/interfaces/**: Transport adapters for different protocols.
- **internal/usecases/**: Application business rules and use cases.

## Examples and Applications

- **cmd/**: Complete server applications that showcase different use cases.
- **examples/**: Code snippets that demonstrate specific features of the SDK.

## Usage Guidelines

1. **For library consumers**:
   - Import only from the `pkg/` directory.
   - Use the builder pattern from `pkg/builder/` to create servers.
   - Use types from `pkg/types/` for interface parameters.
   - Use helpers from `pkg/tools/` to create tools with parameters.

2. **For library developers**:
   - Maintain clean separation between internal and public APIs.
   - Ensure backwards compatibility for the public API.
   - Add adapters between internal and public types to allow for future changes.

## Architectural Principles

The SDK follows clean architecture principles:

1. **Dependency Rule**: Inner layers don't depend on outer layers.
   - Domain doesn't depend on infrastructure.
   - Use cases depend only on domain.

2. **Adapter Pattern**: Adapters connect external interfaces to internal logic.
   - Transport adapters (stdio, rest) abstract communication protocols.
   - Repository adapters abstract data storage.

3. **Dependency Injection**: Dependencies are provided from outside.
   - Services are constructed with their dependencies.
   - This makes testing easier and components more reusable.

4. **Interface Segregation**: Small, focused interfaces.
   - Repositories have specific purposes.
   - Handlers deal with specific types of messages.

These principles make the SDK maintainable, testable, and adaptable to changes in requirements or external systems. 