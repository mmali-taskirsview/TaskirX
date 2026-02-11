/**
 * Dashboard Component Tests
 * React component tests using Jest and React Testing Library
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import Dashboard from '../pages/Dashboard';
import { AuthContext } from '../context/AuthContext';

/**
 * Mock authentication context
 */
const mockAuthContext = {
  user: {
    userId: 'test-user-123',
    email: 'test@example.com',
    firstName: 'Test',
    lastName: 'User'
  },
  token: 'test-token',
  logout: jest.fn()
};

/**
 * Test wrapper with providers
 */
const DashboardWrapper = ({ children }) => (
  <BrowserRouter>
    <AuthContext.Provider value={mockAuthContext}>
      {children || <Dashboard />}
    </AuthContext.Provider>
  </BrowserRouter>
);

describe('Dashboard Component', () => {
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock API calls
    global.fetch = jest.fn();
  });
  
  test('should render dashboard with main sections', () => {
    render(<DashboardWrapper />);
    
    expect(screen.getByText(/Dashboard/i)).toBeInTheDocument();
    expect(screen.getByText(/Campaigns/i)).toBeInTheDocument();
    expect(screen.getByText(/Analytics/i)).toBeInTheDocument();
  });
  
  test('should display key metrics cards', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        impressions: 10000,
        clicks: 500,
        conversions: 50,
        spend: 5000
      })
    });
    
    render(<DashboardWrapper />);
    
    await waitFor(() => {
      expect(screen.getByText(/Impressions/i)).toBeInTheDocument();
      expect(screen.getByText(/Clicks/i)).toBeInTheDocument();
      expect(screen.getByText(/Conversions/i)).toBeInTheDocument();
    });
  });
  
  test('should navigate between tabs', async () => {
    render(<DashboardWrapper />);
    
    const analyticsTab = screen.getByRole('tab', { name: /Analytics/i });
    fireEvent.click(analyticsTab);
    
    await waitFor(() => {
      expect(screen.getByText(/Analytics/i)).toBeInTheDocument();
    });
  });
  
  test('should handle logout', async () => {
    render(<DashboardWrapper />);
    
    const logoutButton = screen.getByRole('button', { name: /Logout/i });
    fireEvent.click(logoutButton);
    
    expect(mockAuthContext.logout).toHaveBeenCalled();
  });
});

/**
 * Campaign Management Component Tests
 */
import CampaignManagement from '../components/dashboard/CampaignManagement';

describe('CampaignManagement Component', () => {
  
  const mockCampaigns = [
    {
      id: '1',
      name: 'Campaign 1',
      budget: 5000,
      status: 'active',
      impressions: 1000,
      clicks: 100
    },
    {
      id: '2',
      name: 'Campaign 2',
      budget: 3000,
      status: 'paused',
      impressions: 500,
      clicks: 25
    }
  ];
  
  beforeEach(() => {
    global.fetch = jest.fn();
  });
  
  test('should render campaign list', () => {
    render(
      <DashboardWrapper>
        <CampaignManagement campaigns={mockCampaigns} />
      </DashboardWrapper>
    );
    
    expect(screen.getByText('Campaign 1')).toBeInTheDocument();
    expect(screen.getByText('Campaign 2')).toBeInTheDocument();
  });
  
  test('should filter campaigns by status', async () => {
    render(
      <DashboardWrapper>
        <CampaignManagement campaigns={mockCampaigns} />
      </DashboardWrapper>
    );
    
    const filterSelect = screen.getByRole('combobox', { name: /Status/i });
    await userEvent.selectOptions(filterSelect, 'active');
    
    expect(screen.getByText('Campaign 1')).toBeInTheDocument();
    expect(screen.queryByText('Campaign 2')).not.toBeInTheDocument();
  });
  
  test('should open create campaign modal', async () => {
    render(
      <DashboardWrapper>
        <CampaignManagement campaigns={mockCampaigns} />
      </DashboardWrapper>
    );
    
    const createButton = screen.getByRole('button', { name: /Create Campaign/i });
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText(/New Campaign/i)).toBeInTheDocument();
    });
  });
  
  test('should submit campaign form', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: '3', name: 'New Campaign' })
    });
    
    render(
      <DashboardWrapper>
        <CampaignManagement campaigns={mockCampaigns} />
      </DashboardWrapper>
    );
    
    const createButton = screen.getByRole('button', { name: /Create Campaign/i });
    fireEvent.click(createButton);
    
    const nameInput = await screen.findByLabelText(/Campaign Name/i);
    const budgetInput = await screen.findByLabelText(/Budget/i);
    
    await userEvent.type(nameInput, 'New Campaign');
    await userEvent.type(budgetInput, '5000');
    
    const submitButton = screen.getByRole('button', { name: /Submit/i });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/campaigns'),
        expect.any(Object)
      );
    });
  });
});

