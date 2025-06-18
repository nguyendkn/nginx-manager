import React, { useState, useEffect } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Badge } from '../components/ui/badge';
import { Alert, AlertDescription, AlertTitle } from '../components/ui/alert';
import { Separator } from '../components/ui/separator';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import {
  Activity,
  AlertTriangle,
  BarChart3,
  Bell,
  Clock,
  Database,
  Gauge,
  Info,
  LineChart,
  Plus,
  RefreshCw,
  Settings,
  TrendingDown,
  TrendingUp,
} from 'lucide-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { analyticsAPI, type SystemMetricsSummary, type AlertRule, type AlertInstance, type Dashboard } from '../services/api/analytics';
import { format, subHours, subDays, subWeeks, subMonths } from 'date-fns';

interface TimeRangeOption {
  label: string;
  value: '1h' | '24h' | '7d' | '30d';
  hours: number;
}

const timeRangeOptions: TimeRangeOption[] = [
  { label: 'Last Hour', value: '1h', hours: 1 },
  { label: 'Last 24 Hours', value: '24h', hours: 24 },
  { label: 'Last 7 Days', value: '7d', hours: 24 * 7 },
  { label: 'Last 30 Days', value: '30d', hours: 24 * 30 },
];

export default function AnalyticsPage() {
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d' | '30d'>('24h');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const queryClient = useQueryClient();

  // Auto-refresh every 30 seconds
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      queryClient.invalidateQueries({ queryKey: ['analytics'] });
    }, 30000);

    return () => clearInterval(interval);
  }, [autoRefresh, queryClient]);

  // System metrics summary
  const { data: systemSummary, isLoading: summaryLoading, error: summaryError } = useQuery({
    queryKey: ['analytics', 'system-summary', timeRange],
    queryFn: () => analyticsAPI.getSystemMetricsSummary(timeRange),
    refetchInterval: autoRefresh ? 30000 : false,
    staleTime: 10000,
  });

  // Alert rules
  const { data: alertRulesData, isLoading: alertRulesLoading } = useQuery({
    queryKey: ['analytics', 'alert-rules'],
    queryFn: () => analyticsAPI.getAlertRules(),
    refetchInterval: autoRefresh ? 60000 : false,
  });

  // Alert instances
  const { data: alertInstancesData, isLoading: alertInstancesLoading } = useQuery({
    queryKey: ['analytics', 'alert-instances'],
    queryFn: () => analyticsAPI.getAlertInstances({ limit: 10 }),
    refetchInterval: autoRefresh ? 30000 : false,
  });

  // Dashboards
  const { data: dashboardsData, isLoading: dashboardsLoading } = useQuery({
    queryKey: ['analytics', 'dashboards'],
    queryFn: () => analyticsAPI.getDashboards(),
    refetchInterval: autoRefresh ? 300000 : false, // Refresh every 5 minutes
  });

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['analytics'] });
    toast.success('Analytics data refreshed');
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'destructive';
      case 'warning':
        return 'default';
      case 'info':
        return 'secondary';
      default:
        return 'secondary';
    }
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'increasing':
        return <TrendingUp className="h-4 w-4 text-red-500" />;
      case 'decreasing':
        return <TrendingDown className="h-4 w-4 text-green-500" />;
      default:
        return <Activity className="h-4 w-4 text-blue-500" />;
    }
  };

  if (summaryError) {
    return (
      <div className="container mx-auto p-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error Loading Analytics</AlertTitle>
          <AlertDescription>
            Failed to load analytics data. Please check your connection and try again.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Analytics Dashboard</h1>
          <p className="text-muted-foreground">
            Comprehensive monitoring and performance insights for your Nginx infrastructure
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={timeRange} onValueChange={(value: any) => setTimeRange(value)}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {timeRangeOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={autoRefresh ? 'bg-green-50 text-green-700 border-green-200' : ''}
          >
            <Clock className="h-4 w-4 mr-2" />
            Auto Refresh
          </Button>
          <Button variant="outline" size="sm" onClick={handleRefresh}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview" className="space-y-6">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="metrics">Historical Metrics</TabsTrigger>
          <TabsTrigger value="alerts">Alerts</TabsTrigger>
          <TabsTrigger value="dashboards">Dashboards</TabsTrigger>
          <TabsTrigger value="insights">Insights</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          {/* System Overview Cards */}
          {summaryLoading ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              {[...Array(4)].map((_, i) => (
                <Card key={i}>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Loading...</CardTitle>
                    <div className="h-4 w-4 bg-gray-200 rounded animate-pulse" />
                  </CardHeader>
                  <CardContent>
                    <div className="h-8 bg-gray-200 rounded animate-pulse mb-2" />
                    <div className="h-4 bg-gray-200 rounded animate-pulse w-1/2" />
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : systemSummary ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              {/* CPU Usage */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">CPU Usage</CardTitle>
                  <Gauge className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {systemSummary.metrics.cpu.current.toFixed(1)}%
                  </div>
                  <div className="flex items-center text-xs text-muted-foreground">
                    {getTrendIcon(systemSummary.metrics.cpu.trend)}
                    <span className="ml-1">
                      Avg: {systemSummary.metrics.cpu.average.toFixed(1)}%
                    </span>
                  </div>
                </CardContent>
              </Card>

              {/* Memory Usage */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
                  <Database className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {systemSummary.metrics.memory.current.toFixed(1)}%
                  </div>
                  <div className="flex items-center text-xs text-muted-foreground">
                    {getTrendIcon(systemSummary.metrics.memory.trend)}
                    <span className="ml-1">
                      Avg: {systemSummary.metrics.memory.average.toFixed(1)}%
                    </span>
                  </div>
                </CardContent>
              </Card>

              {/* Disk Usage */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Disk Usage</CardTitle>
                  <BarChart3 className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {systemSummary.metrics.disk.current.toFixed(1)}%
                  </div>
                  <div className="flex items-center text-xs text-muted-foreground">
                    {getTrendIcon(systemSummary.metrics.disk.trend)}
                    <span className="ml-1">
                      Avg: {systemSummary.metrics.disk.average.toFixed(1)}%
                    </span>
                  </div>
                </CardContent>
              </Card>

              {/* Active Alerts */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Active Alerts</CardTitle>
                  <Bell className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {alertInstancesData?.alert_instances.filter(a => a.status === 'triggered').length || 0}
                  </div>
                  <div className="flex items-center text-xs text-muted-foreground">
                    <Info className="h-3 w-3 mr-1" />
                    <span>
                      {alertRulesData?.alert_rules.filter(r => r.is_enabled).length || 0} rules enabled
                    </span>
                  </div>
                </CardContent>
              </Card>
            </div>
          ) : null}

          {/* Recent Alerts */}
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Bell className="h-5 w-5" />
                  Recent Alerts
                </CardTitle>
                <CardDescription>
                  Latest alert instances and their status
                </CardDescription>
              </CardHeader>
              <CardContent>
                {alertInstancesLoading ? (
                  <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                      <div key={i} className="h-16 bg-gray-200 rounded animate-pulse" />
                    ))}
                  </div>
                ) : alertInstancesData?.alert_instances.length ? (
                  <div className="space-y-3">
                    {alertInstancesData.alert_instances.slice(0, 5).map((alert) => (
                      <div key={alert.id} className="flex items-center justify-between p-3 border rounded-lg">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <Badge variant={getSeverityColor(alert.alert_rule?.severity || 'info')}>
                              {alert.alert_rule?.severity}
                            </Badge>
                            <span className="font-medium">{alert.alert_rule?.name}</span>
                          </div>
                          <p className="text-sm text-muted-foreground">{alert.message}</p>
                        </div>
                        <div className="text-right">
                          <Badge variant={alert.status === 'triggered' ? 'destructive' : 'secondary'}>
                            {alert.status}
                          </Badge>
                          <p className="text-xs text-muted-foreground mt-1">
                            {format(new Date(alert.triggered_at), 'MMM dd, HH:mm')}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <Bell className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>No recent alerts</p>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* System Health Overview */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5" />
                  System Health
                </CardTitle>
                <CardDescription>
                  Overall system performance indicators
                </CardDescription>
              </CardHeader>
              <CardContent>
                {summaryLoading ? (
                  <div className="space-y-4">
                    {[...Array(3)].map((_, i) => (
                      <div key={i} className="h-6 bg-gray-200 rounded animate-pulse" />
                    ))}
                  </div>
                ) : systemSummary ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">CPU Performance</span>
                      <div className="flex items-center gap-2">
                        {getTrendIcon(systemSummary.metrics.cpu.trend)}
                        <span className="text-sm text-muted-foreground">
                          {systemSummary.metrics.cpu.trend}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Memory Efficiency</span>
                      <div className="flex items-center gap-2">
                        {getTrendIcon(systemSummary.metrics.memory.trend)}
                        <span className="text-sm text-muted-foreground">
                          {systemSummary.metrics.memory.trend}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Storage Health</span>
                      <div className="flex items-center gap-2">
                        {getTrendIcon(systemSummary.metrics.disk.trend)}
                        <span className="text-sm text-muted-foreground">
                          {systemSummary.metrics.disk.trend}
                        </span>
                      </div>
                    </div>
                    <Separator />
                    <div className="text-center">
                      <p className="text-sm text-muted-foreground">
                        Last updated: {format(new Date(systemSummary.timestamp), 'MMM dd, HH:mm:ss')}
                      </p>
                    </div>
                  </div>
                ) : null}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Historical Metrics Tab */}
        <TabsContent value="metrics" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Historical Performance Metrics</CardTitle>
              <CardDescription>
                Time-series data for system resource usage and performance indicators
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center py-12 text-muted-foreground">
                <LineChart className="h-16 w-16 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-medium mb-2">Historical Metrics Charts</h3>
                <p>Interactive charts with historical data will be displayed here</p>
                <p className="text-sm mt-2">Feature implementation in progress...</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Alerts Tab */}
        <TabsContent value="alerts" className="space-y-6">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-2xl font-bold">Alert Management</h2>
              <p className="text-muted-foreground">Configure and manage alert rules and notifications</p>
            </div>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Create Alert Rule
            </Button>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Alert Rules</CardTitle>
                <CardDescription>
                  {alertRulesData?.alert_rules.length || 0} total rules,{' '}
                  {alertRulesData?.alert_rules.filter(r => r.is_enabled).length || 0} enabled
                </CardDescription>
              </CardHeader>
              <CardContent>
                {alertRulesLoading ? (
                  <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                      <div key={i} className="h-16 bg-gray-200 rounded animate-pulse" />
                    ))}
                  </div>
                ) : alertRulesData?.alert_rules.length ? (
                  <div className="space-y-3">
                    {alertRulesData.alert_rules.slice(0, 5).map((rule) => (
                      <div key={rule.id} className="flex items-center justify-between p-3 border rounded-lg">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <Badge variant={getSeverityColor(rule.severity)}>
                              {rule.severity}
                            </Badge>
                            <span className="font-medium">{rule.name}</span>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            {rule.metric_type} - {rule.metric_name}
                          </p>
                        </div>
                        <div className="text-right">
                          <Badge variant={rule.is_enabled ? 'default' : 'secondary'}>
                            {rule.is_enabled ? 'Enabled' : 'Disabled'}
                          </Badge>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <Settings className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>No alert rules configured</p>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Alert History</CardTitle>
                <CardDescription>
                  Recent alert instances and their resolution status
                </CardDescription>
              </CardHeader>
              <CardContent>
                {alertInstancesLoading ? (
                  <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                      <div key={i} className="h-16 bg-gray-200 rounded animate-pulse" />
                    ))}
                  </div>
                ) : alertInstancesData?.alert_instances.length ? (
                  <div className="space-y-3">
                    {alertInstancesData.alert_instances.slice(0, 5).map((alert) => (
                      <div key={alert.id} className="flex items-center justify-between p-3 border rounded-lg">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <Badge variant={getSeverityColor(alert.alert_rule?.severity || 'info')}>
                              {alert.alert_rule?.severity}
                            </Badge>
                            <span className="font-medium">{alert.alert_rule?.name}</span>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            Value: {alert.current_value} (threshold: {alert.threshold_value})
                          </p>
                        </div>
                        <div className="text-right">
                          <Badge variant={alert.status === 'triggered' ? 'destructive' : 'secondary'}>
                            {alert.status}
                          </Badge>
                          <p className="text-xs text-muted-foreground mt-1">
                            {format(new Date(alert.triggered_at), 'MMM dd, HH:mm')}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <Bell className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>No alert history</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Dashboards Tab */}
        <TabsContent value="dashboards" className="space-y-6">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-2xl font-bold">Custom Dashboards</h2>
              <p className="text-muted-foreground">Create and manage personalized monitoring dashboards</p>
            </div>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Create Dashboard
            </Button>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Available Dashboards</CardTitle>
              <CardDescription>
                {dashboardsData?.dashboards.length || 0} dashboards available
              </CardDescription>
            </CardHeader>
            <CardContent>
              {dashboardsLoading ? (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {[...Array(3)].map((_, i) => (
                    <div key={i} className="h-32 bg-gray-200 rounded animate-pulse" />
                  ))}
                </div>
              ) : dashboardsData?.dashboards.length ? (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {dashboardsData.dashboards.map((dashboard) => (
                    <Card key={dashboard.id} className="cursor-pointer hover:shadow-md transition-shadow">
                      <CardHeader>
                        <CardTitle className="text-lg">{dashboard.name}</CardTitle>
                        <CardDescription>
                          {dashboard.description || 'No description available'}
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="flex items-center justify-between">
                          <div className="flex gap-2">
                            {dashboard.is_default && (
                              <Badge variant="secondary">Default</Badge>
                            )}
                            {dashboard.is_public && (
                              <Badge variant="outline">Public</Badge>
                            )}
                          </div>
                          <span className="text-sm text-muted-foreground">
                            {dashboard.widgets?.length || 0} widgets
                          </span>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : (
                <div className="text-center py-12 text-muted-foreground">
                  <BarChart3 className="h-16 w-16 mx-auto mb-4 opacity-50" />
                  <h3 className="text-lg font-medium mb-2">No Dashboards</h3>
                  <p>Create your first custom dashboard to get started</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Insights Tab */}
        <TabsContent value="insights" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Performance Insights</CardTitle>
              <CardDescription>
                AI-powered recommendations and anomaly detection for your infrastructure
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center py-12 text-muted-foreground">
                <Activity className="h-16 w-16 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-medium mb-2">Intelligence Engine</h3>
                <p>AI-powered insights and recommendations will be displayed here</p>
                <p className="text-sm mt-2">Advanced analytics feature coming soon...</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
