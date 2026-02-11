import axios from 'axios'

const API_BASE_URL = process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:3000/api'

// Create axios instances
export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Consolidated AI Services into Main Backend
// export const fraudClient = ... (Deprecated)
// export const matchingClient = ... (Deprecated)
// export const optimizationClient = ... (Deprecated)

// Add request interceptor for authentication
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Add response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API methods
export const api = {
  // Auth
  login: (email: string, password: string) =>
    apiClient.post('/auth/login', { email, password }),
  register: (data: { email: string; password: string; name: string }) =>
    apiClient.post('/auth/register', data),
  
  // Campaigns
  getCampaigns: () => apiClient.get('/campaigns'),
  getCampaign: (id: string) => apiClient.get(`/campaigns/${id}`),
  createCampaign: (data: any) => apiClient.post('/campaigns', data),
  updateCampaign: (id: string, data: any) => apiClient.patch(`/campaigns/${id}`, data),
  deleteCampaign: (id: string) => apiClient.delete(`/campaigns/${id}`),
  
  // Analytics
  getAnalytics: (params?: any) => apiClient.get('/analytics', { params }),
  getDashboardStats: () => apiClient.get('/analytics/dashboard'),
  
  // SSP Dashboard & Analytics
  getSSPDashboard: (publisherId?: string) => 
    apiClient.get('/ssp/dashboard', { params: { publisherId } }),
  getInventoryStats: (publisherId: string) => 
    apiClient.get(`/ssp/inventory/stats/${publisherId}`),
  
  // Fraud Detection (Migrated to Main API)
  checkFraud: (data: any) => apiClient.post('/ai/fraud/detect', data),
  getFraudMetrics: () => apiClient.get('/ai/anomalies'), // Mapped to anomalies
  
  // Ad Matching (Migrated to Main API)
  matchAds: (data: any) => apiClient.post('/ai/match', data),
  getMatchingMetrics: () => Promise.resolve({ data: {} }), // Placeholder
  
  // Bid Optimization (Migrated to Main API)
  optimizeBid: (data: any) => apiClient.post('/ai/bid/optimize', data),
  calculatePacing: (data: any) => Promise.resolve({ data: {} }), // Placeholder
  getOptimizationMetrics: () => Promise.resolve({ data: {} }), // Placeholder
  submitFeedback: (data: any) => Promise.resolve({ data: {} }), // Placeholder

  // SSP - Publishers
  getPublishers: () => apiClient.get('/ssp/publishers'),
  getPublisher: (id: string) => apiClient.get(`/ssp/publishers/${id}`),
  createPublisher: (data: any) => apiClient.post('/ssp/publishers', data),
  updatePublisher: (id: string, data: any) => apiClient.patch(`/ssp/publishers/${id}`, data),
  deletePublisher: (id: string) => apiClient.delete(`/ssp/publishers/${id}`),
  getPublisherStats: (id: string) => apiClient.get(`/ssp/publishers/${id}/stats`),

  // SSP - Ad Units (Inventory)
  getAdUnits: (publisherId?: string) => 
    apiClient.get('/ssp/inventory/ad-units', { params: { publisherId } }),
  getAdUnit: (id: string) => apiClient.get(`/ssp/inventory/ad-units/${id}`),
  createAdUnit: (data: any) => apiClient.post('/ssp/inventory/ad-units', data),
  updateAdUnit: (id: string, data: any) => apiClient.patch(`/ssp/inventory/ad-units/${id}`, data),
  deleteAdUnit: (id: string) => apiClient.delete(`/ssp/inventory/ad-units/${id}`),

  // SSP - Placements
  getPlacements: (publisherId?: string) => 
    apiClient.get('/ssp/inventory/placements', { params: { publisherId } }),
  getPlacement: (id: string) => apiClient.get(`/ssp/inventory/placements/${id}`),
  createPlacement: (data: any) => apiClient.post('/ssp/inventory/placements', data),
  updatePlacement: (id: string, data: any) => apiClient.patch(`/ssp/inventory/placements/${id}`, data),
  deletePlacement: (id: string) => apiClient.delete(`/ssp/inventory/placements/${id}`),

  // SSP - Floor Prices
  getFloorPrices: (publisherId?: string) => 
    apiClient.get('/ssp/inventory/floor-prices', { params: { publisherId } }),
  getFloorPrice: (id: string) => apiClient.get(`/ssp/inventory/floor-prices/${id}`),
  createFloorPrice: (data: any) => apiClient.post('/ssp/inventory/floor-prices', data),
  updateFloorPrice: (id: string, data: any) => apiClient.patch(`/ssp/inventory/floor-prices/${id}`, data),
  deleteFloorPrice: (id: string) => apiClient.delete(`/ssp/inventory/floor-prices/${id}`),

  // SSP - Demand Partners
  getDemandPartners: (publisherId?: string) => 
    apiClient.get('/ssp/demand-partners', { params: { publisherId } }),
  getDemandPartner: (id: string) => apiClient.get(`/ssp/demand-partners/${id}`),
  createDemandPartner: (data: any) => apiClient.post('/ssp/demand-partners', data),
  updateDemandPartner: (id: string, data: any) => apiClient.patch(`/ssp/demand-partners/${id}`, data),
  deleteDemandPartner: (id: string) => apiClient.delete(`/ssp/demand-partners/${id}`),
  getDemandPartnerTemplates: () => apiClient.get('/ssp/demand-partners/templates'),

  // SSP - Brand Safety Rules
  getBrandSafetyRules: (publisherId?: string) => 
    apiClient.get('/ssp/inventory/brand-safety', { params: { publisherId } }),
  getBrandSafetyRule: (id: string) => apiClient.get(`/ssp/inventory/brand-safety/${id}`),
  createBrandSafetyRule: (data: any) => apiClient.post('/ssp/inventory/brand-safety', data),
  updateBrandSafetyRule: (id: string, data: any) => apiClient.put(`/ssp/inventory/brand-safety/${id}`, data),
  deleteBrandSafetyRule: (id: string) => apiClient.delete(`/ssp/inventory/brand-safety/${id}`),
  toggleBrandSafetyRule: (id: string) => apiClient.post(`/ssp/inventory/brand-safety/${id}/toggle`),

  // SSP - Health & Auction
  getSSPHealth: () => apiClient.get('/ssp/health'),
  runAuction: (data: any) => apiClient.post('/ssp/auction', data),

  // Billing
  getWallet: () => apiClient.get('/billing/wallet'),
  getBalance: () => apiClient.get('/billing/balance'),
  deposit: (amount: number, description?: string) => 
    apiClient.post('/billing/deposit', { amount, description }),
  getTransactions: (limit?: number) => 
    apiClient.get('/billing/transactions', { params: { limit } }),

  // Users
  getUsers: () => apiClient.get('/users'),
  getUser: (id: string) => apiClient.get(`/users/${id}`),
  createUser: (data: any) => apiClient.post('/users', data),
  updateUser: (id: string, data: any) => apiClient.patch(`/users/${id}`, data),
  deleteUser: (id: string) => apiClient.delete(`/users/${id}`),

  // ==================== DSP APIs ====================
  
  // DSP Dashboard
  getDSPDashboard: () => apiClient.get('/dsp/dashboard'),
  getRTBAnalytics: () => apiClient.get('/dsp/rtb-analytics'),

  // DSP - Supply Partners (SSP connections)
  getSupplyPartners: () => apiClient.get('/dsp/supply-partners'),
  getSupplyPartner: (id: string) => apiClient.get(`/dsp/supply-partners/${id}`),
  createSupplyPartner: (data: any) => apiClient.post('/dsp/supply-partners', data),
  updateSupplyPartner: (id: string, data: any) => apiClient.put(`/dsp/supply-partners/${id}`, data),
  deleteSupplyPartner: (id: string) => apiClient.delete(`/dsp/supply-partners/${id}`),

  // DSP - Audience Segments
  getAudiences: (advertiserId?: string) => 
    apiClient.get('/dsp/audiences', { params: { advertiserId } }),
  getAudience: (id: string) => apiClient.get(`/dsp/audiences/${id}`),
  createAudience: (data: any) => apiClient.post('/dsp/audiences', data),
  updateAudience: (id: string, data: any) => apiClient.put(`/dsp/audiences/${id}`, data),
  deleteAudience: (id: string) => apiClient.delete(`/dsp/audiences/${id}`),

  // DSP - Deals (PMP, Preferred, Guaranteed)
  getDeals: (advertiserId?: string) => 
    apiClient.get('/dsp/deals', { params: { advertiserId } }),
  getDeal: (id: string) => apiClient.get(`/dsp/deals/${id}`),
  createDeal: (data: any) => apiClient.post('/dsp/deals', data),
  updateDeal: (id: string, data: any) => apiClient.put(`/dsp/deals/${id}`, data),
  deleteDeal: (id: string) => apiClient.delete(`/dsp/deals/${id}`),

  // DSP - Bid Strategies
  getBidStrategies: (advertiserId?: string) => 
    apiClient.get('/dsp/bid-strategies', { params: { advertiserId } }),
  getBidStrategy: (id: string) => apiClient.get(`/dsp/bid-strategies/${id}`),
  createBidStrategy: (data: any) => apiClient.post('/dsp/bid-strategies', data),
  updateBidStrategy: (id: string, data: any) => apiClient.put(`/dsp/bid-strategies/${id}`, data),
  deleteBidStrategy: (id: string) => apiClient.delete(`/dsp/bid-strategies/${id}`),

  // DSP - Bidding Operations
  processBidRequest: (bidRequest: any) => apiClient.post('/dsp/bid', bidRequest),
  recordWin: (bidId: string, supplyPartnerId: string, price: number) => 
    apiClient.post('/dsp/win', { bidId, supplyPartnerId, price }),
}
