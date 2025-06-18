# API Endpoints Mapping: NPM → nginx-manager

## Tổng quan API Design

**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: JWT Bearer Token
**Response Format**: JSON với cấu trúc chuẩn
**HTTP Methods**: RESTful conventions

## Authentication & Users API

### Authentication Routes
```go
// NPM: /api/tokens (POST)
POST /api/v1/auth/login
Request: {
  "email": "string",
  "password": "string"
}
Response: {
  "token": "string",
  "user": UserObject,
  "expires": "2024-01-01T00:00:00Z"
}

// NPM: /api/tokens (DELETE)
POST /api/v1/auth/logout
Headers: Authorization: Bearer <token>

// Refresh token
POST /api/v1/auth/refresh
Headers: Authorization: Bearer <token>
Response: {
  "token": "string",
  "expires": "2024-01-01T00:00:00Z"
}
```

### User Management Routes
```go
// NPM: /api/users
GET /api/v1/users
Query: ?page=1&limit=20&search=email
Response: {
  "data": [UserObject],
  "total": 123,
  "page": 1,
  "limit": 20
}

POST /api/v1/users
Request: {
  "email": "string",
  "name": "string",
  "nickname": "string",
  "password": "string",
  "roles": ["admin", "user"]
}

GET /api/v1/users/{id}
PUT /api/v1/users/{id}
DELETE /api/v1/users/{id}

// Current user profile
GET /api/v1/users/me
PUT /api/v1/users/me
```

## Nginx Management APIs

### Proxy Hosts Management
```go
// NPM: /api/nginx/proxy-hosts
GET /api/v1/proxy-hosts
Query: ?page=1&limit=20&search=domain&enabled=true
Response: {
  "data": [ProxyHostObject],
  "total": 45,
  "page": 1,
  "limit": 20
}

POST /api/v1/proxy-hosts
Request: {
  "domain_names": ["example.com", "www.example.com"],
  "forward_scheme": "http",
  "forward_host": "192.168.1.100",
  "forward_port": 3000,
  "caching_enabled": false,
  "block_exploits": true,
  "allow_websocket_upgrade": true,
  "access_list_id": 0,
  "certificate_id": 0,
  "ssl_forced": false,
  "hsts_enabled": false,
  "hsts_subdomains": false,
  "http2_support": true,
  "advanced_config": "string",
  "locations": [LocationObject],
  "meta": {}
}

GET /api/v1/proxy-hosts/{id}
PUT /api/v1/proxy-hosts/{id}
DELETE /api/v1/proxy-hosts/{id}

// Enable/Disable proxy host
PUT /api/v1/proxy-hosts/{id}/status
Request: {
  "enabled": true
}

// Test proxy host configuration
POST /api/v1/proxy-hosts/{id}/test
Response: {
  "valid": true,
  "errors": []
}
```

### SSL Certificate Management
```go
// NPM: /api/nginx/certificates
GET /api/v1/certificates
Query: ?page=1&limit=20&provider=letsencrypt&expires_soon=true
Response: {
  "data": [CertificateObject],
  "total": 12,
  "page": 1,
  "limit": 20
}

POST /api/v1/certificates
Request: {
  "provider": "letsencrypt", // "letsencrypt" | "custom"
  "domain_names": ["example.com", "www.example.com"],
  "meta": {
    "letsencrypt_email": "admin@example.com",
    "dns_challenge": false
  }
}

GET /api/v1/certificates/{id}
PUT /api/v1/certificates/{id}
DELETE /api/v1/certificates/{id}

// Renew certificate
POST /api/v1/certificates/{id}/renew
Response: {
  "status": "success",
  "expires_on": "2025-01-01T00:00:00Z"
}

// Upload custom certificate
POST /api/v1/certificates/upload
Request: {
  "name": "My Custom Cert",
  "certificate": "-----BEGIN CERTIFICATE-----...",
  "certificate_key": "-----BEGIN PRIVATE KEY-----...",
  "intermediate_certificate": "-----BEGIN CERTIFICATE-----..."
}

// Test certificate
POST /api/v1/certificates/{id}/test
Response: {
  "valid": true,
  "expires_on": "2025-01-01T00:00:00Z",
  "issuer": "Let's Encrypt Authority X3",
  "errors": []
}
```

### Access Lists Management
```go
// NPM: /api/nginx/access-lists
GET /api/v1/access-lists
POST /api/v1/access-lists
Request: {
  "name": "Office Network",
  "satisfy": "any", // "any" | "all"
  "pass_auth": true,
  "clients": [
    {
      "address": "192.168.1.0/24",
      "directive": "allow"
    },
    {
      "address": "10.0.0.1",
      "directive": "deny"
    }
  ],
  "items": [
    {
      "username": "admin",
      "password": "password123"
    }
  ]
}

GET /api/v1/access-lists/{id}
PUT /api/v1/access-lists/{id}
DELETE /api/v1/access-lists/{id}

// Test access list
POST /api/v1/access-lists/{id}/test
Request: {
  "client_ip": "192.168.1.100",
  "username": "admin",
  "password": "password123"
}
Response: {
  "allowed": true,
  "matched_rule": "192.168.1.0/24"
}
```

