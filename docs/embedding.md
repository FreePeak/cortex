# Embedding Cortex in External Servers

This document outlines how to embed Cortex functionality into an existing server application like PocketBase.

## Overview

Cortex is designed to be either used as a standalone server or embedded within another application. When embedded, Cortex can provide MCP (Model Context Protocol) capabilities to your existing server without requiring a separate process.

## Integration Methods

There are several ways to integrate Cortex with your existing application:

1. **Middleware Integration** - For web frameworks that support middleware, Cortex can be added as middleware to handle MCP requests.
2. **Route Handler Integration** - For frameworks with custom routing, Cortex handlers can be registered with your existing router.
3. **Plugin System Integration** - For applications with plugin capabilities, Cortex can be packaged as a plugin.

## Implementation Tasks

To support embedding Cortex into external servers, we need to complete the following tasks:

1. Create an `Embeddable` interface that defines the integration points
2. Implement HTTP handler adapter for easy integration with HTTP servers
3. Create middleware adapters for common Go web frameworks
4. Develop a dedicated PocketBase plugin implementation
5. Add examples of embedding Cortex in different server types
6. Create documentation with step-by-step integration guides
7. Implement testing utilities for embedded integrations

## PocketBase Integration

[PocketBase](https://github.com/pocketbase/pocketbase) is a good example of an application where Cortex can be embedded. PocketBase provides an API and database server with authentication. By embedding Cortex, PocketBase applications can gain MCP capabilities without running a separate server.

To integrate Cortex with PocketBase, we'll implement a plugin that:

1. Registers as a PocketBase plugin
2. Exposes Cortex tools and resources through PocketBase's HTTP router
3. Shares the application context with Cortex
4. Provides authentication between PocketBase and Cortex

## Next Steps

The following sections outline the specific tasks to implement embedding support in Cortex, starting with a minimal viable integration approach and building toward more sophisticated integration options. 