import { api } from './client';

export interface DashboardStats {
  systemStats: {
    uptime: string;
    memoryUsage: number;
    diskUsage: number;
    cpuUsage: number;
    activeConnections: number;
  };
  counts: {
    proxyHosts: number;
    certificates: number;
    accessLists: number;
    activeHosts: number;
  };
  recentActivity: Array<{
    id: string;
    action: string;
    resource: string;
    timestamp: string;
    status: 'success' | 'warning' | 'error';
  }>;
  health: {
    nginx: boolean;
    database: boolean;
    storage: boolean;
  };
}

export interface SystemMetrics {
  cpu: {
    usage: number;
    loadAvg1: number;
    loadAvg5: number;
    loadAvg15: number;
  };
  memory: {
    total: number;
    used: number;
    free: number;
    usedPercent: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
    usedPercent: number;
  };
  network: {
    bytesRecv: number;
    bytesSent: number;
    packetsRecv: number;
    packetsSent: number;
  };
  process: {
    pid: number;
    uptime: string;
    goroutines: number;
    version: string;
  };
}

export interface ActivityEvent {
  id: string;
  action: string;
  resource: string;
  resourceId?: string;
  userId?: string;
  userName?: string;
  timestamp: string;
  status: 'success' | 'warning' | 'error';
  details?: string;
}

const dashboardApi = {
  /**
   * Get comprehensive dashboard statistics
   */
  async getDashboardStats(): Promise<DashboardStats> {
    try {
      // For now, return mock data since backend might not be fully available
      // In production, this would call the actual API endpoints
      return {
        systemStats: {
          uptime: '3 days, 2 hours',
          memoryUsage: 67,
          diskUsage: 45,
          cpuUsage: 23,
          activeConnections: 142
        },
        counts: {
          proxyHosts: 12,
          certificates: 8,
          accessLists: 3,
          activeHosts: 10
        },
        recentActivity: [
          {
            id: '1',
            action: 'Created proxy host',
            resource: 'api.example.com',
            timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
            status: 'success'
          },
          {
            id: '2',
            action: 'Certificate renewed',
            resource: 'blog.example.com',
            timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
            status: 'success'
          },
          {
            id: '3',
            action: 'Access list updated',
            resource: 'Admin access list',
            timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
            status: 'warning'
          },
          {
            id: '4',
            action: 'Proxy host disabled',
            resource: 'old.example.com',
            timestamp: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
            status: 'error'
          }
        ],
        health: {
          nginx: true,
          database: true,
          storage: true
        }
      };
    } catch (error) {
      console.error('Dashboard API error:', error);
      // Return fallback data structure
      return {
        systemStats: {
          uptime: 'Unknown',
          memoryUsage: 0,
          diskUsage: 0,
          cpuUsage: 0,
          activeConnections: 0
        },
        counts: {
          proxyHosts: 0,
          certificates: 0,
          accessLists: 0,
          activeHosts: 0
        },
        recentActivity: [],
        health: {
          nginx: false,
          database: false,
          storage: false
        }
      };
    }
  },

  /**
   * Get detailed system metrics
   */
  async getSystemMetrics(): Promise<SystemMetrics> {
    try {
      const response = await api.get('/monitoring/system-metrics');
      return response.data.data as SystemMetrics;
    } catch (error) {
      console.error('System metrics API error:', error);
      throw error;
    }
  },

  /**
   * Get recent activity events
   */
  async getRecentActivity(limit: number = 10): Promise<ActivityEvent[]> {
    try {
      const response = await api.get(`/monitoring/activity-feed?limit=${limit}`);
      return response.data.data as ActivityEvent[];
    } catch (error) {
      console.error('Activity feed API error:', error);
      throw error;
    }
  },

    /**
   * Get system health status
   */
  async getHealthStatus(): Promise<{ nginx: boolean; database: boolean; storage: boolean }> {
    try {
      const response = await api.get<{
        status: string;
        database?: { connected: boolean };
        storage?: { available: boolean };
      }>('/health');
      const data = response.data.data;

      return {
        nginx: data?.status === 'healthy',
        database: data?.database?.connected === true,
        storage: data?.storage?.available === true
      };
    } catch (error) {
      console.error('Health status API error:', error);
      return {
        nginx: false,
        database: false,
        storage: false
      };
    }
  },

  /**
   * Get resource counts (proxy hosts, certificates, etc.)
   */
  async getResourceCounts(): Promise<{
    proxyHosts: number;
    certificates: number;
    accessLists: number;
    activeHosts: number;
  }> {
    try {
      // In a real implementation, these would be separate API calls
      // For now, return mock data
      return {
        proxyHosts: 12,
        certificates: 8,
        accessLists: 3,
        activeHosts: 10
      };
    } catch (error) {
      console.error('Resource counts API error:', error);
      return {
        proxyHosts: 0,
        certificates: 0,
        accessLists: 0,
        activeHosts: 0
      };
    }
  }
};

export { dashboardApi };
