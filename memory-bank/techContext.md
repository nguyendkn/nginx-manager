# Technical Context: Nginx Manager

## Technology Stack

### Backend Technologies

**Language**: Go 1.24.4
- High performance compiled language
- Excellent concurrency support
- Strong standard library
- Cross-platform compatibility
- Memory safety and garbage collection

**Web Framework**: Gin v1.10.1
- High-performance HTTP web framework
- Middleware support for cross-cutting concerns
- JSON binding and validation
- Minimal overhead and fast routing
- Excellent documentation and community support

**Logging**: Uber Zap v1.27.0
- High-performance structured logging
- Zero-allocation logging in production
- Configurable output formats (JSON/Console)
- Level-based logging with runtime configuration
- Context-aware logging with fields

### Frontend Technologies

**Framework**: React 19.1.0
- Latest React with concurrent features
- Server-side rendering support
- Hot module replacement for development
- Modern hooks and component patterns
- Excellent TypeScript integration

**Routing**: React Router v7.5.3
- File-based routing system
- Server-side rendering out of the box
- Type-safe routing with TypeScript
- Data loading and mutations
- Nested routing and layouts

**Language**: TypeScript 5.8.3
- Type safety for JavaScript
- Enhanced IDE support and intellisense
- Compile-time error checking
- Modern ECMAScript features
- Excellent React integration

**Build Tool**: Vite 6.3.3
- Lightning-fast development server
- Hot module replacement (HMR)
- Optimized production builds
- Plugin ecosystem for extensibility
- Native ES modules support

**Styling**: TailwindCSS 4.1.4
- Utility-first CSS framework
- Responsive design utilities
- Modern color palette and spacing
- Component-friendly approach
- Optimized bundle size

**UI Components**: Radix UI + shadcn/ui
- Comprehensive component library with:
- Accessible, unstyled components
- Full keyboard navigation support
- Focus management and ARIA attributes
- Customizable with TailwindCSS
- Production-ready patterns

### Dependencies Analysis

#### Backend Dependencies

```go
github.com/gin-gonic/gin v1.10.1           // Web framework
go.uber.org/zap v1.27.0                    // Structured logging
go.uber.org/multierr v1.10.0               // Error aggregation
github.com/go-playground/validator/v10     // Input validation
```

**Performance Dependencies**:
```go
github.com/bytedance/sonic v1.11.6         // High-performance JSON
github.com/cloudwego/base64x v0.1.4        // Optimized base64
github.com/klauspost/cpuid/v2 v2.2.7       // CPU feature detection
```

**Utility Dependencies**:
```go
github.com/gabriel-vasile/mimetype v1.4.3  // MIME type detection
github.com/pelletier/go-toml/v2 v2.2.2     // TOML configuration
```

#### Frontend Dependencies

**Core React Dependencies**:
- `react@19.1.0` - React framework
- `react-dom@19.1.0` - React DOM bindings
- `react-router@7.5.3` - Routing and navigation
- `@react-router/node@7.5.3` - Server-side rendering
- `@react-router/serve@7.5.3` - Production server

**UI and Styling**:
- `tailwindcss@4.1.4` - Utility-first CSS framework
- `@radix-ui/*` - Accessible UI component primitives
- `lucide-react@0.516.0` - Icon library
- `next-themes@0.4.6` - Theme management

**Forms and Validation**:
- `react-hook-form@7.58.1` - Form state management
- `@hookform/resolvers@5.1.1` - Form validation resolvers
- `zod@3.25.67` - Schema validation library

**Development Dependencies**:
- `typescript@5.8.3` - TypeScript compiler
- `vite@6.3.3` - Build tool and dev server
- `@tailwindcss/vite@4.1.4` - TailwindCSS Vite integration

## Development Environment

### Go Version Requirements

**Minimum Version**: Go 1.24.4
- Latest language features and performance improvements
- Enhanced error handling and generics support
- Improved build performance and module management
- Security updates and bug fixes

### Build Configuration

**Module System**: Go modules for dependency management
- Semantic versioning for dependencies
- Reproducible builds with go.sum
- Vendor directory support for offline builds
- Private module support for internal packages

### Development Tools

**Recommended IDE Setup**:
- VS Code with Go extension
- GoLand for advanced debugging
- Vim/Neovim with go.nvim plugin

**Essential Tools**:
- `go fmt`: Code formatting
- `go vet`: Static analysis
- `golangci-lint`: Comprehensive linting
- `go test`: Testing framework
- `go mod`: Dependency management

## Configuration Architecture

### Environment-Based Configuration

