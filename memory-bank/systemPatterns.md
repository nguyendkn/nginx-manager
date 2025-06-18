# System Patterns: Nginx Manager

## Architecture Overview

### Full-Stack Architecture Structure

The project follows clean architecture principles with clear separation between frontend and backend:

```
nginx-manager/
├── cmd/           # Backend application entry points
├── internal/      # Private backend application code
├── pkg/           # Public reusable backend packages
├── configs/       # Backend configuration management
└── webui/         # Frontend React application
    ├── app/       # React application code
    ├── components/# Reusable UI components
    ├── routes/    # File-based routing
    └── types/     # TypeScript type definitions
```

### Component Boundaries

**Frontend Layer** (`webui/`):
- React components and pages
- TypeScript interfaces and types
- TailwindCSS styling and themes
- Client-side routing and navigation
- Form handling and validation

**Backend External Layer**: HTTP handlers, CLI commands, file system interactions
**Backend Application Layer**: Controllers, routers, middleware
**Backend Domain Layer**: Business logic, models, services
**Backend Infrastructure Layer**: Logging, configuration, utilities

## Core Design Patterns

### 1. Dependency Injection Pattern

**Implementation**: Constructor-based dependency injection
```go
// Controller receives dependencies through constructor
func NewHealthController(env *configs.Environment) *HealthController {
    return &HealthController{env: env}
}
```

**Benefits**:
- Testable components with mock dependencies
- Flexible configuration across environments
- Clear dependency relationships

### 2. Middleware Chain Pattern

**Current Implementation**:
```go
r.Use(logger.RequestIDMiddleware())
r.Use(logger.GinLogger())
r.Use(logger.ErrorLogger())
r.Use(logger.RecoveryLogger())
r.Use(middleware.CORSMiddleware(env))
```

**Pattern Benefits**:
- Cross-cutting concerns handled consistently
- Request processing pipeline is configurable
- Easy to add/remove functionality

### 3. Configuration Pattern

**Environment-based Configuration**:
- Single source of truth for all settings
- Environment variable override support
- Default values for development
- Type-safe configuration access

**Implementation**:
```go
type Environment struct {
    Port           string
    Host           string
    AppName        string
    // ... other fields
}
```

### 4. Response Standardization Pattern

**Consistent API Responses**:
```go
response.SuccessJSONWithLog(c, data, message)
```

**Benefits**:
- Uniform response format across all endpoints
- Integrated logging with responses
- Error handling consistency

## Logging Architecture

### Structured Logging Pattern

**Implementation**: Uber Zap with structured fields
```go
logger.Info("Health check requested",
    logger.String("client_ip", c.ClientIP()),
    logger.String("user_agent", c.Request.UserAgent()),
)
```

**Features**:
- Request ID tracking through entire request lifecycle
- Contextual logging with structured fields
- Performance-optimized logging
- Environment-specific log levels and formats

### Request Lifecycle Logging

1. **Request ID Generation**: Unique identifier for request tracing
2. **Request Logging**: Incoming request details
3. **Processing Logging**: Business logic execution details
4. **Response Logging**: Response status and data
5. **Error Logging**: Structured error information with context

## Error Handling Patterns

### Centralized Error Recovery

**Panic Recovery Middleware**:
- Catches panics and converts to structured errors
- Maintains service stability
- Logs error context for debugging

**Error Response Structure**:
- Consistent error format across all endpoints
- Request ID included for tracing
- Appropriate HTTP status codes

## Multi-Component Architecture

### Service Separation

1. **Frontend Application** (`webui/`):
   - React-based web interface
   - Server-side rendering with React Router
   - Modern UI components with TailwindCSS
   - TypeScript for type safety

2. **HTTP Server** (`cmd/server/main.go`):
   - REST API service
   - Health monitoring endpoints
   - Request/response handling
   - CORS configuration for frontend

3. **CLI Tool** (`cmd/cli/`):
   - Command-line operations
   - Direct system integration
   - Automation support

4. **Cronjob Service** (`cmd/cronjob/`):
   - Scheduled tasks
   - Maintenance operations
   - Background processing

### Frontend Architecture Patterns

**Component Structure** (`webui/app/`):
- `components/`: Reusable UI components
- `routes/`: File-based routing with React Router
- `hooks/`: Custom React hooks
- `contexts/`: React context providers
- `services/`: API service layer
- `lib/`: Utility functions and helpers

**Design Patterns**:
- Component composition with shadcn/ui
- Custom hooks for state management
- Context providers for global state
- Service layer for API communication
- Type-safe routing with TypeScript

### Backend Infrastructure

**Common Packages** (`pkg/`):
- Logger: Centralized logging functionality
- Response: Standardized API responses
- Settings: Configuration management
- Utils: Shared utilities

## Controller Patterns

### Environment-Aware Controllers

**Pattern**: Controllers receive environment configuration
```go
type HealthController struct {
    env *configs.Environment
}
```

**Benefits**:
- Environment-specific behavior
- Configuration-driven functionality
- Easy testing with different configurations

### Health Check Implementation

**Comprehensive Health Data**:
- System metrics (memory, goroutines, GC)
- Application information (version, uptime)
- Environment configuration
- CORS settings

**Performance Considerations**:
- Efficient metric collection
- Minimal performance impact
- Real-time system status

## Middleware Patterns

### Request Processing Pipeline

1. **Request ID**: Generates unique identifier for request tracing
2. **Gin Logger**: HTTP request/response logging
3. **Error Logger**: Captures and logs errors
4. **Recovery Logger**: Handles panics gracefully
5. **CORS**: Cross-origin resource sharing configuration

### Middleware Benefits

- **Observability**: Complete request lifecycle visibility
- **Reliability**: Error recovery and graceful handling
- **Security**: CORS protection and request validation
- **Performance**: Efficient request processing

## Future Pattern Considerations

### Planned Patterns for Nginx Management

**Repository Pattern**: For configuration data persistence
**Service Layer Pattern**: For business logic encapsulation
**Factory Pattern**: For Nginx configuration generation
**Observer Pattern**: For configuration change notifications
**Strategy Pattern**: For different Nginx management approaches

### Scalability Patterns

**Connection Pooling**: For database connections (when added)
**Caching Layer**: For frequently accessed configurations
**Rate Limiting**: For API protection
**Circuit Breaker**: For external service integration

## Testing Patterns

### Current Testing Approach

- Unit testing for individual components
- Integration testing for API endpoints
- Mock dependencies for isolated testing
- Environment-specific test configurations

### Testing Infrastructure

- Test environment configuration
- Mock implementations of external dependencies
- Automated test execution in CI/CD
- Performance testing for critical paths
