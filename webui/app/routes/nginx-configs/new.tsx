import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Save, FileText, Eye, AlertTriangle, CheckCircle, Settings, Code } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import { Textarea } from '~/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '~/components/ui/select'
import { Switch } from '~/components/ui/switch'
import { Badge } from '~/components/ui/badge'
import { Alert, AlertDescription } from '~/components/ui/alert'
import { Separator } from '~/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '~/components/ui/tabs'
import {
  nginxConfigsApi,
  type CreateConfigRequest,
  type ConfigType,
  type ValidationResult,
  type ConfigTemplate,
  type TemplateRenderResponse
} from '~/services/api/nginx-configs'

interface ConfigFormData {
  name: string
  description: string
  type: ConfigType
  content: string
  file_path: string
  is_active: boolean
  template_id?: number
  template_vars?: Record<string, any>
}

export default function NewNginxConfigPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [validation, setValidation] = useState<ValidationResult | null>(null)
  const [isValidating, setIsValidating] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<ConfigTemplate | null>(null)
  const [templateVars, setTemplateVars] = useState<Record<string, string>>({})
  const [renderedContent, setRenderedContent] = useState<string>('')

  const { register, handleSubmit, setValue, watch, formState: { errors } } = useForm<ConfigFormData>({
    defaultValues: {
      name: '',
      description: '',
      type: 'server',
      content: '',
      file_path: '',
      is_active: false,
      template_vars: {}
    }
  })

  const watchedContent = watch('content')
  const watchedType = watch('type')

  // Load templates
  const { data: templatesData } = useQuery({
    queryKey: ['nginx-templates'],
    queryFn: () => nginxConfigsApi.listTemplates({ include_public: true })
  })

  // Create configuration mutation
  const createConfigMutation = useMutation({
    mutationFn: (data: CreateConfigRequest) => nginxConfigsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-configs'] })
      navigate('/nginx-configs')
    }
  })

  // Validate configuration
  const validateConfig = async (content: string) => {
    if (!content.trim()) {
      setValidation(null)
      return
    }

    setIsValidating(true)
    try {
      const result = await nginxConfigsApi.validate({ content })
      setValidation(result)
    } catch (error) {
      setValidation({
        is_valid: false,
        errors: ['Failed to validate configuration'],
        output: ''
      })
    } finally {
      setIsValidating(false)
    }
  }

  // Render template with variables
  const renderTemplate = async (templateId: number, variables: Record<string, string>) => {
    try {
      const result = await nginxConfigsApi.renderTemplate(templateId, { variables })
      setRenderedContent(result.content)
      setValue('content', result.content)
      if (result.is_valid) {
        setValidation({ is_valid: true, errors: [], output: 'Template rendered successfully' })
      } else {
        setValidation({ is_valid: false, errors: result.errors || [], output: '' })
      }
    } catch (error) {
      console.error('Failed to render template:', error)
    }
  }

  // Auto-validate on content change
  useEffect(() => {
    const timer = setTimeout(() => {
      if (watchedContent && watchedContent.length > 10) {
        validateConfig(watchedContent)
      }
    }, 1000)

    return () => clearTimeout(timer)
  }, [watchedContent])

  // Handle template selection
  const handleTemplateSelect = (templateId: string) => {
    const template = templatesData?.templates.find(t => t.id === parseInt(templateId))
    if (template) {
      setSelectedTemplate(template)
      setValue('template_id', template.id)
      setValue('type', getTemplateConfigType(template.category))

      // Initialize template variables
      const vars: Record<string, string> = {}
      if (template.variables) {
        Object.keys(template.variables).forEach(key => {
          vars[key] = ''
        })
      }
      setTemplateVars(vars)
    }
  }

  // Handle template variable change
  const handleTemplateVarChange = (key: string, value: string) => {
    const newVars = { ...templateVars, [key]: value }
    setTemplateVars(newVars)

    if (selectedTemplate && Object.values(newVars).every(v => v.trim() !== '')) {
      renderTemplate(selectedTemplate.id, newVars)
    }
  }

  // Get config type based on template category
  const getTemplateConfigType = (category: string): ConfigType => {
    switch (category) {
      case 'proxy':
      case 'ssl':
        return 'server'
      case 'load_balance':
        return 'upstream'
      default:
        return 'custom'
    }
  }

  const onSubmit = (data: ConfigFormData) => {
    const submitData: CreateConfigRequest = {
      ...data,
      template_vars: selectedTemplate ? templateVars : undefined
    }
    createConfigMutation.mutate(submitData)
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center gap-4 mb-6">
        <Button variant="ghost" size="sm" asChild>
          <Link to="/nginx-configs">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">New Nginx Configuration</h1>
          <p className="text-muted-foreground">
            Create a new nginx configuration file with validation
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Form */}
          <div className="lg:col-span-2 space-y-6">
            {/* Basic Information */}
            <Card>
              <CardHeader>
                <CardTitle>Basic Information</CardTitle>
                <CardDescription>
                  Configure the basic settings for your nginx configuration
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="name">Name *</Label>
                    <Input
                      id="name"
                      {...register('name', { required: 'Name is required' })}
                      placeholder="e.g., my-api-proxy"
                    />
                    {errors.name && (
                      <p className="text-sm text-destructive mt-1">{errors.name.message}</p>
                    )}
                  </div>
                  <div>
                    <Label htmlFor="type">Type *</Label>
                    <Select value={watchedType} onValueChange={(value) => setValue('type', value as ConfigType)}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select configuration type" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="main">Main Config</SelectItem>
                        <SelectItem value="server">Server Block</SelectItem>
                        <SelectItem value="upstream">Upstream</SelectItem>
                        <SelectItem value="location">Location</SelectItem>
                        <SelectItem value="custom">Custom</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div>
                  <Label htmlFor="description">Description</Label>
                  <Input
                    id="description"
                    {...register('description')}
                    placeholder="Optional description"
                  />
                </div>
                <div>
                  <Label htmlFor="file_path">File Path</Label>
                  <Input
                    id="file_path"
                    {...register('file_path')}
                    placeholder="e.g., /etc/nginx/sites-available/my-site"
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="is_active"
                    checked={watch('is_active')}
                    onCheckedChange={(checked) => setValue('is_active', checked)}
                  />
                  <Label htmlFor="is_active">Active Configuration</Label>
                </div>
              </CardContent>
            </Card>

            {/* Configuration Content */}
            <Card>
              <CardHeader>
                <CardTitle>Configuration Content</CardTitle>
                <CardDescription>
                  Write your nginx configuration or use a template
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Tabs defaultValue="editor" className="w-full">
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="editor">Editor</TabsTrigger>
                    <TabsTrigger value="template">Template</TabsTrigger>
                  </TabsList>

                  <TabsContent value="editor" className="space-y-4">
                    <div>
                      <Label htmlFor="content">Configuration Content *</Label>
                      <Textarea
                        id="content"
                        {...register('content', { required: 'Configuration content is required' })}
                        placeholder="Enter your nginx configuration..."
                        className="min-h-[400px] font-mono text-sm"
                      />
                      {errors.content && (
                        <p className="text-sm text-destructive mt-1">{errors.content.message}</p>
                      )}
                    </div>
                  </TabsContent>

                  <TabsContent value="template" className="space-y-4">
                    <div>
                      <Label htmlFor="template">Select Template</Label>
                      <Select onValueChange={handleTemplateSelect}>
                        <SelectTrigger>
                          <SelectValue placeholder="Choose a configuration template" />
                        </SelectTrigger>
                        <SelectContent>
                          {templatesData?.templates.map((template) => (
                            <SelectItem key={template.id} value={template.id.toString()}>
                              <div className="flex items-center gap-2">
                                <Badge variant="secondary">{template.category}</Badge>
                                {template.name}
                              </div>
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    {selectedTemplate && selectedTemplate.variables && (
                      <div className="space-y-4">
                        <Label>Template Variables</Label>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          {Object.entries(selectedTemplate.variables).map(([key, config]: [string, any]) => (
                            <div key={key}>
                              <Label htmlFor={key}>{config.label || key}</Label>
                              <Input
                                id={key}
                                value={templateVars[key] || ''}
                                onChange={(e) => handleTemplateVarChange(key, e.target.value)}
                                placeholder={config.placeholder || `Enter ${key}`}
                              />
                              {config.description && (
                                <p className="text-sm text-muted-foreground mt-1">{config.description}</p>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {renderedContent && (
                      <div>
                        <Label>Generated Configuration</Label>
                        <Textarea
                          value={renderedContent}
                          readOnly
                          className="min-h-[300px] font-mono text-sm bg-muted"
                        />
                      </div>
                    )}
                  </TabsContent>
                </Tabs>
              </CardContent>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Validation Status */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Settings className="h-4 w-4" />
                  Validation
                </CardTitle>
              </CardHeader>
              <CardContent>
                {isValidating ? (
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-gray-300 border-t-blue-600 rounded-full animate-spin" />
                    <span className="text-sm">Validating...</span>
                  </div>
                ) : validation ? (
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      {validation.is_valid ? (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      ) : (
                        <AlertTriangle className="h-4 w-4 text-red-600" />
                      )}
                      <span className="font-medium">
                        {validation.is_valid ? 'Valid Configuration' : 'Invalid Configuration'}
                      </span>
                    </div>
                    {validation.errors.length > 0 && (
                      <div className="space-y-1">
                        {validation.errors.map((error, index) => (
                          <Alert key={index} variant="destructive">
                            <AlertDescription className="text-sm">{error}</AlertDescription>
                          </Alert>
                        ))}
                      </div>
                    )}
                    {validation.output && (
                      <p className="text-sm text-muted-foreground">{validation.output}</p>
                    )}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    Start typing to validate your configuration
                  </p>
                )}
              </CardContent>
            </Card>

            {/* Actions */}
            <Card>
              <CardHeader>
                <CardTitle>Actions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  type="submit"
                  className="w-full"
                  disabled={createConfigMutation.isPending}
                >
                  <Save className="mr-2 h-4 w-4" />
                  {createConfigMutation.isPending ? 'Saving...' : 'Save Configuration'}
                </Button>
                <Button type="button" variant="outline" className="w-full" asChild>
                  <Link to="/nginx-configs">
                    Cancel
                  </Link>
                </Button>
              </CardContent>
            </Card>

            {/* Template Info */}
            {selectedTemplate && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <FileText className="h-4 w-4" />
                    Template Info
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div>
                    <Label className="text-sm font-medium">Name</Label>
                    <p className="text-sm">{selectedTemplate.name}</p>
                  </div>
                  {selectedTemplate.description && (
                    <div>
                      <Label className="text-sm font-medium">Description</Label>
                      <p className="text-sm text-muted-foreground">{selectedTemplate.description}</p>
                    </div>
                  )}
                  <div>
                    <Label className="text-sm font-medium">Category</Label>
                    <Badge variant="secondary">{selectedTemplate.category}</Badge>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </form>
    </div>
  )
}
