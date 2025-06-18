import { useEffect } from 'react';
import { useNavigate } from 'react-router';
import { useAuth } from '../contexts/AuthContext';

export function meta() {
  return [
    { title: "Nginx Manager" },
    { name: "description", content: "Nginx Proxy Manager" },
  ];
}

export default function Home() {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        navigate('/dashboard');
      } else {
        navigate('/login');
      }
    }
  }, [isAuthenticated, isLoading, navigate]);

  // Show loading while checking authentication
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

  return null;
}
