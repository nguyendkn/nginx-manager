import { api } from './client';

// Types for Proxy Host Management
export interface ProxyHost {
  id: number;
  domain_names: string[];
  forward_scheme: 'http' | 'https';
  forward_host: string;
  forward_port: number;
  access_list_id?: number;
  certificate_id?: number;
  ssl_forced: boolean;
  caching_enabled: boolean;
  block_exploits: boolean;
  allow_websocket_upgrade: boolean;
  http2_support: boolean;
  hsts_enabled: boolean;
  hsts_subdomains: boolean;
  advanced_config: string;
  enabled: boolean;
  locations?: Record<string, any>;
  meta?: Record<string, any>;
  created_at: string;
  updated_at: string;

  // Computed fields
  primary_domain: string;
  target_url: string;
  ssl_enabled: boolean;
  has_access_list: boolean;

  // Related entities
  certificate?: Certificate;
  access_list?: AccessList;
}

export interface ProxyHostDetail extends ProxyHost {
  nginx_config?: string;
  config_valid: boolean;
}

export interface Certificate {
  id: number;
  name: string;
  provider: string;
  domain_names: string[];
  expires_on: string;
  created_at: string;
  updated_at: string;
}

export interface AccessList {
  id: number;
  name: string;
  items: AccessListItem[];
  created_at: string;
  updated_at: string;
}

export interface AccessListItem {
  username?: string;
  password?: string;
  address?: string;
  directive: 'allow' | 'deny';
}

export interface CreateProxyHostRequest {
  domain_names: string[];
  forward_scheme: 'http' | 'https';
  forward_host: string;
  forward_port: number;
  access_list_id?: number;
  certificate_id?: number;
  ssl_forced?: boolean;
  caching_enabled?: boolean;
  block_exploits?: boolean;
  allow_websocket_upgrade?: boolean;
  http2_support?: boolean;
  hsts_enabled?: boolean;
  hsts_subdomains?: boolean;
  advanced_config?: string;
  enabled?: boolean;
  locations?: Record<string, any>;
  meta?: Record<string, any>;
}

export interface UpdateProxyHostRequest extends CreateProxyHostRequest {}

export interface ProxyHostListParams {
  page?: number;
  limit?: number;
  search?: string;
  enabled?: boolean;
}

