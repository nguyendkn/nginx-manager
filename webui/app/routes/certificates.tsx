import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Wifi, Shield, AlertTriangle, Calendar, MoreHorizontal, Upload, RotateCcw, Eye, Trash, Edit } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { Input } from '../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '../components/ui/dropdown-menu';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '../components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Alert, AlertDescription } from '../components/ui/alert';
import { certificatesApi, type Certificate, type CertificateRequest } from '../services/api/certificates';
import { toast } from 'sonner';

export default function CertificatesPage() {
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [selectedCertificate, setSelectedCertificate] = useState<Certificate | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showDetailsModal, setShowDetailsModal] = useState(false);

  const queryClient = useQueryClient();

  // Fetch certificates
  const { data: certificatesData, isLoading, error } = useQuery({
    queryKey: ['certificates', page, search],
    queryFn: () => certificatesApi.list({ page, per_page: 10 }),
  });

  // Fetch expiring certificates
  const { data: expiringCertificates } = useQuery({
    queryKey: ['certificates', 'expiring'],
    queryFn: () => certificatesApi.getExpiringSoon(30),
  });

  // Create certificate mutation
  const createMutation = useMutation({
    mutationFn: certificatesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificates'] });
      toast.success('Certificate created successfully');
      setShowCreateModal(false);
    },
    onError: (error: any) => {
      toast.error(`Failed to create certificate: ${error.message}`);
    },
  });

  // Delete certificate mutation
  const deleteMutation = useMutation({
    mutationFn: certificatesApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificates'] });
      toast.success('Certificate deleted successfully');
    },
    onError: (error: any) => {
      toast.error(`Failed to delete certificate: ${error.message}`);
    },
  });

  // Renew certificate mutation
  const renewMutation = useMutation({
    mutationFn: certificatesApi.renew,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificates'] });
      toast.success('Certificate renewed successfully');
    },
    onError: (error: any) => {
      toast.error(`Failed to renew certificate: ${error.message}`);
    },
  });

  // Filter certificates based on search
  const filteredCertificates = useMemo(() => {
    if (!certificatesData?.data) return [];

    if (!search) return certificatesData.data;

    return certificatesData.data.filter(cert =>
      cert.name.toLowerCase().includes(search.toLowerCase()) ||
      cert.domain_names.some(domain => domain.toLowerCase().includes(search.toLowerCase()))
    );
  }, [certificatesData?.data, search]);

  const getStatusBadge = (certificate: Certificate) => {
    const now = new Date();
    const expiryDate = certificate.expires_on ? new Date(certificate.expires_on) : null;

    if (!expiryDate) {
      return <Badge variant="secondary">No Expiry</Badge>;
    }

    const daysUntilExpiry = Math.ceil((expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

    if (daysUntilExpiry < 0) {
      return <Badge variant="destructive">Expired</Badge>;
    } else if (daysUntilExpiry <= 7) {
      return <Badge variant="destructive">Expires Soon</Badge>;
    } else if (daysUntilExpiry <= 30) {
      return <Badge variant="secondary">Expires in {daysUntilExpiry} days</Badge>;
    } else {
      return <Badge variant="default">Valid</Badge>;
    }
  };

  const getProviderIcon = (provider: string) => {
    switch (provider) {
      case 'letsencrypt':
        return <Shield className="h-4 w-4 text-green-500" />;
      case 'custom':
        return <Upload className="h-4 w-4 text-blue-500" />;
      default:
        return <Wifi className="h-4 w-4 text-gray-500" />;
    }
  };

  const handleDeleteCertificate = async (id: number) => {
    if (confirm('Are you sure you want to delete this certificate?')) {
      deleteMutation.mutate(id);
    }
  };

  const handleRenewCertificate = async (id: number) => {
    if (confirm('Are you sure you want to renew this certificate?')) {
      renewMutation.mutate(id);
    }
  };

  return (
    <div className="flex flex-col space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">SSL Certificates</h1>
          <p className="text-muted-foreground">
            Manage your SSL certificates and their renewal status
          </p>
        </div>
        <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              New Certificate
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[600px]">
            <DialogHeader>
              <DialogTitle>Create New Certificate</DialogTitle>
              <DialogDescription>
                Request a new SSL certificate or upload custom certificate files.
              </DialogDescription>
            </DialogHeader>
            <CreateCertificateForm
              onSubmit={(data) => createMutation.mutate(data)}
              isLoading={createMutation.isPending}
            />
          </DialogContent>
        </Dialog>
      </div>

      {/* Alert for expiring certificates */}
      {expiringCertificates && expiringCertificates.length > 0 && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            You have {expiringCertificates.length} certificate(s) expiring within 30 days.
            Consider renewing them soon.
          </AlertDescription>
        </Alert>
      )}

      {/* Search and Filters */}
      <div className="flex items-center space-x-2">
        <Input
          placeholder="Search certificates..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-sm"
        />
      </div>

      {/* Certificates Table */}
      <Card>
        <CardHeader>
          <CardTitle>Certificates</CardTitle>
          <CardDescription>
            {certificatesData?.total || 0} total certificates
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Domains</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead className="w-[100px]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center">
                    Loading certificates...
                  </TableCell>
                </TableRow>
              ) : filteredCertificates.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center">
                    No certificates found
                  </TableCell>
                </TableRow>
              ) : (
                filteredCertificates.map((certificate) => (
                  <TableRow key={certificate.id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center space-x-2">
                        {getProviderIcon(certificate.provider)}
                        <span>{certificate.name}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {certificate.domain_names.slice(0, 2).map((domain) => (
                          <Badge key={domain} variant="outline" className="text-xs">
                            {domain}
                          </Badge>
                        ))}
                        {certificate.domain_names.length > 2 && (
                          <Badge variant="outline" className="text-xs">
                            +{certificate.domain_names.length - 2} more
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={certificate.provider === 'letsencrypt' ? 'default' : 'secondary'}>
                        {certificate.provider === 'letsencrypt' ? "Let's Encrypt" : 'Custom'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {getStatusBadge(certificate)}
                    </TableCell>
                    <TableCell>
                      {certificate.expires_on ? (
                        <div className="flex items-center space-x-2">
                          <Calendar className="h-4 w-4 text-muted-foreground" />
                          <span className="text-sm">
                            {new Date(certificate.expires_on).toLocaleDateString()}
                          </span>
                        </div>
                      ) : (
                        <span className="text-sm text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" className="h-8 w-8 p-0">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => {
                            setSelectedCertificate(certificate);
                            setShowDetailsModal(true);
                          }}>
                            <Eye className="mr-2 h-4 w-4" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => {
                            setSelectedCertificate(certificate);
                            setShowUploadModal(true);
                          }}>
                            <Upload className="mr-2 h-4 w-4" />
                            Upload Files
                          </DropdownMenuItem>
                          {certificate.provider === 'letsencrypt' && (
                            <DropdownMenuItem onClick={() => handleRenewCertificate(certificate.id)}>
                              <RotateCcw className="mr-2 h-4 w-4" />
                              Renew
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuItem
                            onClick={() => handleDeleteCertificate(certificate.id)}
                            className="text-destructive"
                          >
                            <Trash className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          {certificatesData && certificatesData.total_pages > 1 && (
            <div className="flex items-center justify-between space-x-2 py-4">
              <div className="text-sm text-muted-foreground">
                Showing {((page - 1) * 10) + 1} to {Math.min(page * 10, certificatesData.total)} of {certificatesData.total} certificates
              </div>
              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page - 1)}
                  disabled={page <= 1}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page + 1)}
                  disabled={page >= certificatesData.total_pages}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Certificate Details Modal */}
      <Dialog open={showDetailsModal} onOpenChange={setShowDetailsModal}>
        <DialogContent className="sm:max-w-[700px]">
          <DialogHeader>
            <DialogTitle>Certificate Details</DialogTitle>
          </DialogHeader>
          {selectedCertificate && (
            <CertificateDetailsView certificate={selectedCertificate} />
          )}
        </DialogContent>
      </Dialog>

      {/* Upload Certificate Modal */}
      <Dialog open={showUploadModal} onOpenChange={setShowUploadModal}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>Upload Certificate Files</DialogTitle>
            <DialogDescription>
              Upload certificate and private key files for {selectedCertificate?.name}
            </DialogDescription>
          </DialogHeader>
          {selectedCertificate && (
            <UploadCertificateForm
              certificateId={selectedCertificate.id}
              onSuccess={() => {
                queryClient.invalidateQueries({ queryKey: ['certificates'] });
                setShowUploadModal(false);
                toast.success('Certificate files uploaded successfully');
              }}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Create Certificate Form Component
function CreateCertificateForm({
  onSubmit,
  isLoading
}: {
  onSubmit: (data: CertificateRequest) => void;
  isLoading: boolean;
}) {
  const [formData, setFormData] = useState<CertificateRequest>({
    name: '',
    nice_name: '',
    provider: 'letsencrypt',
    domain_names: [''],
  });

  const [activeTab, setActiveTab] = useState('letsencrypt');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    // Filter out empty domain names
    const cleanedData = {
      ...formData,
      domain_names: formData.domain_names.filter(domain => domain.trim() !== ''),
    };

    onSubmit(cleanedData);
  };

  const addDomain = () => {
    setFormData(prev => ({
      ...prev,
      domain_names: [...prev.domain_names, '']
    }));
  };

  const removeDomain = (index: number) => {
    setFormData(prev => ({
      ...prev,
      domain_names: prev.domain_names.filter((_, i) => i !== index)
    }));
  };

  const updateDomain = (index: number, value: string) => {
    setFormData(prev => ({
      ...prev,
      domain_names: prev.domain_names.map((domain, i) => i === index ? value : domain)
    }));
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="letsencrypt">Let's Encrypt</TabsTrigger>
          <TabsTrigger value="custom">Custom Certificate</TabsTrigger>
        </TabsList>

        <TabsContent value="letsencrypt" className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Certificate Name</label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value, provider: 'letsencrypt' }))}
              placeholder="my-certificate"
              required
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Nice Name (Optional)</label>
            <Input
              value={formData.nice_name}
              onChange={(e) => setFormData(prev => ({ ...prev, nice_name: e.target.value }))}
              placeholder="My Website Certificate"
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Domain Names</label>
            {formData.domain_names.map((domain, index) => (
              <div key={index} className="flex space-x-2">
                <Input
                  value={domain}
                  onChange={(e) => updateDomain(index, e.target.value)}
                  placeholder="example.com"
                  required={index === 0}
                />
                {index > 0 && (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => removeDomain(index)}
                  >
                    Remove
                  </Button>
                )}
              </div>
            ))}
            <Button type="button" variant="outline" onClick={addDomain}>
              Add Domain
            </Button>
          </div>
        </TabsContent>

        <TabsContent value="custom" className="space-y-4">
          {/* Custom certificate form fields would go here */}
          <Alert>
            <AlertDescription>
              Custom certificate upload will be available after creating the certificate entry.
            </AlertDescription>
          </Alert>
        </TabsContent>
      </Tabs>

      <div className="flex justify-end space-x-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Creating...' : 'Create Certificate'}
        </Button>
      </div>
    </form>
  );
}

// Certificate Details View Component
function CertificateDetailsView({ certificate }: { certificate: Certificate }) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="text-sm font-medium text-muted-foreground">Name</label>
          <p className="text-sm">{certificate.name}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-muted-foreground">Provider</label>
          <p className="text-sm">{certificate.provider === 'letsencrypt' ? "Let's Encrypt" : 'Custom'}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-muted-foreground">Status</label>
          <p className="text-sm">{certificate.status}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-muted-foreground">Expires On</label>
          <p className="text-sm">
            {certificate.expires_on ? new Date(certificate.expires_on).toLocaleDateString() : 'No expiry date'}
          </p>
        </div>
      </div>

      <div>
        <label className="text-sm font-medium text-muted-foreground">Domain Names</label>
        <div className="flex flex-wrap gap-2 mt-1">
          {certificate.domain_names.map((domain) => (
            <Badge key={domain} variant="outline">
              {domain}
            </Badge>
          ))}
        </div>
      </div>

      <div>
        <label className="text-sm font-medium text-muted-foreground">Created</label>
        <p className="text-sm">{new Date(certificate.created_at).toLocaleString()}</p>
      </div>

      <div>
        <label className="text-sm font-medium text-muted-foreground">Last Updated</label>
        <p className="text-sm">{new Date(certificate.updated_at).toLocaleString()}</p>
      </div>
    </div>
  );
}

