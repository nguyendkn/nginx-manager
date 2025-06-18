import { useAuth, useRequireAuth } from '../contexts/AuthContext';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '../components/ui/avatar';
import { LogOut, User, Shield, Settings } from 'lucide-react';

export function meta() {
  return [
    { title: 'Dashboard - Nginx Manager' },
    { name: 'description', content: 'Nginx Manager Dashboard' },
  ];
}

export default function Dashboard() {
  const { user, logout, isLoading } = useRequireAuth();

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

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-4">
              <h1 className="text-2xl font-bold text-gray-900">Nginx Manager</h1>
              <Badge variant="secondary">Dashboard</Badge>
            </div>

            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <Avatar className="h-8 w-8">
                  <AvatarImage src={user?.avatar} />
                  <AvatarFallback>
                    {user?.name?.charAt(0) || user?.email?.charAt(0) || '?'}
                  </AvatarFallback>
                </Avatar>
                <div className="hidden md:block">
                  <p className="text-sm font-medium text-gray-900">{user?.name || user?.email}</p>
                  <p className="text-xs text-gray-500">{user?.role}</p>
                </div>
              </div>

              <Button
                variant="outline"
                size="sm"
                onClick={handleLogout}
                className="flex items-center space-x-2"
              >
                <LogOut className="h-4 w-4" />
                <span>Logout</span>
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        <div className="space-y-6">
          {/* Welcome Card */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <User className="h-5 w-5" />
                <span>Welcome back, {user?.name || user?.email}!</span>
              </CardTitle>
              <CardDescription>
                You are successfully authenticated and connected to the Nginx Manager API.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <h4 className="font-medium">User Information</h4>
                  <div className="text-sm text-gray-600 space-y-1">
                    <p><strong>Email:</strong> {user?.email}</p>
                    <p><strong>Name:</strong> {user?.name || 'Not set'}</p>
                    <p><strong>Role:</strong>
                      <Badge
                        variant={user?.role === 'admin' ? 'default' : 'secondary'}
                        className="ml-1"
                      >
                        {user?.role === 'admin' && <Shield className="h-3 w-3 mr-1" />}
                        {user?.role}
                      </Badge>
                    </p>
                    <p><strong>Status:</strong>
                      <Badge
                        variant={user?.is_disabled ? 'destructive' : 'default'}
                        className="ml-1"
                      >
                        {user?.is_disabled ? 'Disabled' : 'Active'}
                      </Badge>
                    </p>
                    <p><strong>Created:</strong> {new Date(user?.created_at || '').toLocaleDateString()}</p>
                  </div>
                </div>

                <div className="space-y-2">
                  <h4 className="font-medium">API Status</h4>
                  <div className="text-sm text-gray-600 space-y-1">
                    <p><strong>API Connection:</strong>
                      <Badge variant="default" className="ml-1">Connected</Badge>
                    </p>
                    <p><strong>Authentication:</strong>
                      <Badge variant="default" className="ml-1">Active</Badge>
                    </p>
                    <p><strong>Environment:</strong> Development</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Settings className="h-5 w-5" />
                <span>Quick Actions</span>
              </CardTitle>
              <CardDescription>
                Common tasks and navigation
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Button variant="outline" className="h-20 flex-col space-y-2">
                  <div className="text-lg">🔧</div>
                  <span>Proxy Hosts</span>
                </Button>

                <Button variant="outline" className="h-20 flex-col space-y-2">
                  <div className="text-lg">🔒</div>
                  <span>SSL Certificates</span>
                </Button>

                <Button variant="outline" className="h-20 flex-col space-y-2">
                  <div className="text-lg">👥</div>
                  <span>Access Lists</span>
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Debug Info (Development only) */}
          <Card className="border-dashed border-gray-300">
            <CardHeader>
              <CardTitle className="text-sm text-gray-500">
                Debug Information (Development)
              </CardTitle>
            </CardHeader>
            <CardContent>
              <details className="text-xs">
                <summary className="cursor-pointer text-gray-500 hover:text-gray-700">
                  View user object
                </summary>
                <pre className="mt-2 p-2 bg-gray-100 rounded overflow-auto">
                  {JSON.stringify(user, null, 2)}
                </pre>
              </details>
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  );
}
