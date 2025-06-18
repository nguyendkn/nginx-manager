import { ErrorBoundary as ReactErrorBoundary } from 'react-error-boundary';
import type { ReactNode } from 'react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { AlertCircle, RefreshCw } from 'lucide-react';

interface ErrorFallbackProps {
  error: Error;
  resetErrorBoundary: () => void;
}

function ErrorFallback({ error, resetErrorBoundary }: ErrorFallbackProps) {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <AlertCircle className="h-12 w-12 text-destructive" />
          </div>
          <CardTitle className="text-xl">Something went wrong</CardTitle>
          <CardDescription>
            An unexpected error occurred. Please try refreshing the page.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <details className="text-sm">
            <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
              Error details
            </summary>
            <pre className="mt-2 p-2 bg-muted rounded text-xs overflow-auto">
              {error.message}
            </pre>
          </details>
          <div className="flex gap-2">
            <Button
              onClick={resetErrorBoundary}
              className="flex-1"
              variant="default"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Try again
            </Button>
            <Button
              onClick={() => window.location.reload()}
              variant="outline"
              className="flex-1"
            >
              Reload page
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function logError(error: Error, errorInfo: any) {
  console.error('Error caught by boundary:', error);
  console.error('Component stack:', errorInfo.componentStack);

  // In production, you might want to send this to an error reporting service
  // like Sentry, LogRocket, etc.
}

interface ErrorBoundaryProps {
  children: ReactNode;
  onReset?: () => void;
  fallback?: React.ComponentType<ErrorFallbackProps>;
}

export function ErrorBoundary({
  children,
  onReset,
  fallback: FallbackComponent = ErrorFallback
}: ErrorBoundaryProps) {
  return (
    <ReactErrorBoundary
      FallbackComponent={FallbackComponent}
      onError={logError}
      onReset={onReset}
      resetKeys={[]} // Add any values that should trigger a reset when changed
    >
      {children}
    </ReactErrorBoundary>
  );
}

// HOC for wrapping components with error boundary
export function withErrorBoundary<T extends object>(
  Component: React.ComponentType<T>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
) {
  return function WrappedComponent(props: T) {
    return (
      <ErrorBoundary {...errorBoundaryProps}>
        <Component {...props} />
      </ErrorBoundary>
    );
  };
}

export default ErrorBoundary;