// Upload Certificate Form Component
function UploadCertificateForm({
  certificateId,
  onSuccess
}: {
  certificateId: number;
  onSuccess: () => void;
}) {
  const [formData, setFormData] = useState({
    certificate: '',
    certificate_key: '',
    intermediate_certificate: '',
  });

  const uploadMutation = useMutation({
    mutationFn: (data: any) => certificatesApi.upload(certificateId, data),
    onSuccess: onSuccess,
    onError: (error: any) => {
      toast.error(`Failed to upload certificate: ${error.message}`);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    uploadMutation.mutate(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm font-medium">Certificate</label>
        <textarea
          className="w-full min-h-32 p-2 border rounded"
          value={formData.certificate}
          onChange={(e) => setFormData(prev => ({ ...prev, certificate: e.target.value }))}
          placeholder="-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Private Key</label>
        <textarea
          className="w-full min-h-32 p-2 border rounded"
          value={formData.certificate_key}
          onChange={(e) => setFormData(prev => ({ ...prev, certificate_key: e.target.value }))}
          placeholder="-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Intermediate Certificate (Optional)</label>
        <textarea
          className="w-full min-h-24 p-2 border rounded"
          value={formData.intermediate_certificate}
          onChange={(e) => setFormData(prev => ({ ...prev, intermediate_certificate: e.target.value }))}
          placeholder="-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----"
        />
      </div>

      <div className="flex justify-end space-x-2">
        <Button type="submit" disabled={uploadMutation.isPending}>
          {uploadMutation.isPending ? 'Uploading...' : 'Upload Certificate'}
        </Button>
      </div>
    </form>
  );
}
