import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Button } from '~/components/ui/button'
import { Badge } from '~/components/ui/badge'
import { Input } from '~/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '~/components/ui/tabs'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Search, Copy, Plus, Globe, Shield, Zap, Database, Settings } from 'lucide-react'

interface Snippet {
  id: string
  name: string
  description: string
  category: 'proxy' | 'ssl' | 'cache' | 'security' | 'load_balance' | 'general'
  content: string
  variables?: string[]
}

const NGINX_SNIPPETS: Snippet[] = [
  {
    id: 'basic-proxy',
    name: 'Basic Proxy',
    category: 'proxy',
    description: 'Simple reverse proxy configuration',
    content: `location / {
    proxy_pass http://{{backend_url}};
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}`,
    variables: ['backend_url']
  },
  {
    id: 'ssl-config',
    name: 'SSL Configuration',
    category: 'ssl',
    description: 'Modern SSL/TLS configuration with security headers',
    content: `listen 443 ssl http2;
ssl_certificate {{cert_path}};
ssl_certificate_key {{key_path}};

# SSL Settings
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;

# Security Headers
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Frame-Options DENY always;
add_header X-Content-Type-Options nosniff always;`,
    variables: ['cert_path', 'key_path']
  },
  {
    id: 'load-balancer',
    name: 'Load Balancer',
    category: 'load_balance',
    description: 'Upstream load balancing with health checks',
    content: `upstream {{upstream_name}} {
    least_conn;
    server {{server1}} weight=3 max_fails=2 fail_timeout=30s;
    server {{server2}} weight=3 max_fails=2 fail_timeout=30s;
    server {{server3}} backup;
}

location / {
    proxy_pass http://{{upstream_name}};
    proxy_next_upstream error timeout http_500 http_502 http_503;
    proxy_connect_timeout 5s;
    proxy_send_timeout 10s;
    proxy_read_timeout 10s;
}`,
    variables: ['upstream_name', 'server1', 'server2', 'server3']
  },
  {
    id: 'cache-config',
    name: 'Proxy Cache',
    category: 'cache',
    description: 'Proxy caching configuration for better performance',
    content: `proxy_cache_path /var/cache/nginx levels=1:2 keys_zone={{cache_name}}:10m max_size=10g
                 inactive=60m use_temp_path=off;

location / {
    proxy_cache {{cache_name}};
    proxy_cache_valid 200 302 10m;
    proxy_cache_valid 404 1m;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
    proxy_cache_lock on;

    add_header X-Cache-Status $upstream_cache_status;

    proxy_pass http://{{backend}};
}`,
    variables: ['cache_name', 'backend']
  },
  {
    id: 'rate-limit',
    name: 'Rate Limiting',
    category: 'security',
    description: 'Rate limiting to prevent abuse',
    content: `limit_req_zone $binary_remote_addr zone={{zone_name}}:10m rate={{rate}};

location / {
    limit_req zone={{zone_name}} burst={{burst}} nodelay;
    limit_req_status 429;

    # Optional: Custom error page for rate limiting
    error_page 429 /rate_limit.html;

    proxy_pass http://{{backend}};
}`,
    variables: ['zone_name', 'rate', 'burst', 'backend']
  },
  {
    id: 'gzip-compression',
    name: 'Gzip Compression',
    category: 'general',
    description: 'Gzip compression for better performance',
    content: `gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_comp_level 6;
gzip_types
    text/plain
    text/css
    text/xml
    text/javascript
    application/json
    application/javascript
    application/xml+rss
    application/atom+xml
    image/svg+xml;`,
    variables: []
  },
  {
    id: 'websocket-proxy',
    name: 'WebSocket Proxy',
    category: 'proxy',
    description: 'WebSocket proxying configuration',
    content: `location /ws {
    proxy_pass http://{{websocket_backend}};
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket specific timeouts
    proxy_read_timeout 86400;
    proxy_send_timeout 86400;
}`,
    variables: ['websocket_backend']
  }
]

const CATEGORY_ICONS = {
  proxy: Globe,
  ssl: Shield,
  cache: Zap,
  security: Shield,
  load_balance: Database,
  general: Settings
}

