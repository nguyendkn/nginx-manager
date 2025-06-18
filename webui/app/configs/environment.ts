/**
 * Environment configuration for web application
 * Uses Vite's import.meta.env for environment variables
 */

interface Environment {
  API_URL: string;
  API_VERSION: string;
  API_BASE_PATH: string;
}

// Get environment variables from Vite (types defined in vite-env.d.ts)
const env = import.meta.env;

export const environment: Environment = {
  // Use relative URL in development to leverage Vite proxy
  // Use full URL in production or when VITE_API_URL is explicitly set
  API_URL: env.VITE_API_URL ?? "http://localhost:5000",
  API_VERSION: env.VITE_API_VERSION ?? "v3.1",
  API_BASE_PATH: `/api/${env.VITE_API_VERSION ?? "v3.1"}`,
};

// Export individual values for convenience
export const { API_URL, API_VERSION, API_BASE_PATH } = environment;

export default environment;