### Stream Proxies Management
```go
// NPM: /api/nginx/streams
GET /api/v1/streams
POST /api/v1/streams
Request: {
  "incoming_port": 3306,
  "forwarding_host": "db.internal.com",
  "forwarding_port": 3306,
  "tcp_forwarding": true,
  "udp_forwarding": false,
  "ssl_enabled": false,
  "certificate_id": 0
}

GET /api/v1/streams/{id}
PUT /api/v1/streams/{id}
DELETE /api/v1/streams/{id}
```

### Redirection Hosts Management
```go
// NPM: /api/nginx/redirection-hosts
GET /api/v1/redirection-hosts
POST /api/v1/redirection-hosts
Request: {
  "domain_names": ["old.example.com"],
  "forward_scheme": "https",
  "forward_domain_name": "new.example.com",
  "preserve_path": true,
  "certificate_id": 0,
  "ssl_forced": true,
  "hsts_enabled": false,
  "hsts_subdomains": false,
  "http2_support": true,
  "block_exploits": true,
  "advanced_config": "string"
}

GET /api/v1/redirection-hosts/{id}
PUT /api/v1/redirection-hosts/{id}
DELETE /api/v1/redirection-hosts/{id}
```

### Dead Hosts (404 Pages)
```go
// NPM: /api/nginx/dead-hosts
GET /api/v1/dead-hosts
POST /api/v1/dead-hosts
Request: {
  "domain_names": ["dead.example.com"],
  "certificate_id": 0,
  "ssl_forced": false,
  "hsts_enabled": false,
  "hsts_subdomains": false,
  "http2_support": true,
  "block_exploits": true,
  "advanced_config": "string"
}

GET /api/v1/dead-hosts/{id}
PUT /api/v1/dead-hosts/{id}
DELETE /api/v1/dead-hosts/{id}
```

## System Management APIs

### Settings Management
```go
// NPM: /api/settings
GET /api/v1/settings
Response: {
  "data": [SettingObject]
}

PUT /api/v1/settings
Request: {
  "settings": [
    {
      "id": "default-site",
      "name": "Default Site",
      "description": "What to show when Nginx gets a request that doesn't match anything",
      "value": "congratulations",
      "meta": {}
    }
  ]
}

GET /api/v1/settings/{id}
PUT /api/v1/settings/{id}
```

### Audit Logs
```go
// NPM: /api/audit-log
GET /api/v1/audit-logs
Query: ?page=1&limit=50&action=created&object_type=proxy-host&user_id=1
Response: {
  "data": [AuditLogObject],
  "total": 234,
  "page": 1,
  "limit": 50
}

GET /api/v1/audit-logs/{id}
```

### System Health & Reports
```go
// NPM: /api/reports/hosts
GET /api/v1/reports/hosts
Response: {
  "proxy_hosts": 15,
  "redirection_hosts": 3,
  "dead_hosts": 1,
  "streams": 2
}

// NPM: /api/reports/certificates
GET /api/v1/reports/certificates
Response: {
  "total": 8,
  "valid": 7,
  "expiring_soon": 1,
  "expired": 0
}

// System health
GET /api/v1/health
Response: {
  "status": "healthy",
  "nginx": {
    "running": true,
    "config_valid": true
  },
  "database": {
    "connected": true
  },
  "certificates": {
    "expiring_soon": 1
  }
}
```

## Go Controller Structure

```go
// internal/controllers/
├── auth_controller.go          // Authentication & JWT
├── user_controller.go          // User management
├── proxy_host_controller.go    // Proxy hosts CRUD
├── certificate_controller.go   // SSL certificate management
├── access_list_controller.go   // Access control
├── stream_controller.go        // Stream proxies
├── redirection_controller.go   // Redirection hosts
├── dead_host_controller.go     // Dead hosts (404 pages)
├── setting_controller.go       // System settings
├── audit_log_controller.go     // Audit logs
├── report_controller.go        // Reports & analytics
└── health_controller.go        // System health (already exists)
```

## Request/Response DTOs

```go
// internal/dto/
├── auth_dto.go
├── user_dto.go
├── proxy_host_dto.go
├── certificate_dto.go
├── access_list_dto.go
├── stream_dto.go
├── redirection_dto.go
├── dead_host_dto.go
├── setting_dto.go
├── audit_log_dto.go
└── common_dto.go  // Pagination, error responses
```

## Middleware Requirements

```go
// internal/middleware/
├── auth.go          // JWT validation
├── rbac.go          // Role-based access control
├── rate_limit.go    // API rate limiting
├── validation.go    // Request validation
└── audit.go         // Audit logging middleware
```

## Error Response Standard

```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Message string            `json:"message"`
    Code    int              `json:"code"`
    Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse struct {
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total  int `json:"total"`
    Page   int `json:"page"`
    Limit  int `json:"limit"`
    Pages  int `json:"pages"`
}
```

---

**Status**: ✅ Complete
**Next**: Business Logic Analysis
