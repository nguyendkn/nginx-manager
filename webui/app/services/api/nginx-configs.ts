import { apiClient } from './client'

// Types
export interface NginxConfig {
  id: number
  name: string
  description?: string
  type: ConfigType
  status: ConfigStatus
  content: string
  file_path?: string
  is_active: boolean
  is_read_only: boolean
  user_id: number
  is_valid: boolean
  validation_time: string
  validation_logs?: string
  template_id?: number
  template_vars?: Record<string, any>
  created_at: string
  updated_at: string
  user?: {
    id: number
    name: string
    email: string
  }
  versions?: ConfigVersion[]
  template?: ConfigTemplate
}

export interface ConfigVersion {
  id: number
  config_id: number
  version: number
  content: string
  comment: string
  is_backup: boolean
  created_by: number
  created_at: string
  created_by_user?: {
    id: number
    name: string
  }
}

export interface ConfigTemplate {
  id: number
  name: string
  description?: string
  category: TemplateCategory
  content: string
  variables?: Record<string, any>
  is_built_in: boolean
  is_public: boolean
  usage_count: number
  user_id: number
  created_at: string
  updated_at: string
}

export interface ConfigBackup {
  id: number
  config_id: number
  backup_name: string
  content: string
  file_path: string
  reason: string
  auto_backup: boolean
  created_by: number
  created_at: string
}

export interface ValidationResult {
  is_valid: boolean
  errors: string[]
  output: string
}

export interface TemplateRenderResponse {
  content: string
  is_valid: boolean
  errors?: string[]
}

export type ConfigType = 'main' | 'server' | 'upstream' | 'location' | 'custom'
export type ConfigStatus = 'draft' | 'active' | 'inactive' | 'error'
export type TemplateCategory = 'proxy' | 'load_balance' | 'ssl' | 'cache' | 'security' | 'custom'

// Request types
export interface CreateConfigRequest {
  name: string
  description?: string
  type: ConfigType
  content: string
  file_path?: string
  is_active?: boolean
  template_id?: number
  template_vars?: Record<string, any>
}

export interface UpdateConfigRequest extends CreateConfigRequest {}

export interface CreateTemplateRequest {
  name: string
  description?: string
  category: TemplateCategory
  content: string
  variables?: Record<string, any>
  is_public?: boolean
}

export interface UpdateTemplateRequest extends CreateTemplateRequest {}

export interface RenderTemplateRequest {
  variables: Record<string, any>
}

export interface ValidateConfigRequest {
  content: string
}

// Response types
export interface ConfigListResponse {
  configs: NginxConfig[]
  total: number
  page: number
  limit: number
}

export interface TemplateListResponse {
  templates: ConfigTemplate[]
  total: number
  page: number
  limit: number
}

// API functions for nginx configurations
export const nginxConfigsApi = {
  // Configuration CRUD operations
  async list(params?: {
    page?: number
    limit?: number
    type?: string
  }): Promise<ConfigListResponse> {
    const response = await apiClient.get('/nginx/configs', { params })
    return response.data.data
  },

  async get(id: number): Promise<NginxConfig> {
    const response = await apiClient.get(`/nginx/configs/${id}`)
    return response.data.data
  },

  async create(data: CreateConfigRequest): Promise<NginxConfig> {
    const response = await apiClient.post('/nginx/configs', data)
    return response.data.data
  },

  async update(id: number, data: UpdateConfigRequest): Promise<NginxConfig> {
    const response = await apiClient.put(`/nginx/configs/${id}`, data)
    return response.data.data
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete(`/nginx/configs/${id}`)
  },

  // Configuration operations
  async validate(data: ValidateConfigRequest): Promise<ValidationResult> {
    const response = await apiClient.post('/nginx/configs/validate', data)
    return response.data.data
  },

  async deploy(id: number): Promise<void> {
    await apiClient.post(`/nginx/configs/${id}/deploy`)
  },

  async getHistory(id: number): Promise<ConfigVersion[]> {
    const response = await apiClient.get(`/nginx/configs/${id}/history`)
    return response.data.data
  },

  async createBackup(id: number, reason?: string): Promise<void> {
    await apiClient.post(`/nginx/configs/${id}/backup`, { reason })
  },

  async restore(id: number, version: number): Promise<void> {
    await apiClient.post(`/nginx/configs/${id}/restore/${version}`)
  },

  // Template operations
  async listTemplates(params?: {
    page?: number
    limit?: number
    category?: string
    include_public?: boolean
  }): Promise<TemplateListResponse> {
    const response = await apiClient.get('/nginx/templates', { params })
    return response.data.data
  },

  async getTemplate(id: number): Promise<ConfigTemplate> {
    const response = await apiClient.get(`/nginx/templates/${id}`)
    return response.data.data
  },

  async createTemplate(data: CreateTemplateRequest): Promise<ConfigTemplate> {
    const response = await apiClient.post('/nginx/templates', data)
    return response.data.data
  },

  async updateTemplate(id: number, data: UpdateTemplateRequest): Promise<ConfigTemplate> {
    const response = await apiClient.put(`/nginx/templates/${id}`, data)
    return response.data.data
  },

  async deleteTemplate(id: number): Promise<void> {
    await apiClient.delete(`/nginx/templates/${id}`)
  },

  async renderTemplate(id: number, data: RenderTemplateRequest): Promise<TemplateRenderResponse> {
    const response = await apiClient.post(`/nginx/templates/${id}/render`, data)
    return response.data.data
  },

  async getCategories(): Promise<string[]> {
    const response = await apiClient.get('/nginx/templates/categories')
    return response.data.data
  },

  async initBuiltInTemplates(): Promise<void> {
    await apiClient.post('/nginx/templates/init-builtin')
  },
}

// Constants
export const CONFIG_TYPES: { value: ConfigType; label: string }[] = [
  { value: 'main', label: 'Main Configuration' },
  { value: 'server', label: 'Server Block' },
  { value: 'upstream', label: 'Upstream Configuration' },
  { value: 'location', label: 'Location Block' },
  { value: 'custom', label: 'Custom Configuration' },
]

export const CONFIG_STATUSES: { value: ConfigStatus; label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }[] = [
  { value: 'draft', label: 'Draft', variant: 'outline' },
  { value: 'active', label: 'Active', variant: 'default' },
  { value: 'inactive', label: 'Inactive', variant: 'secondary' },
  { value: 'error', label: 'Error', variant: 'destructive' },
]

export const TEMPLATE_CATEGORIES: { value: TemplateCategory; label: string }[] = [
  { value: 'proxy', label: 'Reverse Proxy' },
  { value: 'load_balance', label: 'Load Balancing' },
  { value: 'ssl', label: 'SSL/TLS' },
  { value: 'cache', label: 'Caching' },
  { value: 'security', label: 'Security' },
  { value: 'custom', label: 'Custom' },
]

// Utility functions
export const getConfigStatusVariant = (status: ConfigStatus) => {
  return CONFIG_STATUSES.find(s => s.value === status)?.variant || 'outline'
}

export const getConfigTypeLabel = (type: ConfigType) => {
  return CONFIG_TYPES.find(t => t.value === type)?.label || type
}

export const getTemplateCategoryLabel = (category: TemplateCategory) => {
  return TEMPLATE_CATEGORIES.find(c => c.value === category)?.label || category
}
