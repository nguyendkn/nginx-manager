import { api } from './client';

export interface MetricQuery {
  metric_type: string;
  metric_name: string;
  time_range: {
    start: string;
    end: string;
  };
  aggregation?: string;
  group_by?: string;
  limit?: number;
  tags?: Record<string, any>;
}

export interface DataPoint {
  timestamp: string;
  value: number;
  tags?: Record<string, any>;
}

export interface AlertRule {
  id?: number;
  name: string;
  description?: string;
  metric_type: string;
  metric_name: string;
  condition: 'gt' | 'lt' | 'eq' | 'ne' | 'between';
  threshold: number;
  threshold_max?: number;
  severity: 'info' | 'warning' | 'critical';
  is_enabled: boolean;
  evaluation_window: number;
  notification_channels?: NotificationChannel[];
  tags?: Record<string, any>;
  last_triggered?: string;
  user_id?: number;
}

export interface AlertInstance {
  id: number;
  alert_rule_id: number;
  alert_rule?: AlertRule;
  triggered_at: string;
  resolved_at?: string;
  status: 'triggered' | 'resolved' | 'suppressed';
  current_value: number;
  threshold_value: number;
  message: string;
  context?: Record<string, any>;
  notifications_sent: number;
}

export interface NotificationChannel {
  id?: number;
  name: string;
  type: 'email' | 'slack' | 'webhook' | 'teams';
  is_enabled: boolean;
  configuration: Record<string, any>;
  user_id?: number;
}

export interface Dashboard {
  id?: number;
  name: string;
  description?: string;
  is_default: boolean;
  is_public: boolean;
  layout?: Record<string, any>;
  widgets?: DashboardWidget[];
  user_id?: number;
  shared_with?: any[];
  created_at?: string;
  updated_at?: string;
}

export interface DashboardWidget {
  id?: number;
  dashboard_id?: number;
  type: 'chart' | 'metric' | 'table' | 'gauge';
  title: string;
  position: {
    x: number;
    y: number;
    width: number;
    height: number;
  };
  configuration: Record<string, any>;
  data_source: string;
  query: string;
  refresh_interval: number;
  is_visible: boolean;
}

export interface SystemMetricsSummary {
  time_range: {
    start: string;
    end: string;
  };
  metrics: {
    cpu: MetricSummary;
    memory: MetricSummary;
    disk: MetricSummary;
  };
  timestamp: string;
}

export interface MetricSummary {
  current: number;
  average: number;
  peak: number;
  minimum: number;
  trend: 'increasing' | 'decreasing' | 'stable' | 'unknown';
  data_points: number;
}

class AnalyticsAPI {
  // Metrics endpoints
  async queryMetrics(query: MetricQuery): Promise<{ data_points: DataPoint[]; count: number; query: MetricQuery; timestamp: string }> {
    const response = await api.post('/analytics/metrics/query', query);
    return response.data.data as { data_points: DataPoint[]; count: number; query: MetricQuery; timestamp: string };
  }

