import { apiClient, type ApiResponse } from './client';

// Monitoring types
export interface SystemMetrics {
  timestamp: string;
  cpu: CPUStats;
  memory: MemStats;
  disk: DiskStats;
  network: NetStats;
  process: ProcStats;
}

export interface CPUStats {
  usage: number;
  load_avg_1: number;
  load_avg_5: number;
  load_avg_15: number;
}

export interface MemStats {
  total: number;
  available: number;
  used: number;
  used_percent: number;
  go_alloc: number;
  go_total: number;
  go_sys: number;
}

export interface DiskStats {
  total: number;
  free: number;
  used: number;
  used_percent: number;
}

export interface NetStats {
  bytes_recv: number;
  bytes_sent: number;
  packets_recv: number;
  packets_sent: number;
}

export interface ProcStats {
  goroutines: number;
  gc_runs: number;
  uptime: number;
  go_version: string;
  pid: number;
}

export interface NginxStatus {
  running: boolean;
  pid: number;
  version: string;
  config_test: boolean;
  last_reload: string;
  connections: number;
}

export interface ActivityEvent {
  id: string;
  timestamp: string;
  type: string;
  message: string;
  level: string;
  details: Record<string, any>;
}

export interface ActivityFeedResponse {
  activities: ActivityEvent[];
  total: number;
  limit: number;
  timestamp: string;
}

export interface DashboardStats {
  system_metrics: SystemMetrics;
  nginx_status: NginxStatus;
  recent_activity: {
    activities: ActivityEvent[];
    count: number;
  };
  summary: {
    uptime: string;
    memory_usage: string;
    memory_percent: string;
    disk_usage: string;
    disk_percent: string;
    cpu_usage: string;
    goroutines: number;
    nginx_running: boolean;
    nginx_config_ok: boolean;
  };
  timestamp: string;
}

export interface NginxControlRequest {
  action: 'start' | 'stop' | 'restart' | 'reload' | 'test';
}

export interface NginxControlResponse {
  action: string;
  success: boolean;
  message: string;
  timestamp: string;
}

export interface WebSocketMessage {
  type: 'metrics' | 'nginx_status' | 'activity' | 'error';
  timestamp: string;
  data: any;
}

const ENDPOINTS = {
  DASHBOARD: '/api/v1/monitoring/dashboard',
  SYSTEM_METRICS: '/api/v1/monitoring/system-metrics',
  NGINX_STATUS: '/api/v1/monitoring/nginx-status',
  ACTIVITY_FEED: '/api/v1/monitoring/activity-feed',
  WEBSOCKET: '/api/v1/monitoring/ws',
  NGINX_CONTROL: '/api/v1/monitoring/nginx/control',
} as const;

export const monitoringApi = {
  // Get dashboard statistics
  getDashboardStats: async (): Promise<DashboardStats> => {
    const response = await apiClient.get<ApiResponse<DashboardStats>>(
      ENDPOINTS.DASHBOARD
    );
    return response.data.data as DashboardStats;
  },

  // Get system metrics
  getSystemMetrics: async (): Promise<SystemMetrics> => {
    const response = await apiClient.get<ApiResponse<SystemMetrics>>(
      ENDPOINTS.SYSTEM_METRICS
    );
    return response.data.data as SystemMetrics;
  },

  // Get nginx status
  getNginxStatus: async (): Promise<NginxStatus> => {
    const response = await apiClient.get<ApiResponse<NginxStatus>>(
      ENDPOINTS.NGINX_STATUS
    );
    return response.data.data as NginxStatus;
  },

  // Get activity feed
  getActivityFeed: async (limit?: number): Promise<ActivityFeedResponse> => {
    const response = await apiClient.get<ApiResponse<ActivityFeedResponse>>(
      ENDPOINTS.ACTIVITY_FEED,
      { params: limit ? { limit } : undefined }
    );
    return response.data.data as ActivityFeedResponse;
  },

  // Control nginx service
  controlNginx: async (data: NginxControlRequest): Promise<NginxControlResponse> => {
    const response = await apiClient.post<ApiResponse<NginxControlResponse>>(
      ENDPOINTS.NGINX_CONTROL,
      data
    );
    return response.data.data as NginxControlResponse;
  },

  // WebSocket connection for real-time updates
  createWebSocket: (
    onMessage: (message: WebSocketMessage) => void,
    onError?: (error: Event) => void,
    onClose?: (event: CloseEvent) => void
  ): WebSocket => {
    // Get the base URL and convert to WebSocket protocol
    const baseUrl = apiClient.defaults.baseURL || window.location.origin;
    const wsUrl = baseUrl.replace(/^http/, 'ws') + ENDPOINTS.WEBSOCKET;

    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        onMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
        onError?.(event);
      }
    };

    ws.onerror = (event) => {
      console.error('WebSocket error:', event);
      onError?.(event);
    };

    ws.onclose = (event) => {
      console.log('WebSocket connection closed:', event);
      onClose?.(event);
    };

    return ws;
  },
};

// Helper functions for formatting
export const formatBytes = (bytes: number): string => {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

export const formatPercentage = (percent: number): string => {
  return `${percent.toFixed(1)}%`;
};

export const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / (24 * 3600));
  const hours = Math.floor((seconds % (24 * 3600)) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
};