export interface ProxyHostListResponse {
  data: ProxyHost[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
}

export interface BulkToggleRequest {
  ids: number[];
  enabled: boolean;
}

export interface BulkToggleResponse {
  updated: number;
  enabled: boolean;
}

// Proxy Host API Service
export const proxyHostsApi = {
  // List proxy hosts with pagination and filtering
  list: async (params: ProxyHostListParams = {}): Promise<ProxyHostListResponse> => {
    const searchParams = new URLSearchParams();

    if (params.page) searchParams.append('page', params.page.toString());
    if (params.limit) searchParams.append('limit', params.limit.toString());
    if (params.search) searchParams.append('search', params.search);
    if (params.enabled !== undefined) searchParams.append('enabled', params.enabled.toString());

    const response = await api.get<ProxyHostListResponse>(`/api/v1/proxy-hosts?${searchParams.toString()}`);
    return response.data.data as ProxyHostListResponse;
  },

  // Get a single proxy host by ID
  get: async (id: number): Promise<ProxyHostDetail> => {
    const response = await api.get<ProxyHostDetail>(`/api/v1/proxy-hosts/${id}`);
    return response.data.data as ProxyHostDetail;
  },

  // Create a new proxy host
  create: async (data: CreateProxyHostRequest): Promise<ProxyHost> => {
    const response = await api.post<ProxyHost>('/api/v1/proxy-hosts', data);
    return response.data.data as ProxyHost;
  },

  // Update an existing proxy host
  update: async (id: number, data: UpdateProxyHostRequest): Promise<ProxyHost> => {
    const response = await api.put<ProxyHost>(`/api/v1/proxy-hosts/${id}`, data);
    return response.data.data as ProxyHost;
  },

  // Delete a proxy host
  delete: async (id: number): Promise<{ id: number }> => {
    const response = await api.delete<{ id: number }>(`/api/v1/proxy-hosts/${id}`);
    return response.data.data as { id: number };
  },

  // Toggle enabled status of a proxy host
  toggle: async (id: number): Promise<{ id: number; enabled: boolean }> => {
    const response = await api.post<{ id: number; enabled: boolean }>(`/api/v1/proxy-hosts/${id}/toggle`);
    return response.data.data as { id: number; enabled: boolean };
  },

  // Bulk toggle multiple proxy hosts
  bulkToggle: async (data: BulkToggleRequest): Promise<BulkToggleResponse> => {
    const response = await api.post<BulkToggleResponse>('/api/v1/proxy-hosts/bulk-toggle', data);
    return response.data.data as BulkToggleResponse;
  },

  // Validate domain names (client-side validation)
  validateDomainNames: (domains: string[]): string[] => {
    const errors: string[] = [];

    if (domains.length === 0) {
      errors.push('At least one domain name is required');
      return errors;
    }

    domains.forEach((domain, index) => {
      const trimmed = domain.trim();
      if (!trimmed) {
        errors.push(`Domain ${index + 1} cannot be empty`);
        return;
      }

      // Basic domain validation
      if (trimmed.length > 253) {
        errors.push(`Domain ${index + 1} is too long (max 253 characters)`);
      }

      if (trimmed.includes(' ')) {
        errors.push(`Domain ${index + 1} cannot contain spaces`);
      }

      // Basic domain format check
      const domainRegex = /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
      if (!domainRegex.test(trimmed) && !trimmed.includes('*')) { // Allow wildcards
        errors.push(`Domain ${index + 1} has invalid format`);
      }
    });

    return errors;
  },

  // Validate forward configuration
  validateForwardConfig: (host: string, port: number, scheme: string): string[] => {
    const errors: string[] = [];

    if (!host.trim()) {
      errors.push('Forward host is required');
    }

    if (port < 1 || port > 65535) {
      errors.push('Forward port must be between 1 and 65535');
    }

    if (!['http', 'https'].includes(scheme)) {
      errors.push('Forward scheme must be http or https');
    }

    return errors;
  },

  // Generate nginx configuration preview (simplified)
  generateConfigPreview: (proxyHost: CreateProxyHostRequest | UpdateProxyHostRequest): string => {
    const serverNames = proxyHost.domain_names.join(' ');
    const upstream = `${proxyHost.forward_scheme}://${proxyHost.forward_host}:${proxyHost.forward_port}`;

    let config = `# Nginx configuration for ${proxyHost.domain_names[0]}\n`;
    config += `server {\n`;
    config += `    listen 80;\n`;

    if (proxyHost.ssl_forced || proxyHost.certificate_id) {
      config += `    listen 443 ssl${proxyHost.http2_support ? ' http2' : ''};\n`;
    }

    config += `    server_name ${serverNames};\n\n`;

    if (proxyHost.certificate_id) {
      config += `    # SSL Certificate\n`;
      config += `    ssl_certificate /path/to/certificate.crt;\n`;
      config += `    ssl_certificate_key /path/to/certificate.key;\n\n`;
    }

    if (proxyHost.ssl_forced && proxyHost.certificate_id) {
      config += `    # Force HTTPS\n`;
      config += `    if ($scheme != "https") {\n`;
      config += `        return 301 https://$host$request_uri;\n`;
      config += `    }\n\n`;
    }

    if (proxyHost.hsts_enabled) {
      config += `    # HSTS\n`;
      config += `    add_header Strict-Transport-Security "max-age=31536000${proxyHost.hsts_subdomains ? '; includeSubDomains' : ''}" always;\n\n`;
    }

    if (proxyHost.block_exploits) {
      config += `    # Block common exploits\n`;
      config += `    location ~* \.(aspx|php|jsp|cgi)$ {\n`;
      config += `        return 410;\n`;
      config += `    }\n\n`;
    }

    config += `    location / {\n`;
    config += `        proxy_pass ${upstream};\n`;
    config += `        proxy_set_header Host $host;\n`;
    config += `        proxy_set_header X-Real-IP $remote_addr;\n`;
    config += `        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n`;
    config += `        proxy_set_header X-Forwarded-Proto $scheme;\n`;

    if (proxyHost.allow_websocket_upgrade) {
      config += `        \n`;
      config += `        # WebSocket support\n`;
      config += `        proxy_http_version 1.1;\n`;
      config += `        proxy_set_header Upgrade $http_upgrade;\n`;
      config += `        proxy_set_header Connection "upgrade";\n`;
    }

    if (proxyHost.caching_enabled) {
      config += `        \n`;
      config += `        # Caching\n`;
      config += `        proxy_cache_bypass $http_upgrade;\n`;
      config += `        proxy_cache nginx_cache;\n`;
      config += `        proxy_cache_valid 200 302 10m;\n`;
      config += `        proxy_cache_valid 404 1m;\n`;
    }

    config += `    }\n`;

    if (proxyHost.advanced_config) {
      config += `\n    # Advanced Configuration\n`;
      config += `    ${proxyHost.advanced_config.split('\n').join('\n    ')}\n`;
    }

    config += `}\n`;

    return config;
  }
};

export default proxyHostsApi;
