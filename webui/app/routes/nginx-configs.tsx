import { Link, useLoaderData } from 'react-router'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, FileText, Search, Filter, Settings, Play, Pause, Edit, Trash2, Eye, Download, Upload } from 'lucide-react'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Input } from '~/components/ui/input'
import { Badge } from '~/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '~/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '~/components/ui/table'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '~/components/ui/dropdown-menu'
import { Alert, AlertDescription } from '~/components/ui/alert'
import { Skeleton } from '~/components/ui/skeleton'
import { nginxConfigsApi, type NginxConfig, type ConfigType, type ConfigStatus, getConfigStatusVariant, getConfigTypeLabel } from '~/services/api/nginx-configs'

export default function NginxConfigsPage() {
  const [searchTerm, setSearchTerm] = useState('')
  const [typeFilter, setTypeFilter] = useState<ConfigType | 'all'>('all')
  const [statusFilter, setStatusFilter] = useState<ConfigStatus | 'all'>('all')
  const [page, setPage] = useState(1)
  const limit = 10

  // Load configurations with filters
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['nginx-configs', page, limit, typeFilter === 'all' ? undefined : typeFilter, searchTerm],
    queryFn: () => nginxConfigsApi.list({
      page,
      limit,
      type: typeFilter === 'all' ? undefined : typeFilter
    }),
    placeholderData: (prev) => prev
  })

  // Filter configurations by search term and status locally
  const filteredConfigs = data?.configs?.filter((config) => {
    const matchesSearch = config.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         config.description?.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = statusFilter === 'all' || config.status === statusFilter
    return matchesSearch && matchesStatus
  }) || []

  const handleDeploy = async (configId: number) => {
    try {
      await nginxConfigsApi.deploy(configId)
      refetch()
    } catch (error) {
      console.error('Failed to deploy configuration:', error)
    }
  }

  const handleDelete = async (configId: number) => {
    if (confirm('Are you sure you want to delete this configuration?')) {
      try {
        await nginxConfigsApi.delete(configId)
        refetch()
      } catch (error) {
        console.error('Failed to delete configuration:', error)
      }
    }
  }

  const handleCreateBackup = async (configId: number) => {
    try {
      await nginxConfigsApi.createBackup(configId, 'Manual backup')
      refetch()
    } catch (error) {
      console.error('Failed to create backup:', error)
    }
  }

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <Alert variant="destructive">
          <AlertDescription>
            Failed to load nginx configurations. Please try again.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Nginx Configurations</h1>
          <p className="text-muted-foreground">
            Manage your nginx configuration files with validation and deployment
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" asChild>
            <Link to="/nginx-templates">
              <FileText className="mr-2 h-4 w-4" />
              Templates
            </Link>
          </Button>
          <Button asChild>
            <Link to="/nginx-configs/new">
              <Plus className="mr-2 h-4 w-4" />
              New Configuration
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
                  placeholder="Search configurations..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
            <Select value={typeFilter} onValueChange={(value) => setTypeFilter(value as ConfigType | 'all')}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="main">Main Config</SelectItem>
                <SelectItem value="server">Server Block</SelectItem>
                <SelectItem value="upstream">Upstream</SelectItem>
                <SelectItem value="location">Location</SelectItem>
                <SelectItem value="custom">Custom</SelectItem>
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as ConfigStatus | 'all')}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="inactive">Inactive</SelectItem>
                <SelectItem value="error">Error</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Configurations Table */}
      <Card>
        <CardHeader>
          <CardTitle>Configurations ({data?.total || 0})</CardTitle>
          <CardDescription>
            Manage nginx configuration files and templates
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : filteredConfigs.length === 0 ? (
            <div className="text-center py-8">
              <FileText className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No configurations found</h3>
              <p className="mt-1 text-sm text-gray-500">
                {searchTerm || typeFilter !== 'all' || statusFilter !== 'all'
                  ? 'Try adjusting your filters or search terms.'
                  : 'Get started by creating a new nginx configuration.'
                }
              </p>
              <div className="mt-6">
                <Button asChild>
                  <Link to="/nginx-configs/new">
                    <Plus className="mr-2 h-4 w-4" />
                    New Configuration
                  </Link>
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Valid</TableHead>
                    <TableHead>Last Modified</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredConfigs.map((config) => (
                    <TableRow key={config.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{config.name}</div>
                          {config.description && (
                            <div className="text-sm text-muted-foreground">{config.description}</div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">
                          {getConfigTypeLabel(config.type)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={getConfigStatusVariant(config.status)}>
                          {config.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={config.is_valid ? 'default' : 'destructive'}>
                          {config.is_valid ? 'Valid' : 'Invalid'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {new Date(config.updated_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <Settings className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem asChild>
                              <Link to={`/nginx-configs/${config.id}`}>
                                <Eye className="mr-2 h-4 w-4" />
                                View
                              </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem asChild>
                              <Link to={`/nginx-configs/${config.id}/edit`}>
                                <Edit className="mr-2 h-4 w-4" />
                                Edit
                              </Link>
                            </DropdownMenuItem>
                            {config.is_valid && (
                              <DropdownMenuItem onClick={() => handleDeploy(config.id)}>
                                <Play className="mr-2 h-4 w-4" />
                                Deploy
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuItem onClick={() => handleCreateBackup(config.id)}>
                              <Download className="mr-2 h-4 w-4" />
                              Backup
                            </DropdownMenuItem>
                            {!config.is_read_only && (
                              <DropdownMenuItem
                                onClick={() => handleDelete(config.id)}
                                className="text-destructive"
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                Delete
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Pagination */}
              {data && data.total > limit && (
                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">
                    Showing {((page - 1) * limit) + 1} to {Math.min(page * limit, data.total)} of {data.total} configurations
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
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
