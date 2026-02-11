import axios from 'axios';

const API_BASE_URL = '/api';

// Create axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  login: (email, password) => api.post('/auth/login', { email, password }),
  register: (data) => api.post('/auth/register', data),
  me: () => api.get('/auth/me'),
  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }
};

// Campaigns API
export const campaignsAPI = {
  getAll: (params) => api.get('/campaigns', { params }),
  getById: (id) => api.get(`/campaigns/${id}`),
  create: (data) => api.post('/campaigns', data),
  update: (id, data) => api.put(`/campaigns/${id}`, data),
  delete: (id) => api.delete(`/campaigns/${id}`),
  getStats: (id) => api.get(`/campaigns/${id}/stats`),
  pause: (id) => api.post(`/campaigns/${id}/pause`),
  resume: (id) => api.post(`/campaigns/${id}/resume`)
};

// Analytics API
export const analyticsAPI = {
  getDashboard: (params) => api.get('/analytics/dashboard', { params }),
  getTimeSeries: (params) => api.get('/analytics/timeseries', { params }),
  getGeoStats: (params) => api.get('/analytics/geo', { params }),
  getFunnel: (params) => api.get('/analytics/funnel', { params }),
  getTopCampaigns: (params) => api.get('/analytics/top-campaigns', { params })
};

// Users API (Admin)
export const usersAPI = {
  getAll: (params) => api.get('/users', { params }),
  getById: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
  suspend: (id) => api.post(`/users/${id}/suspend`),
  activate: (id) => api.post(`/users/${id}/activate`)
};

// Bids API
export const bidsAPI = {
  getAll: (params) => api.get('/bids', { params }),
  getById: (id) => api.get(`/bids/${id}`),
  getStats: () => api.get('/bids/stats')
};

// RTB API
export const rtbAPI = {
  getBidRequest: (data) => api.post('/rtb/bid-request', data),
  getStats: () => api.get('/rtb/stats')
};

// Health API
export const healthAPI = {
  check: () => api.get('/health')
};

export default api;
