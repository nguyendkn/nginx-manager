import { useState, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'react-hot-toast';
import {
  monitoringApi,
  type DashboardStats,
  type SystemMetrics,
  type NginxStatus,
  type ActivityEvent,
  type WebSocketMessage,
  type NginxControlRequest,
  formatBytes,
  formatPercentage,
  formatUptime
} from '../services/api/monitoring';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Badge } from '../components/ui/badge';
import { Separator } from '../components/ui/separator';
import { Progress } from '../components/ui/progress';
import { ScrollArea } from '../components/ui/scroll-area';
import {
  Activity,
  Server,
  Cpu,
  MemoryStick,
  HardDrive,
  Network,
  Clock,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Play,
  Square,
  RotateCcw,
  RefreshCw,
  TestTube
} from 'lucide-react';

export default function MonitoringPage() {
  const queryClient = useQueryClient();
  const [realTimeMetrics, setRealTimeMetrics] = useState<SystemMetrics | null>(null);
  const [realTimeNginxStatus, setRealTimeNginxStatus] = useState<NginxStatus | null>(null);
  const [realtimeActivities, setRealtimeActivities] = useState<ActivityEvent[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  // Query for initial dashboard data
  const { data: dashboardStats, isLoading, error } = useQuery({
    queryKey: ['monitoring', 'dashboard'],
    queryFn: monitoringApi.getDashboardStats,
    refetchInterval: 30000, // Fallback polling every 30 seconds
  });

  // Mutation for nginx control
  const nginxControlMutation = useMutation({
    mutationFn: (request: NginxControlRequest) => monitoringApi.controlNginx(request),
    onSuccess: (data) => {
      toast.success(data.message);
      // Refetch nginx status after control action
      queryClient.invalidateQueries({ queryKey: ['monitoring', 'nginx-status'] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to control nginx service');
    },
  });

  // WebSocket connection for real-time updates
  useEffect(() => {
    const connectWebSocket = () => {
      try {
        const ws = monitoringApi.createWebSocket(
          (message: WebSocketMessage) => {
            switch (message.type) {
              case 'metrics':
                setRealTimeMetrics(message.data);
                break;
              case 'nginx_status':
                setRealTimeNginxStatus(message.data);
                break;
              case 'activity':
                setRealtimeActivities(prev => [message.data, ...prev.slice(0, 9)]);
                break;
              default:
                console.log('Unknown WebSocket message type:', message.type);
            }
          },
          (error) => {
            console.error('WebSocket error:', error);
            setIsConnected(false);
            // Attempt reconnection after 5 seconds
            setTimeout(connectWebSocket, 5000);
          },
          (event) => {
            console.log('WebSocket closed:', event);
            setIsConnected(false);
            // Attempt reconnection after 3 seconds if not intentional
            if (!event.wasClean) {
              setTimeout(connectWebSocket, 3000);
            }
          }
        );

        ws.onopen = () => {
          console.log('WebSocket connected');
          setIsConnected(true);
        };

        wsRef.current = ws;
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
      }
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Handle nginx control actions
  const handleNginxControl = (action: NginxControlRequest['action']) => {
    nginxControlMutation.mutate({ action });
  };

  // Use real-time data if available, otherwise fallback to polled data
  const currentMetrics = realTimeMetrics || dashboardStats?.system_metrics;
  const currentNginxStatus = realTimeNginxStatus || dashboardStats?.nginx_status;
  const currentActivities = realtimeActivities.length > 0 ? realtimeActivities : dashboardStats?.recent_activity.activities || [];

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <XCircle className="h-16 w-16 text-destructive mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">Failed to load monitoring data</h2>
            <p className="text-muted-foreground">Please check your connection and try again.</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">System Monitoring</h1>
          <p className="text-muted-foreground">
            Real-time system metrics and nginx status
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={isConnected ? "default" : "destructive"}>
            {isConnected ? "Live" : "Offline"}
          </Badge>
          {isLoading && (
            <Badge variant="outline">
              <RefreshCw className="h-3 w-3 mr-1 animate-spin" />
              Loading
            </Badge>
          )}
        </div>
      </div>

      {/* System Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* CPU Usage */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">CPU Usage</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {currentMetrics ? formatPercentage(currentMetrics.cpu.usage) : '...'}
            </div>
            {currentMetrics && (
              <Progress value={currentMetrics.cpu.usage} className="mt-2" />
            )}
            <p className="text-xs text-muted-foreground mt-2">
              Load: {currentMetrics?.cpu.load_avg_1.toFixed(2) || '...'}
            </p>
          </CardContent>
        </Card>

        {/* Memory Usage */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
            <MemoryStick className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {currentMetrics ? formatPercentage(currentMetrics.memory.used_percent) : '...'}
            </div>
            {currentMetrics && (
              <Progress value={currentMetrics.memory.used_percent} className="mt-2" />
            )}
            <p className="text-xs text-muted-foreground mt-2">
              {currentMetrics ? formatBytes(currentMetrics.memory.used) : '...'} / {currentMetrics ? formatBytes(currentMetrics.memory.total) : '...'}
            </p>
          </CardContent>
        </Card>

        {/* Disk Usage */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Disk Usage</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {currentMetrics ? formatPercentage(currentMetrics.disk.used_percent) : '...'}
            </div>
            {currentMetrics && (
              <Progress value={currentMetrics.disk.used_percent} className="mt-2" />
            )}
            <p className="text-xs text-muted-foreground mt-2">
              {currentMetrics ? formatBytes(currentMetrics.disk.used) : '...'} / {currentMetrics ? formatBytes(currentMetrics.disk.total) : '...'}
            </p>
          </CardContent>
        </Card>

        {/* Uptime */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Uptime</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {currentMetrics ? formatUptime(currentMetrics.process.uptime) : '...'}
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              {currentMetrics?.process.goroutines || '...'} goroutines
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Dashboard Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Nginx Status */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Nginx Status
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">Service Status</span>
              <Badge variant={currentNginxStatus?.running ? "default" : "destructive"}>
                {currentNginxStatus?.running ? (
                  <>
                    <CheckCircle className="h-3 w-3 mr-1" />
                    Running
                  </>
                ) : (
                  <>
                    <XCircle className="h-3 w-3 mr-1" />
                    Stopped
                  </>
                )}
              </Badge>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">Configuration</span>
              <Badge variant={currentNginxStatus?.config_test ? "default" : "destructive"}>
                {currentNginxStatus?.config_test ? (
                  <>
                    <CheckCircle className="h-3 w-3 mr-1" />
                    Valid
                  </>
                ) : (
                  <>
                    <AlertTriangle className="h-3 w-3 mr-1" />
                    Error
                  </>
                )}
              </Badge>
            </div>

            {currentNginxStatus?.version && (
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Version</span>
                <span className="text-sm text-muted-foreground">{currentNginxStatus.version}</span>
              </div>
            )}

            {currentNginxStatus?.pid && (
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">PID</span>
                <span className="text-sm text-muted-foreground">{currentNginxStatus.pid}</span>
              </div>
            )}

            <Separator />

            <div className="space-y-2">
              <p className="text-sm font-medium">Control Actions</p>
              <div className="grid grid-cols-2 gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleNginxControl('start')}
                  disabled={nginxControlMutation.isPending}
                >
                  <Play className="h-3 w-3 mr-1" />
                  Start
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleNginxControl('stop')}
                  disabled={nginxControlMutation.isPending}
                >
                  <Square className="h-3 w-3 mr-1" />
                  Stop
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleNginxControl('restart')}
                  disabled={nginxControlMutation.isPending}
                >
                  <RotateCcw className="h-3 w-3 mr-1" />
                  Restart
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleNginxControl('reload')}
                  disabled={nginxControlMutation.isPending}
                >
                  <RefreshCw className="h-3 w-3 mr-1" />
                  Reload
                </Button>
              </div>
              <Button
                size="sm"
                variant="outline"
                className="w-full"
                onClick={() => handleNginxControl('test')}
                disabled={nginxControlMutation.isPending}
              >
                <TestTube className="h-3 w-3 mr-1" />
                Test Configuration
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* System Details */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              System Details
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {currentMetrics && (
              <>
                <div className="space-y-2">
                  <p className="text-sm font-medium">CPU Load Average</p>
                  <div className="grid grid-cols-3 gap-2 text-sm">
                    <div className="text-center">
                      <div className="font-mono">{currentMetrics.cpu.load_avg_1.toFixed(2)}</div>
                      <div className="text-xs text-muted-foreground">1min</div>
                    </div>
                    <div className="text-center">
                      <div className="font-mono">{currentMetrics.cpu.load_avg_5.toFixed(2)}</div>
                      <div className="text-xs text-muted-foreground">5min</div>
                    </div>
                    <div className="text-center">
                      <div className="font-mono">{currentMetrics.cpu.load_avg_15.toFixed(2)}</div>
                      <div className="text-xs text-muted-foreground">15min</div>
                    </div>
                  </div>
                </div>

                <Separator />

                <div className="space-y-2">
                  <p className="text-sm font-medium">Memory Breakdown</p>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span>Total</span>
                      <span className="font-mono">{formatBytes(currentMetrics.memory.total)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Used</span>
                      <span className="font-mono">{formatBytes(currentMetrics.memory.used)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Available</span>
                      <span className="font-mono">{formatBytes(currentMetrics.memory.available)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Go Allocated</span>
                      <span className="font-mono">{formatBytes(currentMetrics.memory.go_alloc)}</span>
                    </div>
                  </div>
                </div>

                <Separator />

                <div className="space-y-2">
                  <p className="text-sm font-medium">Network</p>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span>Bytes Received</span>
                      <span className="font-mono">{formatBytes(currentMetrics.network.bytes_recv)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Bytes Sent</span>
                      <span className="font-mono">{formatBytes(currentMetrics.network.bytes_sent)}</span>
                    </div>
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Activity Feed */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Recent Activity
            </CardTitle>
            <CardDescription>
              Latest system events and changes
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ScrollArea className="h-[400px]">
              <div className="space-y-4">
                {currentActivities.length > 0 ? (
                  currentActivities.map((activity) => (
                    <div key={activity.id} className="border-l-2 border-muted pl-4 pb-4">
                      <div className="flex items-center gap-2 mb-1">
                        <Badge variant="outline" className="text-xs">
                          {activity.type}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {new Date(activity.timestamp).toLocaleTimeString()}
                        </span>
                      </div>
                      <p className="text-sm font-medium">{activity.message}</p>
                      {activity.details && Object.keys(activity.details).length > 0 && (
                        <div className="mt-1 text-xs text-muted-foreground">
                          {Object.entries(activity.details).map(([key, value]) => (
                            <div key={key}>
                              {key}: {String(value)}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="text-center text-muted-foreground py-8">
                    <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">No recent activity</p>
                  </div>
                )}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
