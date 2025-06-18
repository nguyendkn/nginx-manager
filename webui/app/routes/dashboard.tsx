import { useAuth } from '~/contexts/AuthContext';
import { Button } from '~/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { Badge } from '~/components/ui/badge';
import { Avatar, AvatarFallback } from '~/components/ui/avatar';
import {
  User,
  Shield,
  Activity,
  Server,
  Globe,
  Lock,
  Settings,
  Plus,
  ArrowRight,
  LogOut,
  TrendingUp,
  Clock,
  AlertTriangle,
  CheckCircle,
  MemoryStick
} from 'lucide-react';
import { Link, useLoaderData } from 'react-router';
import { useEffect, useState } from 'react';
import { dashboardApi } from '~/services/api/dashboard';
import { Skeleton } from '~/components/ui/skeleton';
import { Progress } from '~/components/ui/progress';
import { Alert, AlertDescription } from '~/components/ui/alert';

export function meta() {
  return [
    { title: 'Dashboard - Nginx Manager' },
    { name: 'description', content: 'Nginx Manager Dashboard' },
  ];
}

interface DashboardData {
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

export async function loader() {
  try {
    const data = await dashboardApi.getDashboardStats();
    return { data, error: null };
  } catch (error) {
    return { data: null, error: error instanceof Error ? error.message : 'Failed to load dashboard data' };
  }
}

export default function Dashboard() {
  const { user, logout, isLoading } = useAuth();
  const { data: initialData, error } = useLoaderData<typeof loader>();
  const [data, setData] = useState<DashboardData | null>(initialData);
  const [isLoadingState, setIsLoadingState] = useState(!initialData);
  const [lastUpdated, setLastUpdated] = useState(new Date());

  // Auto-refresh dashboard data every 30 seconds
  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const newData = await dashboardApi.getDashboardStats();
        setData(newData);
        setLastUpdated(new Date());
      } catch (error) {
        console.error('Failed to refresh dashboard:', error);
      }
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  // Refresh function for manual updates
  const handleRefresh = async () => {
    setIsLoadingState(true);
    try {
      const newData = await dashboardApi.getDashboardStats();
      setData(newData);
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to refresh dashboard:', error);
    } finally {
      setIsLoadingState(false);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const userInitials = user?.name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase();

  if (error && !data) {
    return (
      <div className="container mx-auto p-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            {error}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            System overview and monitoring for your nginx proxy manager
          </p>
        </div>
        <div className="flex items-center gap-3">
          <div className="text-sm text-muted-foreground">
            Last updated: {lastUpdated.toLocaleTimeString()}
          </div>
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isLoadingState}>
            {isLoadingState ? <Skeleton className="h-4 w-4" /> : <TrendingUp className="h-4 w-4" />}
            Refresh
          </Button>
        </div>
      </div>

      {/* System Health Status */}
      {data?.health && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              System Health
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-center gap-3">
                <div className={`h-3 w-3 rounded-full ${data.health.nginx ? 'bg-green-500' : 'bg-red-500'}`} />
                <span className="font-medium">Nginx Service</span>
                <Badge variant={data.health.nginx ? "default" : "destructive"}>
                  {data.health.nginx ? "Running" : "Stopped"}
                </Badge>
              </div>
              <div className="flex items-center gap-3">
                <div className={`h-3 w-3 rounded-full ${data.health.database ? 'bg-green-500' : 'bg-red-500'}`} />
                <span className="font-medium">Database</span>
                <Badge variant={data.health.database ? "default" : "destructive"}>
                  {data.health.database ? "Connected" : "Disconnected"}
                </Badge>
              </div>
              <div className="flex items-center gap-3">
                <div className={`h-3 w-3 rounded-full ${data.health.storage ? 'bg-green-500' : 'bg-red-500'}`} />
                <span className="font-medium">Storage</span>
                <Badge variant={data.health.storage ? "default" : "destructive"}>
                  {data.health.storage ? "Available" : "Full"}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* System Metrics */}
      {data?.systemStats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">System Uptime</CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.systemStats.uptime}</div>
              <p className="text-xs text-muted-foreground">
                Since last restart
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
              <MemoryStick className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.systemStats.memoryUsage}%</div>
              <Progress value={data.systemStats.memoryUsage} className="mt-2" />
              <p className="text-xs text-muted-foreground mt-1">
                System memory utilization
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Disk Usage</CardTitle>
              <Server className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.systemStats.diskUsage}%</div>
              <Progress value={data.systemStats.diskUsage} className="mt-2" />
              <p className="text-xs text-muted-foreground mt-1">
                Storage space used
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Connections</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.systemStats.activeConnections}</div>
              <p className="text-xs text-muted-foreground">
                Current active connections
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Resource Counts */}
      {data?.counts && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <Link to="/proxy-hosts" className="block">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Proxy Hosts</CardTitle>
                <Globe className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{data.counts.proxyHosts}</div>
                <p className="text-xs text-muted-foreground">
                  {data.counts.activeHosts} active
                </p>
              </CardContent>
            </Link>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <Link to="/certificates" className="block">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">SSL Certificates</CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{data.counts.certificates}</div>
                <p className="text-xs text-muted-foreground">
                  Managed certificates
                </p>
              </CardContent>
            </Link>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Access Lists</CardTitle>
              <Settings className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.counts.accessLists}</div>
              <p className="text-xs text-muted-foreground">
                Security policies
              </p>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <Link to="/monitoring" className="block">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Monitoring</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">Live</div>
                <p className="text-xs text-muted-foreground">
                  Real-time metrics
                </p>
              </CardContent>
            </Link>
          </Card>
        </div>
      )}

      {/* Recent Activity */}
      {data?.recentActivity && (
        <Card>
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
            {data.recentActivity.length === 0 ? (
              <p className="text-muted-foreground text-center py-4">
                No recent activity
              </p>
            ) : (
              <div className="space-y-3">
                {data.recentActivity.slice(0, 5).map((activity) => (
                  <div key={activity.id} className="flex items-center gap-3 p-3 rounded-lg border">
                    <div className={`h-2 w-2 rounded-full ${
                      activity.status === 'success' ? 'bg-green-500' :
                      activity.status === 'warning' ? 'bg-yellow-500' : 'bg-red-500'
                    }`} />
                    <div className="flex-1">
                      <p className="text-sm font-medium">{activity.action}</p>
                      <p className="text-xs text-muted-foreground">{activity.resource}</p>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(activity.timestamp).toLocaleTimeString()}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Loading State */}
      {isLoadingState && !data && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {Array.from({ length: 8 }).map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-4 w-24" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-16" />
                <Skeleton className="h-3 w-32 mt-2" />
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