  async getHistoricalMetrics(
    metricType: string,
    metricName: string,
    params?: {
      start?: string;
      end?: string;
      aggregation?: string;
      group_by?: string;
      limit?: number;
    }
  ): Promise<{
    metric_type: string;
    metric_name: string;
    data_points: DataPoint[];
    count: number;
    time_range: { start: string; end: string };
    aggregation: string;
    timestamp: string;
  }> {
    const searchParams = new URLSearchParams();
    if (params?.start) searchParams.append('start', params.start);
    if (params?.end) searchParams.append('end', params.end);
    if (params?.aggregation) searchParams.append('aggregation', params.aggregation);
    if (params?.group_by) searchParams.append('group_by', params.group_by);
    if (params?.limit) searchParams.append('limit', params.limit.toString());

    const url = `/analytics/metrics/${metricType}/${metricName}${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
    const response = await api.get(url);
    return response.data.data as {
      metric_type: string;
      metric_name: string;
      data_points: DataPoint[];
      count: number;
      time_range: { start: string; end: string };
      aggregation: string;
      timestamp: string;
    };
  }

  async getSystemMetricsSummary(range: '1h' | '24h' | '7d' | '30d' = '24h'): Promise<SystemMetricsSummary> {
    const response = await api.get(`/analytics/system/summary?range=${range}`);
    return response.data.data as SystemMetricsSummary;
  }

  // Alert Rules endpoints
  async createAlertRule(alertRule: Omit<AlertRule, 'id' | 'user_id'>): Promise<AlertRule> {
    const response = await api.post('/analytics/alerts/rules', alertRule);
    return response.data.data as AlertRule;
  }

  async getAlertRules(): Promise<{ alert_rules: AlertRule[]; count: number; timestamp: string }> {
    const response = await api.get('/analytics/alerts/rules');
    return response.data.data as { alert_rules: AlertRule[]; count: number; timestamp: string };
  }

  async updateAlertRule(id: number, alertRule: Partial<AlertRule>): Promise<AlertRule> {
    const response = await api.put(`/analytics/alerts/rules/${id}`, alertRule);
    return response.data.data as AlertRule;
  }

  async deleteAlertRule(id: number): Promise<{ id: number }> {
    const response = await api.delete(`/analytics/alerts/rules/${id}`);
    return response.data.data as { id: number };
  }

  // Alert Instances endpoints
  async getAlertInstances(params?: {
    status?: string;
    severity?: string;
    limit?: number;
    offset?: number;
  }): Promise<{
    alert_instances: AlertInstance[];
    total: number;
    limit: number;
    offset: number;
    timestamp: string;
  }> {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.append('status', params.status);
    if (params?.severity) searchParams.append('severity', params.severity);
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    if (params?.offset) searchParams.append('offset', params.offset.toString());

    const url = `/analytics/alerts/instances${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
    const response = await api.get(url);
    return response.data.data as {
      alert_instances: AlertInstance[];
      total: number;
      limit: number;
      offset: number;
      timestamp: string;
    };
  }

  // Dashboard endpoints
  async createDashboard(dashboard: Omit<Dashboard, 'id' | 'user_id' | 'created_at' | 'updated_at'>): Promise<Dashboard> {
    const response = await api.post('/analytics/dashboards', dashboard);
    return response.data.data as Dashboard;
  }

  async getDashboards(): Promise<{ dashboards: Dashboard[]; count: number; timestamp: string }> {
    const response = await api.get('/analytics/dashboards');
    return response.data.data as { dashboards: Dashboard[]; count: number; timestamp: string };
  }

  async getDashboard(id: number): Promise<Dashboard> {
    const response = await api.get(`/analytics/dashboards/${id}`);
    return response.data.data as Dashboard;
  }

  async updateDashboard(id: number, dashboard: Partial<Dashboard>): Promise<Dashboard> {
    const response = await api.put(`/analytics/dashboards/${id}`, dashboard);
    return response.data.data as Dashboard;
  }

  async deleteDashboard(id: number): Promise<{ id: number }> {
    const response = await api.delete(`/analytics/dashboards/${id}`);
    return response.data.data as { id: number };
  }

  // Utility methods for common queries
  async getCPUMetrics(timeRange: { start: string; end: string }): Promise<DataPoint[]> {
    const result = await this.queryMetrics({
      metric_type: 'system',
      metric_name: 'cpu_usage',
      time_range: timeRange,
      aggregation: 'avg',
      group_by: '5m',
      limit: 500,
    });
    return result.data_points;
  }

  async getMemoryMetrics(timeRange: { start: string; end: string }): Promise<DataPoint[]> {
    const result = await this.queryMetrics({
      metric_type: 'system',
      metric_name: 'memory_usage',
      time_range: timeRange,
      aggregation: 'avg',
      group_by: '5m',
      limit: 500,
    });
    return result.data_points;
  }

  async getDiskMetrics(timeRange: { start: string; end: string }): Promise<DataPoint[]> {
    const result = await this.queryMetrics({
      metric_type: 'system',
      metric_name: 'disk_usage',
      time_range: timeRange,
      aggregation: 'avg',
      group_by: '5m',
      limit: 500,
    });
    return result.data_points;
  }

  async getNetworkMetrics(timeRange: { start: string; end: string }): Promise<DataPoint[]> {
    const result = await this.queryMetrics({
      metric_type: 'system',
      metric_name: 'network_bandwidth',
      time_range: timeRange,
      aggregation: 'avg',
      group_by: '5m',
      limit: 500,
    });
    return result.data_points;
  }

  async getNginxMetrics(timeRange: { start: string; end: string }): Promise<DataPoint[]> {
    const result = await this.queryMetrics({
      metric_type: 'nginx',
      metric_name: 'requests_per_second',
      time_range: timeRange,
      aggregation: 'avg',
      group_by: '5m',
      limit: 500,
    });
    return result.data_points;
  }
}

export const analyticsAPI = new AnalyticsAPI();
