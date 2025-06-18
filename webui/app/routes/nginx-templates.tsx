import { Link } from 'react-router'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, FileText, Search, Filter, Settings, Edit, Trash2, Eye, Download, Upload, Code, Copy } from 'lucide-react'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Input } from '~/components/ui/input'
import { Badge } from '~/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '~/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '~/components/ui/table'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '~/components/ui/dropdown-menu'
import { Alert, AlertDescription } from '~/components/ui/alert'
import { Skeleton } from '~/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '~/components/ui/dialog'
import { Textarea } from '~/components/ui/textarea'
import { Label } from '~/components/ui/label'
import { nginxConfigsApi, type ConfigTemplate, type TemplateCategory, getTemplateCategoryLabel } from '~/services/api/nginx-configs'

export default function NginxTemplatesPage() {
  const [searchTerm, setSearchTerm] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<TemplateCategory | 'all'>('all')
  const [page, setPage] = useState(1)
  const [selectedTemplate, setSelectedTemplate] = useState<ConfigTemplate | null>(null)
  const [showPreview, setShowPreview] = useState(false)
  const limit = 10

  const queryClient = useQueryClient()

  // Load templates with filters
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['nginx-templates', page, limit, categoryFilter === 'all' ? undefined : categoryFilter],
    queryFn: () => nginxConfigsApi.listTemplates({
      page,
      limit,
      category: categoryFilter === 'all' ? undefined : categoryFilter,
      include_public: true
    }),
    placeholderData: (prev) => prev
  })

  // Filter templates by search term locally
  const filteredTemplates = data?.templates?.filter((template) => {
    const matchesSearch = template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         template.description?.toLowerCase().includes(searchTerm.toLowerCase())
    return matchesSearch
  }) || []

  // Delete template mutation
  const deleteTemplateMutation = useMutation({
    mutationFn: (templateId: number) => nginxConfigsApi.deleteTemplate(templateId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-templates'] })
    }
  })

  // Initialize built-in templates mutation
  const initTemplatesMutation = useMutation({
    mutationFn: () => nginxConfigsApi.initBuiltInTemplates(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-templates'] })
    }
  })

  const handleDelete = async (templateId: number, templateName: string) => {
    if (confirm(`Are you sure you want to delete template "${templateName}"?`)) {
      try {
        await deleteTemplateMutation.mutateAsync(templateId)
      } catch (error) {
        console.error('Failed to delete template:', error)
      }
    }
  }

  const handlePreview = (template: ConfigTemplate) => {
    setSelectedTemplate(template)
    setShowPreview(true)
  }

  const handleCopyTemplate = (template: ConfigTemplate) => {
    navigator.clipboard.writeText(template.content)
    // You might want to show a toast notification here
  }

  const getCategoryVariant = (category: TemplateCategory) => {
    switch (category) {
      case 'proxy': return 'default'
      case 'ssl': return 'secondary'
      case 'load_balance': return 'outline'
      case 'cache': return 'destructive'
      case 'security': return 'default'
      default: return 'secondary'
    }
  }

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <Alert variant="destructive">
          <AlertDescription>
            Failed to load nginx templates. Please try again.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Nginx Templates</h1>
          <p className="text-muted-foreground">
            Browse and manage nginx configuration templates
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => initTemplatesMutation.mutate()}
            disabled={initTemplatesMutation.isPending}
          >
            <Download className="mr-2 h-4 w-4" />
            {initTemplatesMutation.isPending ? 'Initializing...' : 'Init Built-in Templates'}
          </Button>
          <Button asChild>
            <Link to="/nginx-templates/new">
              <Plus className="mr-2 h-4 w-4" />
              New Template
            </Link>
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search templates..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
            <Select value={categoryFilter} onValueChange={(value) => setCategoryFilter(value as TemplateCategory | 'all')}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by category" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Categories</SelectItem>
                <SelectItem value="proxy">Proxy</SelectItem>
                <SelectItem value="ssl">SSL/TLS</SelectItem>
                <SelectItem value="load_balance">Load Balance</SelectItem>
                <SelectItem value="cache">Cache</SelectItem>
                <SelectItem value="security">Security</SelectItem>
                <SelectItem value="custom">Custom</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Templates Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-6">
        {isLoading ? (
          [...Array(6)].map((_, i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <Skeleton className="h-4 w-3/4 mb-2" />
                <Skeleton className="h-3 w-full mb-4" />
                <Skeleton className="h-20 w-full" />
              </CardContent>
            </Card>
          ))
        ) : filteredTemplates.length === 0 ? (
          <div className="col-span-full text-center py-12">
            <FileText className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No templates found</h3>
            <p className="mt-1 text-sm text-gray-500">
              {searchTerm || categoryFilter !== 'all'
                ? 'Try adjusting your filters or search terms.'
                : 'Get started by creating a new template or initializing built-in templates.'
              }
            </p>
            <div className="mt-6 space-x-2">
              <Button asChild>
                <Link to="/nginx-templates/new">
                  <Plus className="mr-2 h-4 w-4" />
                  New Template
                </Link>
              </Button>
              <Button variant="outline" onClick={() => initTemplatesMutation.mutate()}>
                <Download className="mr-2 h-4 w-4" />
                Init Built-in Templates
              </Button>
            </div>
          </div>
        ) : (
          filteredTemplates.map((template) => (
            <Card key={template.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <CardTitle className="text-lg">{template.name}</CardTitle>
                    <div className="flex items-center gap-2">
                      <Badge variant={getCategoryVariant(template.category)}>
                        {getTemplateCategoryLabel(template.category)}
                      </Badge>
                      {template.is_built_in && (
                        <Badge variant="outline">Built-in</Badge>
                      )}
                      {template.is_public && (
                        <Badge variant="secondary">Public</Badge>
                      )}
                    </div>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <Settings className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => handlePreview(template)}>
                        <Eye className="mr-2 h-4 w-4" />
                        Preview
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => handleCopyTemplate(template)}>
                        <Copy className="mr-2 h-4 w-4" />
                        Copy Content
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild>
                        <Link to={`/nginx-configs/new?template=${template.id}`}>
                          <Plus className="mr-2 h-4 w-4" />
                          Use Template
                        </Link>
                      </DropdownMenuItem>
                      {!template.is_built_in && (
                        <>
                          <DropdownMenuItem asChild>
                            <Link to={`/nginx-templates/${template.id}`}>
                              <Edit className="mr-2 h-4 w-4" />
                              Edit
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(template.id, template.name)}
                            className="text-destructive"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </>
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent>
                {template.description && (
                  <p className="text-sm text-muted-foreground mb-4">{template.description}</p>
                )}
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Usage Count:</span>
                    <span>{template.usage_count}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Variables:</span>
                    <span>{template.variables ? Object.keys(template.variables).length : 0}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Created:</span>
                    <span>{new Date(template.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                <div className="mt-4 space-x-2">
                  <Button size="sm" onClick={() => handlePreview(template)}>
                    <Eye className="mr-2 h-4 w-4" />
                    Preview
                  </Button>
                  <Button size="sm" variant="outline" asChild>
                    <Link to={`/nginx-configs/new?template=${template.id}`}>
                      <Plus className="mr-2 h-4 w-4" />
                      Use
                    </Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {/* Pagination */}
      {data && data.total > limit && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            Showing {((page - 1) * limit) + 1} to {Math.min(page * limit, data.total)} of {data.total} templates
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => p + 1)}
              disabled={page * limit >= data.total}
            >
              Next
            </Button>
          </div>
        </div>
      )}

      {/* Template Preview Dialog */}
      <Dialog open={showPreview} onOpenChange={setShowPreview}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>Template Preview: {selectedTemplate?.name}</DialogTitle>
            <DialogDescription>
              {selectedTemplate?.description}
            </DialogDescription>
          </DialogHeader>
          {selectedTemplate && (
            <div className="space-y-4">
              <div className="flex flex-wrap gap-2">
                <Badge variant={getCategoryVariant(selectedTemplate.category)}>
                  {getTemplateCategoryLabel(selectedTemplate.category)}
                </Badge>
                {selectedTemplate.is_built_in && (
                  <Badge variant="outline">Built-in</Badge>
                )}
                {selectedTemplate.is_public && (
                  <Badge variant="secondary">Public</Badge>
                )}
              </div>

              {selectedTemplate.variables && Object.keys(selectedTemplate.variables).length > 0 && (
                <div>
                  <Label className="text-sm font-medium">Template Variables:</Label>
                  <div className="mt-2 grid grid-cols-2 gap-2">
                    {Object.entries(selectedTemplate.variables).map(([key, config]: [string, any]) => (
                      <div key={key} className="text-sm">
                        <span className="font-mono bg-muted px-1 rounded">{key}</span>
                        {config.description && (
                          <span className="text-muted-foreground ml-2">- {config.description}</span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div>
                <Label className="text-sm font-medium">Template Content:</Label>
                <Textarea
                  value={selectedTemplate.content}
                  readOnly
                  className="mt-2 min-h-[400px] font-mono text-sm"
                />
              </div>

              <div className="flex gap-2">
                <Button onClick={() => handleCopyTemplate(selectedTemplate)}>
                  <Copy className="mr-2 h-4 w-4" />
                  Copy Content
                </Button>
                <Button variant="outline" asChild>
                  <Link to={`/nginx-configs/new?template=${selectedTemplate.id}`}>
                    <Plus className="mr-2 h-4 w-4" />
                    Use Template
                  </Link>
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
