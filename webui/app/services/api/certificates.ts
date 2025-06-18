import { apiClient, type ApiResponse } from './client';

// Certificate types
export interface Certificate {
  id: number;
  name: string;
  nice_name: string;
  provider: 'letsencrypt' | 'custom';
  domain_names: string[];
  expires_on: string | null;
  status: string;
  has_validation: boolean;
  certificate?: string;
  certificate_key?: string;
  intermediate_certificate?: string;
  meta?: Record<string, any>;
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface CertificateRequest {
  name: string;
  nice_name?: string;
  provider: 'letsencrypt' | 'custom';
  domain_names: string[];
  certificate?: string;
  certificate_key?: string;
  intermediate_certificate?: string;
  meta?: Record<string, any>;
}

export interface CertificateListResponse {
  data: Certificate[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface UploadCertificateRequest {
  certificate: string;
  certificate_key: string;
  intermediate_certificate?: string;
}

export interface TestCertificateRequest {
  domains: string[];
}

export interface DomainTestResult {
  domain: string;
  reachable: boolean;
  ssl: boolean;
  port_80: boolean;
  port_443: boolean;
  message: string;
  response_time_ms: number;
}

export interface TestCertificateResponse {
  success: boolean;
  results: DomainTestResult[];
  errors: string[];
}

export interface RenewCertificateResponse {
  success: boolean;
  certificate: Certificate;
  message: string;
}

const ENDPOINTS = {
  CERTIFICATES: '/api/v1/certificates',
  CERTIFICATE_BY_ID: (id: number) => `/api/v1/certificates/${id}`,
  UPLOAD_CERTIFICATE: (id: number) => `/api/v1/certificates/${id}/upload`,
  RENEW_CERTIFICATE: (id: number) => `/api/v1/certificates/${id}/renew`,
  TEST_CERTIFICATE: '/api/v1/certificates/test',
  EXPIRING_SOON: '/api/v1/certificates/expiring-soon',
} as const;

export const certificatesApi = {
  // List certificates with pagination
  list: async (params?: {
    page?: number;
    per_page?: number;
  }): Promise<CertificateListResponse> => {
    const response = await apiClient.get<ApiResponse<CertificateListResponse>>(
      ENDPOINTS.CERTIFICATES,
      { params }
    );
    return response.data.data as CertificateListResponse;
  },

  // Get a single certificate
  get: async (id: number): Promise<Certificate> => {
    const response = await apiClient.get<ApiResponse<Certificate>>(
      ENDPOINTS.CERTIFICATE_BY_ID(id)
    );
    return response.data.data as Certificate;
  },

  // Create a new certificate
  create: async (data: CertificateRequest): Promise<Certificate> => {
    const response = await apiClient.post<ApiResponse<Certificate>>(
      ENDPOINTS.CERTIFICATES,
      data
    );
    return response.data.data as Certificate;
  },

  // Update a certificate
  update: async (id: number, data: CertificateRequest): Promise<Certificate> => {
    const response = await apiClient.put<ApiResponse<Certificate>>(
      ENDPOINTS.CERTIFICATE_BY_ID(id),
      data
    );
    return response.data.data as Certificate;
  },

  // Delete a certificate
  delete: async (id: number): Promise<{ id: number }> => {
    const response = await apiClient.delete<ApiResponse<{ id: number }>>(
      ENDPOINTS.CERTIFICATE_BY_ID(id)
    );
    return response.data.data as { id: number };
  },

  // Upload certificate files
  upload: async (id: number, data: UploadCertificateRequest): Promise<Certificate> => {
    const response = await apiClient.post<ApiResponse<Certificate>>(
      ENDPOINTS.UPLOAD_CERTIFICATE(id),
      data
    );
    return response.data.data as Certificate;
  },

  // Renew a certificate
  renew: async (id: number): Promise<RenewCertificateResponse> => {
    const response = await apiClient.post<ApiResponse<RenewCertificateResponse>>(
      ENDPOINTS.RENEW_CERTIFICATE(id)
    );
    return response.data.data as RenewCertificateResponse;
  },

  // Test domains for certificate validation
  test: async (data: TestCertificateRequest): Promise<TestCertificateResponse> => {
    const response = await apiClient.post<ApiResponse<TestCertificateResponse>>(
      ENDPOINTS.TEST_CERTIFICATE,
      data
    );
    return response.data.data as TestCertificateResponse;
  },

  // Get certificates expiring soon
  getExpiringSoon: async (days: number = 30): Promise<Certificate[]> => {
    const response = await apiClient.get<ApiResponse<Certificate[]>>(
      ENDPOINTS.EXPIRING_SOON,
      { params: { days } }
    );
    return response.data.data as Certificate[];
  },
};
