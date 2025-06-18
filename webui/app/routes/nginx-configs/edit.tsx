import { useState, useEffect } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Save, FileText, Eye, AlertTriangle, CheckCircle, Settings, Code, History, Play, Download, Upload } from 'lucide-react'
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
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '~/components/ui/dialog'
import {
  nginxConfigsApi,
  type UpdateConfigRequest,
  type ConfigType,
  type ValidationResult,
  type NginxConfig,
  type ConfigVersion
} from '~/services/api/nginx-configs'

interface ConfigFormData {
  name: string
  description: string
  type: ConfigType
  content: string
  file_path: string
  is_active: boolean
}

export default function EditNginxConfigPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [validation, setValidation] = useState<ValidationResult | null>(null)
  const [isValidating, setIsValidating] = useState(false)
  const [showHistory, setShowHistory] = useState(false)

  const configId = parseInt(id as string)

  const { register, handleSubmit, setValue, watch, formState: { errors }, reset } = useForm<ConfigFormData>()

  const watchedContent = watch('content')

  // Load configuration
  const { data: config, isLoading, error } = useQuery({
    queryKey: ['nginx-config', configId],
    queryFn: () => nginxConfigsApi.get(configId),
    enabled: !!configId
  })

  // Load configuration history
  const { data: history } = useQuery({
    queryKey: ['nginx-config-history', configId],
    queryFn: () => nginxConfigsApi.getHistory(configId),
    enabled: !!configId
  })

  // Update configuration mutation
  const updateConfigMutation = useMutation({
    mutationFn: (data: UpdateConfigRequest) => nginxConfigsApi.update(configId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-configs'] })
      queryClient.invalidateQueries({ queryKey: ['nginx-config', configId] })
      navigate('/nginx-configs')
    }
  })

  // Deploy configuration mutation
  const deployMutation = useMutation({
    mutationFn: () => nginxConfigsApi.deploy(configId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-config', configId] })
    }
  })

  // Create backup mutation
  const backupMutation = useMutation({
    mutationFn: (reason: string) => nginxConfigsApi.createBackup(configId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-config-history', configId] })
    }
  })

  // Restore configuration mutation
  const restoreMutation = useMutation({
    mutationFn: (version: number) => nginxConfigsApi.restore(configId, version),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-config', configId] })
      queryClient.invalidateQueries({ queryKey: ['nginx-config-history', configId] })
    }
  })

  // Initialize form with loaded data
  useEffect(() => {
    if (config) {
      reset({
        name: config.name,
        description: config.description || '',
        type: config.type,
        content: config.content,
        file_path: config.file_path || '',
        is_active: config.is_active
      })

      // Set initial validation state
      setValidation({
        is_valid: config.is_valid,
        errors: config.validation_logs ? [config.validation_logs] : [],
        output: config.validation_time ? `Last validated: ${new Date(config.validation_time).toLocaleString()}` : ''
      })
    }
  }, [config, reset])

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

  // Auto-validate on content change
  useEffect(() => {
    const timer = setTimeout(() => {
      if (watchedContent && watchedContent.length > 10 && watchedContent !== config?.content) {
        validateConfig(watchedContent)
      }
    }, 1000)

    return () => clearTimeout(timer)
  }, [watchedContent, config?.content])

  const onSubmit = (data: ConfigFormData) => {
    updateConfigMutation.mutate(data)
  }

  const handleDeploy = () => {
    if (config && config.is_valid) {
      deployMutation.mutate()
    }
  }

  const handleBackup = () => {
    const reason = prompt('Enter backup reason (optional):') || 'Manual backup'
    backupMutation.mutate(reason)
  }

  const handleRestore = (version: ConfigVersion) => {
    if (confirm(`Are you sure you want to restore to version ${version.version}? This will overwrite the current configuration.`)) {
      restoreMutation.mutate(version.version)
    }
  }

  const handleLoadVersion = (version: ConfigVersion) => {
    setValue('content', version.content)
    setShowHistory(false)
  }

  if (isLoading) {
    return (
      <div className="container mx-auto py-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    )
  }

  if (error || !config) {
    return (
      <div className="container mx-auto py-6">
        <Alert variant="destructive">
          <AlertDescription>
            Failed to load configuration. Please try again.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center gap-4 mb-6">
        <Button variant="ghost" size="sm" asChild>
          <Link to="/nginx-configs">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex-1">
          <h1 className="text-3xl font-bold tracking-tight">Edit Configuration: {config.name}</h1>
          <p className="text-muted-foreground">
            Modify nginx configuration with validation and deployment
          </p>
        </div>
        <div className="flex gap-2">
          <Dialog open={showHistory} onOpenChange={setShowHistory}>
            <DialogTrigger asChild>
              <Button variant="outline">
                <History className="mr-2 h-4 w-4" />
                History
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-4xl">
              <DialogHeader>
                <DialogTitle>Configuration History</DialogTitle>
                <DialogDescription>
                  View and restore previous versions of this configuration
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 max-h-96 overflow-y-auto">
                {history?.map((version) => (
                  <div key={version.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div>
                        <div className="font-medium">Version {version.version}</div>
                        <div className="text-sm text-muted-foreground">
                          {new Date(version.created_at).toLocaleString()} by {version.created_by_user?.name}
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleLoadVersion(version)}
                        >
                          <Code className="mr-2 h-4 w-4" />
                          Load
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleRestore(version)}
                        >
                          <Download className="mr-2 h-4 w-4" />
                          Restore
                        </Button>
                      </div>
                    </div>
                    {version.comment && (
                      <p className="text-sm text-muted-foreground mb-2">{version.comment}</p>
                    )}
                    <pre className="text-xs bg-muted p-2 rounded overflow-x-auto max-h-32">
                      {version.content}
                    </pre>
                  </div>
                ))}
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Form */}
          <div className="lg:col-span-2 space-y-6">
            {/* Configuration Status */}
            <Card>
              <CardHeader>
                <CardTitle>Configuration Status</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-4">
                  <Badge variant={config.is_active ? 'default' : 'secondary'}>
                    {config.is_active ? 'Active' : 'Inactive'}
                  </Badge>
                  <Badge variant={config.is_valid ? 'default' : 'destructive'}>
                    {config.is_valid ? 'Valid' : 'Invalid'}
                  </Badge>
                  <Badge variant="outline">{config.type}</Badge>
                  {config.is_read_only && (
                    <Badge variant="secondary">Read Only</Badge>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Basic Information */}
            <Card>
              <CardHeader>
                <CardTitle>Basic Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="name">Name *</Label>
                    <Input
                      id="name"
                      {...register('name', { required: 'Name is required' })}
                      disabled={config.is_read_only}
                    />
                    {errors.name && (
                      <p className="text-sm text-destructive mt-1">{errors.name.message}</p>
                    )}
                  </div>
                  <div>
                    <Label htmlFor="type">Type *</Label>
                    <Select
                      value={watch('type')}
                      onValueChange={(value) => setValue('type', value as ConfigType)}
                      disabled={config.is_read_only}
                    >
                      <SelectTrigger>
                        <SelectValue />
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
                    disabled={config.is_read_only}
                  />
                </div>
                <div>
                  <Label htmlFor="file_path">File Path</Label>
                  <Input
                    id="file_path"
                    {...register('file_path')}
                    disabled={config.is_read_only}
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="is_active"
                    checked={watch('is_active')}
                    onCheckedChange={(checked) => setValue('is_active', checked)}
                    disabled={config.is_read_only}
                  />
                  <Label htmlFor="is_active">Active Configuration</Label>
                </div>
              </CardContent>
            </Card>

            {/* Configuration Content */}
            <Card>
              <CardHeader>
                <CardTitle>Configuration Content</CardTitle>
              </CardHeader>
              <CardContent>
                <div>
                  <Label htmlFor="content">Configuration Content *</Label>
                  <Textarea
                    id="content"
                    {...register('content', { required: 'Configuration content is required' })}
                    className="min-h-[400px] font-mono text-sm"
                    disabled={config.is_read_only}
                  />
                  {errors.content && (
                    <p className="text-sm text-destructive mt-1">{errors.content.message}</p>
                  )}
                </div>
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
                    Configuration validation status
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
                {!config.is_read_only && (
                  <Button
                    type="submit"
                    className="w-full"
                    disabled={updateConfigMutation.isPending}
                  >
                    <Save className="mr-2 h-4 w-4" />
                    {updateConfigMutation.isPending ? 'Saving...' : 'Save Changes'}
                  </Button>
                )}

                {config.is_valid && (
                  <Button
                    type="button"
                    variant="outline"
                    className="w-full"
                    onClick={handleDeploy}
                    disabled={deployMutation.isPending}
                  >
                    <Play className="mr-2 h-4 w-4" />
                    {deployMutation.isPending ? 'Deploying...' : 'Deploy Configuration'}
                  </Button>
                )}

                <Button
                  type="button"
                  variant="outline"
                  className="w-full"
                  onClick={handleBackup}
                  disabled={backupMutation.isPending}
                >
                  <Download className="mr-2 h-4 w-4" />
                  {backupMutation.isPending ? 'Creating...' : 'Create Backup'}
                </Button>

                <Button type="button" variant="outline" className="w-full" asChild>
                  <Link to="/nginx-configs">
                    Cancel
                  </Link>
                </Button>
              </CardContent>
            </Card>

            {/* Configuration Info */}
            <Card>
              <CardHeader>
                <CardTitle>Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div>
                  <Label className="font-medium">Created</Label>
                  <p className="text-muted-foreground">{new Date(config.created_at).toLocaleString()}</p>
                </div>
                <div>
                  <Label className="font-medium">Last Modified</Label>
                  <p className="text-muted-foreground">{new Date(config.updated_at).toLocaleString()}</p>
                </div>
                {config.validation_time && (
                  <div>
                    <Label className="font-medium">Last Validated</Label>
                    <p className="text-muted-foreground">{new Date(config.validation_time).toLocaleString()}</p>
                  </div>
                )}
                {config.user && (
                  <div>
                    <Label className="font-medium">Owner</Label>
                    <p className="text-muted-foreground">{config.user.name}</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </form>
    </div>
  )
}
