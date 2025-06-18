# Business Logic Analysis: NPM Core Features

## 1. Authentication & Authorization Logic

### JWT Authentication Flow
```go
// NPM Implementation (Node.js)
// backend/lib/express/jwt.js

// Go Implementation Target
type AuthService struct {
    jwtSecret     string
    tokenExpiry   time.Duration
    refreshExpiry time.Duration
    userService   *UserService
}

// Key Components:
// 1. JWT token generation with user claims
// 2. Token refresh mechanism
// 3. Password hashing with bcrypt
// 4. Role-based permissions checking
```

**Core Business Rules**:
- JWT tokens expire after configurable time (default: 1 day)
- Refresh tokens válid longer (default: 30 days)
- Passwords must be bcrypt hashed
- Default admin user created on first setup
- Email uniqueness constraint
- Role-based access control (admin, user)

### User Permission System
```go
// Roles and Permissions
type Role string
const (
    RoleAdmin Role = "admin"
    RoleUser  Role = "user"
)

// Permission matrix
var PermissionMatrix = map[Role][]string{
    RoleAdmin: {
        "users:create", "users:read", "users:update", "users:delete",
        "proxy-hosts:*", "certificates:*", "access-lists:*",
        "streams:*", "redirections:*", "dead-hosts:*",
        "settings:*", "audit-logs:read",
    },
    RoleUser: {
        "proxy-hosts:read", "proxy-hosts:create", "proxy-hosts:update",
        "certificates:read", "certificates:create",
        "users:read:own", "users:update:own",
    },
}
```

## 2. SSL Certificate Management Logic

### Let's Encrypt Integration
```go
// NPM: backend/lib/certbot.js analysis

type CertificateService struct {
    acmeClient    *acme.Client
    storageDir    string
    nginxService  *NginxService
    scheduler     *cron.Cron
}

// Core Business Logic:
// 1. ACME challenge handling (HTTP-01, DNS-01)
// 2. Certificate installation và validation
// 3. Auto-renewal scheduling (30 days before expiry)
// 4. Nginx configuration update after cert changes
// 5. Certificate backup và rollback
```

**Key Business Rules**:
- Certificates auto-renew 30 days before expiration
- HTTP-01 challenge requires domain pointing to server
- DNS-01 challenge requires DNS provider API
- Certificate files stored securely with proper permissions
- Nginx reloaded gracefully after certificate changes
- Failed renewals trigger alerts
- Multiple domains per certificate supported (SAN)

### Custom Certificate Upload
```go
// Certificate Validation Logic
func (cs *CertificateService) ValidateCustomCertificate(cert, key, intermediate string) error {
    // 1. Parse certificate and key
    // 2. Verify key matches certificate
    // 3. Check certificate chain validity
    // 4. Validate expiration date
    // 5. Extract domain names (CN, SAN)
    return nil
}
```

## 3. Nginx Configuration Generation

### Template System Analysis
```go
// NPM: Nginx config template generation
// Location: backend/internal/nginx.js

type NginxConfigGenerator struct {
    templateDir    string
    configDir      string
    backupDir      string
    testCommand    string // "nginx -t"
    reloadCommand  string // "nginx -s reload"
}

// Core Templates:
// 1. proxy_host.conf.template
// 2. redirection_host.conf.template
// 3. dead_host.conf.template
// 4. stream.conf.template
// 5. default.conf.template
```

**Template Variables**:
```nginx
# proxy_host.conf.template example
server {
    listen 80;
    {{#each domain_names}}
    server_name {{this}};
    {{/each}}

    {{#if certificate_id}}
    listen 443 ssl{{#if http2_support}} http2{{/if}};
    ssl_certificate {{certificate_path}};
    ssl_certificate_key {{certificate_key_path}};
    {{/if}}

    {{#if access_list_id}}
    include {{access_list_path}};
    {{/if}}

    location / {
        proxy_pass {{forward_scheme}}://{{forward_host}}:{{forward_port}};
        {{#if allow_websocket_upgrade}}
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        {{/if}}
        {{advanced_config}}
    }

    {{#each locations}}
    location {{path}} {
        {{config}}
    }
    {{/each}}
}
```

### Configuration Management Process
```go
// Safe Configuration Update Flow
func (ng *NginxConfigGenerator) UpdateConfiguration(host *ProxyHost) error {
    // 1. Generate new configuration
    newConfig := ng.generateConfig(host)

    // 2. Write to temporary file
    tempFile := ng.writeTemporaryConfig(newConfig)

    // 3. Test configuration validity
    if err := ng.testConfiguration(); err != nil {
        return err
    }

    // 4. Backup current configuration
    ng.backupCurrentConfig(host.ID)

    // 5. Move temp file to active location
    ng.activateConfiguration(tempFile, host.ID)

    // 6. Reload Nginx gracefully
    return ng.reloadNginx()
}
```

## 4. Proxy Host Management Logic

### Domain Name Validation
```go
type ProxyHostService struct {
    nginxGenerator *NginxConfigGenerator
    certificateService *CertificateService
    auditService   *AuditService
}

// Business Rules:
// 1. Domain names must be válid FQDN or wildcard
// 2. No duplicate domain names across proxy hosts
// 3. Automatic SSL certificate assignment
// 4. Custom locations supported với priority ordering
// 5. Access list integration
// 6. WebSocket upgrade support
// 7. HSTS configuration
// 8. HTTP/2 support
```

### Custom Locations Logic
```go
type Location struct {
    ID          uint   `json:"id"`
    ProxyHostID uint   `json:"proxy_host_id"`
    Path        string `json:"path" validate:"required"`
    ForwardScheme string `json:"forward_scheme"`
    ForwardHost   string `json:"forward_host"`
    ForwardPort   int    `json:"forward_port"`
    AdvancedConfig string `json:"advanced_config"`
    Priority       int    `json:"priority"` // Higher = processed first
}

// Location matching priority:
// 1. Exact match
// 2. Longest prefix match
// 3. Regular expression match
// 4. Default location (/)
```