const CATEGORY_COLORS = {
  proxy: 'bg-blue-100 text-blue-800',
  ssl: 'bg-green-100 text-green-800',
  cache: 'bg-yellow-100 text-yellow-800',
  security: 'bg-red-100 text-red-800',
  load_balance: 'bg-purple-100 text-purple-800',
  general: 'bg-gray-100 text-gray-800'
}

interface ConfigSnippetsProps {
  onInsert: (content: string) => void
}

export default function ConfigSnippets({ onInsert }: ConfigSnippetsProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('all')

  // Filter snippets based on search and category
  const filteredSnippets = NGINX_SNIPPETS.filter(snippet => {
    const matchesSearch = snippet.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         snippet.description.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesCategory = selectedCategory === 'all' || snippet.category === selectedCategory
    return matchesSearch && matchesCategory
  })

  const categories = Array.from(new Set(NGINX_SNIPPETS.map(s => s.category)))

  const handleInsertSnippet = (snippet: Snippet) => {
    let content = snippet.content

    // Replace variables with placeholders
    if (snippet.variables && snippet.variables.length > 0) {
      snippet.variables.forEach((variable, index) => {
        const placeholder = `${variable.replace(/_/g, ' ')}`
        content = content.replace(new RegExp(`{{${variable}}}`, 'g'), placeholder)
      })
    }

    onInsert(content)
  }

  const copyToClipboard = async (content: string) => {
    try {
      await navigator.clipboard.writeText(content)
    } catch (err) {
      console.error('Failed to copy to clipboard:', err)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Settings className="h-4 w-4" />
          Configuration Snippets
        </CardTitle>
        <CardDescription>
          Quick access to common nginx configuration patterns
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Search and Filter */}
          <div className="space-y-2">
            <div className="relative">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search snippets..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-8"
              />
            </div>

            {/* Category Tabs */}
            <Tabs value={selectedCategory} onValueChange={setSelectedCategory}>
              <TabsList className="w-full h-auto p-1 grid grid-cols-3 gap-1">
                <TabsTrigger value="all" className="text-xs">All</TabsTrigger>
                <TabsTrigger value="proxy" className="text-xs">Proxy</TabsTrigger>
                <TabsTrigger value="ssl" className="text-xs">SSL</TabsTrigger>
                <TabsTrigger value="cache" className="text-xs">Cache</TabsTrigger>
                <TabsTrigger value="security" className="text-xs">Security</TabsTrigger>
                <TabsTrigger value="load_balance" className="text-xs">Load Balance</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          {/* Snippets List */}
          <ScrollArea className="h-96">
            <div className="space-y-2">
              {filteredSnippets.length === 0 ? (
                <div className="text-center py-4 text-muted-foreground">
                  No snippets found matching your criteria.
                </div>
              ) : (
                filteredSnippets.map((snippet) => {
                  const Icon = CATEGORY_ICONS[snippet.category]
                  return (
                    <Card key={snippet.id} className="p-3">
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <Icon className="h-4 w-4" />
                            <span className="font-medium text-sm">{snippet.name}</span>
                            <Badge className={`text-xs ${CATEGORY_COLORS[snippet.category]}`}>
                              {snippet.category.replace('_', ' ')}
                            </Badge>
                          </div>
                          <div className="flex gap-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => copyToClipboard(snippet.content)}
                              className="h-6 w-6 p-0"
                            >
                              <Copy className="h-3 w-3" />
                            </Button>
                            <Button
                              variant="default"
                              size="sm"
                              onClick={() => handleInsertSnippet(snippet)}
                              className="h-6 px-2 text-xs"
                            >
                              <Plus className="h-3 w-3 mr-1" />
                              Insert
                            </Button>
                          </div>
                        </div>
                        <p className="text-xs text-muted-foreground">{snippet.description}</p>
                        {snippet.variables && snippet.variables.length > 0 && (
                          <div className="flex flex-wrap gap-1">
                            {snippet.variables.map((variable) => (
                              <Badge key={variable} variant="outline" className="text-xs">
                                {variable}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </div>
                    </Card>
                  )
                })
              )}
            </div>
          </ScrollArea>
        </div>
      </CardContent>
    </Card>
  )
}