/**
 * Analytics Dashboard Tests
 */
import AnalyticsDashboard from '../components/dashboard/AnalyticsDashboard';

describe('AnalyticsDashboard Component', () => {
  
  const mockAnalytics = {
    impressions: 10000,
    clicks: 500,
    conversions: 50,
    ctr: 0.05,
    crr: 0.1,
    spent: 5000,
    revenue: 7500,
    roi: 0.5
  };
  
  test('should render analytics metrics', () => {
    render(
      <DashboardWrapper>
        <AnalyticsDashboard data={mockAnalytics} />
      </DashboardWrapper>
    );
    
    expect(screen.getByText(/10000/)).toBeInTheDocument(); // impressions
    expect(screen.getByText(/500/)).toBeInTheDocument(); // clicks
  });
  
  test('should update date range', async () => {
    render(
      <DashboardWrapper>
        <AnalyticsDashboard data={mockAnalytics} />
      </DashboardWrapper>
    );
    
    const dateRangeSelect = screen.getByRole('combobox', { name: /Date Range/i });
    await userEvent.selectOptions(dateRangeSelect, '7d');
    
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/analytics'),
        expect.any(Object)
      );
    });
  });
  
  test('should export data as CSV', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      blob: async () => new Blob(['data'], { type: 'text/csv' })
    });
    
    render(
      <DashboardWrapper>
        <AnalyticsDashboard data={mockAnalytics} />
      </DashboardWrapper>
    );
    
    const exportButton = screen.getByRole('button', { name: /Export/i });
    fireEvent.click(exportButton);
    
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/export'),
        expect.any(Object)
      );
    });
  });
});

/**
 * Auth Component Tests
 */
import LoginForm from '../components/auth/LoginForm';

describe('LoginForm Component', () => {
  
  const mockOnLogin = jest.fn();
  
  beforeEach(() => {
    jest.clearAllMocks();
    global.fetch = jest.fn();
  });
  
  test('should render login form', () => {
    render(
      <DashboardWrapper>
        <LoginForm onLogin={mockOnLogin} />
      </DashboardWrapper>
    );
    
    expect(screen.getByLabelText(/Email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Login/i })).toBeInTheDocument();
  });
  
  test('should validate email format', async () => {
    render(
      <DashboardWrapper>
        <LoginForm onLogin={mockOnLogin} />
      </DashboardWrapper>
    );
    
    const emailInput = screen.getByLabelText(/Email/i);
    await userEvent.type(emailInput, 'invalid-email');
    
    const submitButton = screen.getByRole('button', { name: /Login/i });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(/valid email/i)).toBeInTheDocument();
    });
  });
  
  test('should submit login form', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ token: 'test-token', userId: '123' })
    });
    
    render(
      <DashboardWrapper>
        <LoginForm onLogin={mockOnLogin} />
      </DashboardWrapper>
    );
    
    const emailInput = screen.getByLabelText(/Email/i);
    const passwordInput = screen.getByLabelText(/Password/i);
    
    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'SecurePass123!');
    
    const submitButton = screen.getByRole('button', { name: /Login/i });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockOnLogin).toHaveBeenCalled();
    });
  });
  
  test('should display error on login failure', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: 'Invalid credentials' })
    });
    
    render(
      <DashboardWrapper>
        <LoginForm onLogin={mockOnLogin} />
      </DashboardWrapper>
    );
    
    const emailInput = screen.getByLabelText(/Email/i);
    const passwordInput = screen.getByLabelText(/Password/i);
    
    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.type(passwordInput, 'WrongPassword');
    
    const submitButton = screen.getByRole('button', { name: /Login/i });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(/Invalid credentials/i)).toBeInTheDocument();
    });
  });
});