## 5. Access Control Logic

### IP-based Access Control
```go
type AccessListService struct {
    ipRangeService *IPRangeService
    authService    *AuthenticationService
}

// Access Control Rules:
// 1. IP address/CIDR range checking
// 2. Username/password authentication
// 3. "Satisfy any" vs "Satisfy all" logic
// 4. Pass authentication option
// 5. Client directive (allow/deny)
```

### Access Control Implementation
```go
func (als *AccessListService) CheckAccess(clientIP, username, password string, accessList *AccessList) (bool, error) {
    ipAllowed := false
    authPassed := false

    // Check IP-based rules
    for _, client := range accessList.Clients {
        if als.ipRangeService.Contains(client.Address, clientIP) {
            if client.Directive == "allow" {
                ipAllowed = true
            } else {
                return false, nil // Explicit deny
            }
        }
    }

    // Check authentication if required
    if len(accessList.Items) > 0 {
        authPassed = als.authService.ValidateCredentials(username, password, accessList.Items)
    }

    // Apply satisfy logic
    switch accessList.Satisfy {
    case "any":
        return ipAllowed || authPassed, nil
    case "all":
        return ipAllowed && authPassed, nil
    default:
        return ipAllowed, nil
    }
}
```

## 6. Stream Proxy Logic

### TCP/UDP Proxying
```go
type StreamService struct {
    configGenerator *NginxStreamConfigGenerator
    portManager     *PortManager
}

// Business Rules:
// 1. Port conflict detection
// 2. TCP và UDP protocol support
// 3. SSL termination for streams
// 4. Load balancing support
// 5. Health checking for upstream servers
```

### Stream Configuration Template
```nginx
# stream.conf.template
upstream stream_{{id}} {
    server {{forwarding_host}}:{{forwarding_port}};
}

server {
    listen {{incoming_port}}{{#if udp_forwarding}} udp{{/if}};
    {{#if ssl_enabled}}
    ssl_certificate {{certificate_path}};
    ssl_certificate_key {{certificate_key_path}};
    {{/if}}
    proxy_pass stream_{{id}};
    proxy_timeout 1s;
    proxy_responses 1;
}
```

## 7. Audit Logging Logic

### Activity Tracking
```go
type AuditService struct {
    repository *AuditLogRepository
    userService *UserService
}

// Tracked Events:
// 1. All CRUD operations
// 2. Authentication events
// 3. Configuration changes
// 4. Certificate operations
// 5. System setting changes

type AuditLog struct {
    ID         uint      `json:"id"`
    Action     string    `json:"action"` // created, updated, deleted
    ObjectType string    `json:"object_type"` // proxy_host, certificate, etc.
    ObjectID   uint      `json:"object_id"`
    UserID     uint      `json:"user_id"`
    Meta       JSON      `json:"meta"` // Changed fields, old values
    CreatedAt  time.Time `json:"created_at"`
}
```

## 8. Settings Management Logic

### System Configuration
```go
type SettingService struct {
    repository *SettingRepository
    cache      *Cache
}

// Core Settings:
// 1. Default site behavior
// 2. Global SSL settings
// 3. Rate limiting configuration
// 4. Backup retention policy
// 5. Certificate renewal settings
// 6. Email notification settings

type Setting struct {
    ID          string      `json:"id" gorm:"primaryKey"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Value       interface{} `json:"value" gorm:"type:text"`
    Meta        JSON        `json:"meta"`
    UpdatedAt   time.Time   `json:"updated_at"`
}
```

## 9. Background Tasks & Scheduling

### Automated Processes
```go
type SchedulerService struct {
    cron           *cron.Cron
    certService    *CertificateService
    backupService  *BackupService
    cleanupService *CleanupService
}

// Scheduled Tasks:
// 1. Certificate renewal check (daily)
// 2. Configuration backup (weekly)
// 3. Log cleanup (daily)
// 4. Health status check (hourly)
// 5. IP range updates (daily)
// 6. Dead host monitoring (hourly)
```

## 10. Error Handling & Recovery

### Rollback Mechanisms
```go
type ConfigurationManager struct {
    backupService *BackupService
    nginx         *NginxService
}

func (cm *ConfigurationManager) SafeUpdate(updateFunc func() error) error {
    // 1. Create backup point
    backupID := cm.backupService.CreateBackup()

    // 2. Apply changes
    if err := updateFunc(); err != nil {
        // 3. Rollback on failure
        cm.backupService.RestoreBackup(backupID)
        cm.nginx.Reload()
        return err
    }

    // 4. Verify configuration
    if err := cm.nginx.TestConfiguration(); err != nil {
        cm.backupService.RestoreBackup(backupID)
        cm.nginx.Reload()
        return err
    }

    // 5. Reload if successful
    return cm.nginx.Reload()
}
```

---

## Implementation Priority

### Phase 2A (Week 3): Core Services
1. **AuthService** - JWT + user management
2. **NginxConfigGenerator** - Template system
3. **CertificateService** - Let's Encrypt integration

### Phase 2B (Week 4): Feature Services
1. **ProxyHostService** - Main proxy management
2. **AccessListService** - Access control
3. **AuditService** - Activity logging

### Key Integration Points
- All config changes go through SafeUpdate pattern
- All user actions logged via AuditService
- Certificate operations integrated with Nginx reload
- Background tasks managed by SchedulerService

**Status**: ✅ Complete
**Next**: Begin Phase 2 Implementation