**Design Philosophy**:
- Environment variables override defaults
- Type-safe configuration structs
- Validation at startup
- Clear separation of concerns

**Configuration Categories**:

1. **Server Configuration**:
   ```go
   Port: "8080"              // Server port
   Host: "0.0.0.0"          // Bind address
   ```

2. **Application Settings**:
   ```go
   AppName: "nginx-manager"  // Service name
   AppVersion: "1.0.0"      // Version identifier
   AppEnvironment: "dev"    // Environment mode
   ```

3. **Framework Configuration**:
   ```go
   GinMode: "debug"         // Gin framework mode
   ```

4. **CORS Settings**:
   ```go
   CORSAllowedOrigins: ["*"]
   CORSAllowedMethods: ["GET", "POST", "PUT", "DELETE"]
   CORSAllowedHeaders: ["*"]
   ```

5. **Logging Configuration**:
   ```go
   LogLevel: "info"         // Minimum log level
   LogEncoding: "console"   // Output format
   ```

### Configuration Validation

**Startup Validation**:
- Required fields verification
- Format validation (ports, URLs)
- Dependency checks
- Environment-specific validations

## Logging Infrastructure

### Zap Logger Configuration

**Production Settings**:
```go
LogLevel: "info"
LogEncoding: "json"      // Structured JSON for parsing
```

**Development Settings**:
```go
LogLevel: "debug"
LogEncoding: "console"   // Human-readable output
```

### Logging Features

**Request Tracking**:
- Unique request IDs for tracing
- Request/response logging
- Error context preservation
- Performance metrics

**Structured Fields**:
- Client IP tracking
- User agent logging
- Response time measurement
- Error categorization

## Performance Considerations

### Framework Performance

**Gin Framework Benefits**:
- 40x faster than Martini
- Zero allocation router
- Efficient middleware pipeline
- Memory pooling for requests

**Zap Logging Performance**:
- Zero allocation in production
- 4-10x faster than standard library
- Structured output without reflection
- Configurable sampling rates

### Memory Management

**Go Runtime Optimization**:
- Garbage collector tuning for low latency
- Memory pooling for frequent allocations
- Efficient string handling
- Minimal heap allocations

### Concurrency Model

**Goroutine-Based Concurrency**:
- One goroutine per HTTP request
- Channel-based communication
- Context-aware cancellation
- Resource pooling for shared resources

## Security Considerations

### Input Validation

**Gin Validator Integration**:
- Automatic JSON binding validation
- Custom validation rules
- Error message standardization
- Type safety enforcement

### CORS Configuration

**Cross-Origin Protection**:
- Configurable allowed origins
- Method-specific permissions
- Header control and validation
- Preflight request handling

### Security Headers

**Planned Security Enhancements**:
- Rate limiting middleware
- Authentication middleware
- Authorization controls
- Request size limits

## Deployment Architecture

### Build Process

**Single Binary Deployment**:
- Statically linked Go binary
- Minimal runtime dependencies
- Cross-platform compilation
- Container-friendly design

### Container Support

**Backend Docker Configuration**:
- Multi-stage builds for minimal image size
- Non-root user execution
- Health check endpoints
- Graceful shutdown handling

**Frontend Docker Configuration** (`webui/Dockerfile`):
- Multi-stage Node.js builds
- Development and production dependency separation
- Optimized build process with caching
- Production-ready container with minimal footprint

### Environment Support

**Multi-Environment Deployment**:
- Development: Debug mode with console logging
- Staging: Production-like with enhanced logging
- Production: Optimized performance and JSON logging

## Integration Patterns

### External System Integration

**File System Integration**:
- Direct file operations for Nginx configs
- Atomic file operations for safety
- Backup and rollback mechanisms
- Permission and ownership management

**Process Control Integration**:
- Nginx service control via system calls
- Configuration validation through nginx -t
- Signal handling for graceful reloads
- Process monitoring and health checks

### API Design Principles

**RESTful Design**:
- Resource-based URLs
- HTTP method semantics
- Status code standards
- Content negotiation support

**JSON API Standards**:
- Consistent response formats
- Error response standardization
- Pagination support for collections
- Versioning strategy for compatibility

## Future Technical Considerations

### Planned Technology Additions

**Database Integration**:
- PostgreSQL or SQLite for persistence
- Migration management
- Connection pooling
- Query optimization

**Monitoring and Observability**:
- Prometheus metrics integration
- OpenTelemetry tracing
- Health check standardization
- Performance monitoring

**Authentication and Authorization**:
- JWT token management
- Role-based access control
- API key authentication
- OAuth2 integration support