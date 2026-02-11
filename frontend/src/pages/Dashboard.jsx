import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { analyticsAPI, campaignsAPI } from '../services/api';
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Eye,
  MousePointer,
  Target,
  ArrowRight,
  RefreshCw
} from 'lucide-react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';
import { Line, Bar } from 'react-chartjs-2';

// Register ChartJS
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const Dashboard = () => {
  const [stats, setStats] = useState(null);
  const [campaigns, setCampaigns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchData = async () => {
    try {
      const [dashboardRes, campaignsRes] = await Promise.all([
        analyticsAPI.getDashboard().catch(() => ({ data: null })),
        campaignsAPI.getAll({ limit: 5 }).catch(() => ({ data: { campaigns: [] } }))
      ]);
      
      setStats(dashboardRes.data || {
        totalCampaigns: 12,
        activeCampaigns: 8,
        totalImpressions: 1250000,
        totalClicks: 45000,
        totalSpend: 15750.50,
        avgCTR: 3.6,
        avgCPM: 12.60,
        todayImpressions: 85000,
        todayClicks: 3200,
        todaySpend: 1250.00
      });
      
      setCampaigns(campaignsRes.data?.campaigns || []);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      // Set mock data for demo
      setStats({
        totalCampaigns: 12,
        activeCampaigns: 8,
        totalImpressions: 1250000,
        totalClicks: 45000,
        totalSpend: 15750.50,
        avgCTR: 3.6,
        avgCPM: 12.60,
        todayImpressions: 85000,
        todayClicks: 3200,
        todaySpend: 1250.00
      });
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleRefresh = () => {
    setRefreshing(true);
    fetchData();
  };

  // Chart data
  const impressionsChartData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Impressions',
        data: [65000, 72000, 68000, 85000, 90000, 78000, 85000],
        borderColor: '#0066FF',
        backgroundColor: 'rgba(0, 102, 255, 0.1)',
        fill: true,
        tension: 0.4,
      },
    ],
  };

  const revenueChartData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Revenue',
        data: [1200, 1500, 1350, 1800, 2100, 1650, 1900],
        backgroundColor: '#00FF00',
        borderRadius: 8,
      },
    ],
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: {
          color: 'rgba(0, 0, 0, 0.05)',
        },
      },
      x: {
        grid: {
          display: false,
        },
      },
    },
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-cyber-blue"></div>
      </div>
    );
  }

  const statCards = [
    {
      title: 'Total Impressions',
      value: stats?.totalImpressions?.toLocaleString() || '0',
      change: '+12.5%',
      positive: true,
      icon: Eye,
      color: 'bg-blue-500',
    },
    {
      title: 'Total Clicks',
      value: stats?.totalClicks?.toLocaleString() || '0',
      change: '+8.2%',
      positive: true,
      icon: MousePointer,
      color: 'bg-green-500',
    },
    {
      title: 'Total Spend',
      value: `$${stats?.totalSpend?.toLocaleString() || '0'}`,
      change: '+15.3%',
      positive: true,
      icon: DollarSign,
      color: 'bg-purple-500',
    },
    {
      title: 'Avg CTR',
      value: `${stats?.avgCTR || '0'}%`,
      change: '-2.1%',
      positive: false,
      icon: Target,
      color: 'bg-orange-500',
    },
  ];

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-500">Welcome back! Here's what's happening.</p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
        >
          <RefreshCw size={16} className={refreshing ? 'animate-spin' : ''} />
          Refresh
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((card, index) => {
          const Icon = card.icon;
          return (
            <div key={index} className="bg-white rounded-xl p-6 shadow-sm card-hover">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-500">{card.title}</p>
                  <p className="text-2xl font-bold text-gray-900 mt-1">{card.value}</p>
                  <div className={`flex items-center gap-1 mt-2 text-sm ${card.positive ? 'text-green-600' : 'text-red-600'}`}>
                    {card.positive ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                    <span>{card.change}</span>
                    <span className="text-gray-400">vs last week</span>
                  </div>
                </div>
                <div className={`${card.color} p-3 rounded-lg`}>
                  <Icon size={24} className="text-white" />
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Impressions Chart */}
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900">Impressions Trend</h3>
            <select className="text-sm border border-gray-200 rounded-lg px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-cyber-blue">
              <option>Last 7 days</option>
              <option>Last 30 days</option>
              <option>Last 90 days</option>
            </select>
          </div>
          <div className="h-64">
            <Line data={impressionsChartData} options={chartOptions} />
          </div>
        </div>

        {/* Revenue Chart */}
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900">Daily Revenue</h3>
            <select className="text-sm border border-gray-200 rounded-lg px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-cyber-blue">
              <option>Last 7 days</option>
              <option>Last 30 days</option>
              <option>Last 90 days</option>
            </select>
          </div>
          <div className="h-64">
            <Bar data={revenueChartData} options={chartOptions} />
          </div>
        </div>
      </div>

      {/* Bottom Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Campaigns */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm">
          <div className="flex items-center justify-between p-6 border-b border-gray-100">
            <h3 className="text-lg font-semibold text-gray-900">Recent Campaigns</h3>
            <Link to="/campaigns" className="text-cyber-blue text-sm hover:underline flex items-center gap-1">
              View all <ArrowRight size={14} />
            </Link>
          </div>
          <div className="p-6">
            {campaigns.length > 0 ? (
              <div className="space-y-4">
                {campaigns.map((campaign) => (
                  <div key={campaign._id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                    <div>
                      <p className="font-medium text-gray-900">{campaign.name}</p>
                      <p className="text-sm text-gray-500">{campaign.type || 'Display'}</p>
                    </div>
                    <div className="text-right">
                      <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                        campaign.status === 'active' ? 'bg-green-100 text-green-800' :
                        campaign.status === 'paused' ? 'bg-yellow-100 text-yellow-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {campaign.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500">No campaigns yet</p>
                <Link to="/campaigns/new" className="text-cyber-blue hover:underline text-sm">
                  Create your first campaign
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Quick Stats */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-6">Today's Overview</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
              <div className="flex items-center gap-3">
                <Eye className="text-blue-600" size={20} />
                <span className="text-gray-700">Impressions</span>
              </div>
              <span className="font-semibold text-gray-900">{stats?.todayImpressions?.toLocaleString() || '85,000'}</span>
            </div>
            <div className="flex items-center justify-between p-4 bg-green-50 rounded-lg">
              <div className="flex items-center gap-3">
                <MousePointer className="text-green-600" size={20} />
                <span className="text-gray-700">Clicks</span>
              </div>
              <span className="font-semibold text-gray-900">{stats?.todayClicks?.toLocaleString() || '3,200'}</span>
            </div>
            <div className="flex items-center justify-between p-4 bg-purple-50 rounded-lg">
              <div className="flex items-center gap-3">
                <DollarSign className="text-purple-600" size={20} />
                <span className="text-gray-700">Spend</span>
              </div>
              <span className="font-semibold text-gray-900">${stats?.todaySpend?.toLocaleString() || '1,250'}</span>
            </div>
            <div className="flex items-center justify-between p-4 bg-orange-50 rounded-lg">
              <div className="flex items-center gap-3">
                <Target className="text-orange-600" size={20} />
                <span className="text-gray-700">Active Campaigns</span>
              </div>
              <span className="font-semibold text-gray-900">{stats?.activeCampaigns || '8'}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
