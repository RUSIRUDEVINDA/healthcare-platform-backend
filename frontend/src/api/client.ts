import axios from 'axios';
import { redirectToLogin } from '../utils/navigation';

// When running with Nginx acting as Reverse Proxy locally via docker-compose,
// /api requests are automatically intercepted. 
// If running dev server without docker proxy, update baseURL to http://localhost:80/api (or wherever Nginx runs)
const apiClient = axios.create({
  baseURL: '/api', 
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to attach JWT Token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  // Default Content-Type is application/json; FormData must use multipart with a boundary (set by the runtime).
  if (config.data instanceof FormData && config.headers) {
    config.headers.delete('Content-Type');
  }
  return config;
});

// Response interceptor to handle unauthenticated logic globally (e.g., redirect to login)
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Only redirect if not already on the auth page
      if (!window.location.pathname.startsWith('/auth')) {
        localStorage.removeItem('access_token');
        redirectToLogin();
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient;
