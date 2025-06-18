import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { Link } from 'react-router';
import { Plus, Search, Filter, MoreHorizontal, Globe, Shield, CheckCircle, XCircle, Edit, Trash2, Power } from 'lucide-react';
import toast from 'react-hot-toast';

import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { Badge } from '~/components/ui/badge';
import { Checkbox } from '~/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator
} from '~/components/ui/dropdown-menu';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '~/components/ui/select';

import { proxyHostsApi, type ProxyHost, type ProxyHostListParams, type ProxyHostListResponse } from '~/services/api/proxy-hosts';
import { useRequireAuth } from '~/contexts/AuthContext';

export default function ProxyHostsPage() {
  useRequireAuth();

  const queryClient = useQueryClient();
  const [selectedHosts, setSelectedHosts] = useState<number[]>([]);
  const [filters, setFilters] = useState<ProxyHostListParams>({
    page: 1,
    limit: 10,
    search: '',
    enabled: undefined
  });

  // Fetch proxy hosts
  const { data, isLoading, error } = useQuery<ProxyHostListResponse>({
    queryKey: ['proxy-hosts', filters],
    queryFn: () => proxyHostsApi.list(filters),
    placeholderData: keepPreviousData
  });

  // Toggle proxy host mutation
  const toggleMutation = useMutation({
    mutationFn: proxyHostsApi.toggle,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxy-hosts'] });
      toast.success('Proxy host status updated');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to update proxy host');
    }
  });

  // Delete proxy host mutation
  const deleteMutation = useMutation({
    mutationFn: proxyHostsApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxy-hosts'] });
      toast.success('Proxy host deleted');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to delete proxy host');
    }
  });

  // Bulk toggle mutation
  const bulkToggleMutation = useMutation({
    mutationFn: proxyHostsApi.bulkToggle,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['proxy-hosts'] });
      toast.success(`${data.updated} proxy hosts ${data.enabled ? 'enabled' : 'disabled'}`);
      setSelectedHosts([]);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to update proxy hosts');
    }
  });

  const handleToggle = (id: number) => {
    toggleMutation.mutate(id);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure you want to delete this proxy host?')) {
      deleteMutation.mutate(id);
    }
  };

  const handleBulkToggle = (enabled: boolean) => {
    if (selectedHosts.length === 0) return;
    bulkToggleMutation.mutate({ ids: selectedHosts, enabled });
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedHosts(data?.data?.map((host: ProxyHost) => host.id) || []);
    } else {
      setSelectedHosts([]);
    }
  };

  const handleSelectHost = (id: number, checked: boolean) => {
    if (checked) {
      setSelectedHosts(prev => [...prev, id]);
    } else {
      setSelectedHosts(prev => prev.filter(hostId => hostId !== id));
    }
  };

  const formatDomains = (domains: string[]) => {
    if (domains.length === 0) return 'No domains';
    if (domains.length === 1) return domains[0];
    return `${domains[0]} +${domains.length - 1} more`;
  };

  const getStatusColor = (enabled: boolean) => {
    return enabled ? 'text-green-600' : 'text-gray-400';
  };

  const getStatusIcon = (enabled: boolean) => {
    return enabled ? CheckCircle : XCircle;
  };

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <div className="text-center">
              <h3 className="text-lg font-semibold text-red-600">Error loading proxy hosts</h3>
              <p className="text-sm text-muted-foreground mt-2">
                {(error as any)?.response?.data?.message || 'Failed to fetch proxy hosts'}
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Proxy Hosts</h1>
          <p className="text-muted-foreground">
            Manage your nginx proxy host configurations
          </p>
        </div>
        <Button asChild>
          <Link to="/proxy-hosts/new">
            <Plus className="w-4 h-4 mr-2" />
            Add Proxy Host
          </Link>
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <Input
                placeholder="Search proxy hosts..."
                value={filters.search}
                onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value, page: 1 }))}
                className="max-w-md"
              />
            </div>
            <Select
              value={filters.enabled?.toString() || 'all'}
              onValueChange={(value) => setFilters(prev => ({
                ...prev,
                enabled: value === 'all' ? undefined : value === 'true',
                page: 1
              }))}
            >
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="true">Enabled</SelectItem>
                <SelectItem value="false">Disabled</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Bulk Actions */}
      {selectedHosts.length > 0 && (
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                {selectedHosts.length} proxy host{selectedHosts.length !== 1 ? 's' : ''} selected
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleBulkToggle(true)}
                  disabled={bulkToggleMutation.isPending}
                >
                  Enable Selected
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleBulkToggle(false)}
                  disabled={bulkToggleMutation.isPending}
                >
                  Disable Selected
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Proxy Hosts Table */}
      <Card>
        <CardHeader>
          <CardTitle>Proxy Hosts</CardTitle>
          <CardDescription>
            {data?.pagination.total || 0} proxy host{(data?.pagination.total || 0) !== 1 ? 's' : ''} total
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
                <p className="text-sm text-muted-foreground mt-2">Loading proxy hosts...</p>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">
                      <Checkbox
                        checked={selectedHosts.length === data?.data?.length && data?.data?.length > 0}
                        onCheckedChange={handleSelectAll}
                      />
                    </TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Domains</TableHead>
                    <TableHead>Target</TableHead>
                    <TableHead>SSL</TableHead>
                    <TableHead>Access</TableHead>
                    <TableHead className="w-12"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data?.map((host: ProxyHost) => {
                    const StatusIcon = getStatusIcon(host.enabled);
                    return (
                      <TableRow key={host.id}>
                        <TableCell>
                          <Checkbox
                            checked={selectedHosts.includes(host.id)}
                            onCheckedChange={(checked: boolean) => handleSelectHost(host.id, checked)}
                          />
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <StatusIcon className={`w-4 h-4 ${getStatusColor(host.enabled)}`} />
                            <span className={`text-sm ${getStatusColor(host.enabled)}`}>
                              {host.enabled ? 'Enabled' : 'Disabled'}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Globe className="w-4 h-4 text-muted-foreground" />
                            <div>
                              <div className="font-medium">{formatDomains(host.domain_names)}</div>
                              <div className="text-sm text-muted-foreground">
                                Primary: {host.primary_domain}
                              </div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <div className="font-medium">{host.target_url}</div>
                            <div className="text-muted-foreground">
                              {host.forward_host}:{host.forward_port}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          {host.ssl_enabled ? (
                            <Badge variant="secondary" className="text-green-600">
                              <Shield className="w-3 h-3 mr-1" />
                              SSL
                            </Badge>
                          ) : (
                            <Badge variant="outline">
                              No SSL
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          {host.has_access_list ? (
                            <Badge variant="secondary">
                              Protected
                            </Badge>
                          ) : (
                            <Badge variant="outline">
                              Public
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="sm">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem asChild>
                                <Link to={`/proxy-hosts/${host.id}`}>
                                  <Edit className="w-4 h-4 mr-2" />
                                  Edit
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleToggle(host.id)}
                                disabled={toggleMutation.isPending}
                              >
                                <Power className="w-4 h-4 mr-2" />
                                {host.enabled ? 'Disable' : 'Enable'}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() => handleDelete(host.id)}
                                disabled={deleteMutation.isPending}
                                className="text-red-600"
                              >
                                <Trash2 className="w-4 h-4 mr-2" />
                                Delete
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>

              {/* Pagination */}
              {data && data.pagination.pages > 1 && (
                <div className="flex items-center justify-between">
                  <p className="text-sm text-muted-foreground">
                    Showing {((data.pagination.page - 1) * data.pagination.limit) + 1} to{' '}
                    {Math.min(data.pagination.page * data.pagination.limit, data.pagination.total)} of{' '}
                    {data.pagination.total} results
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={!data.pagination.has_prev}
                      onClick={() => setFilters(prev => ({ ...prev, page: prev.page! - 1 }))}
                    >
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={!data.pagination.has_next}
                      onClick={() => setFilters(prev => ({ ...prev, page: prev.page! + 1 }))}
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
  );
}
