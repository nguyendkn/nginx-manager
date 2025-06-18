import { API_VERSION, API_BASE_PATH } from '~/configs/environment';

const API_VERSION_CACHED = API_VERSION;
const API_BASE_PATH_CACHED = API_BASE_PATH;

export const API_CONFIG = {
  VERSION: API_VERSION_CACHED,
  BASE_PATH: API_BASE_PATH_CACHED,
} as const;

/**
 * Helper function to build full API URL
 */
export const buildApiUrl = (path: string): string => {
  return `${API_CONFIG.BASE_PATH}${path}`;
};

/**
 * Helper function to build API URL with query parameters
 */
export const buildApiUrlWithParams = (
  path: string,
  params?: Record<string, any>
): string => {
  const baseUrl = buildApiUrl(path);
  if (!params || Object.keys(params).length === 0) {
    return baseUrl;
  }

  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      searchParams.append(key, String(value));
    }
  });

  const queryString = searchParams.toString();
  return queryString ? `${baseUrl}?${queryString}` : baseUrl;
};

export const STORAGE_PATHS = {
  FILE_PREVIEW: (id: string) => `/storage/files/${id}/preview`,
};
