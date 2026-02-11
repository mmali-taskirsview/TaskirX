import { useState, useEffect } from 'react';
import { analyticsAPI } from '../services/api';
import {
  Calendar,
  Download,
  RefreshCw,
  TrendingUp,
  TrendingDown,
  Eye,
  MousePointer,
  DollarSign,
  Target
} from 'lucide-react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';
import { Line, Bar, Doughnut } from 'react-chartjs-2';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const Analytics = () => {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [dateRange, setDateRange] = useState('7d');

  useEffect(() => {
    fetchAnalytics();
  }, [dateRange]);

  const fetchAnalytics = async () => {
    setLoading(true);
    try {
      const response = await analyticsAPI.getDashboard({ range: dateRange });
      setStats(response.data);
    } catch (error) {
      console.error('Failed to fetch analytics:', error);
      // Demo data
      setStats({
        impressions: { total: 1250000, change: 12.5 },
        clicks: { total: 45000, change: 8.2 },
        spend: { total: 15750, change: 15.3 },
        ctr: { total: 3.6, change: -2.1 },
        conversions: { total: 3200, change: 22.1 },
        revenue: { total: 28500, change: 18.7 }
      });
    } finally {
      setLoading(false);
    }
  };

  // Chart data
  const performanceData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Impressions',
        data: [165000, 172000, 168000, 185000, 190000, 178000, 192000],
        borderColor: '#0066FF',
        backgroundColor: 'rgba(0, 102, 255, 0.1)',
        fill: true,
        tension: 0.4,
        yAxisID: 'y',
      },
      {
        label: 'Clicks',
        data: [5900, 6200, 6100, 6800, 7100, 6500, 7200],
        borderColor: '#00FF00',
        backgroundColor: 'rgba(0, 255, 0, 0.1)',
        fill: true,
        tension: 0.4,
        yAxisID: 'y1',
      },
    ],
  };

  const revenueData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Revenue',
        data: [3200, 3800, 3500, 4200, 4800, 3900, 4500],
        backgroundColor: '#0066FF',
        borderRadius: 8,
      },
      {
        label: 'Spend',
        data: [2100, 2400, 2200, 2800, 3200, 2600, 3000],
        backgroundColor: '#00FF00',
        borderRadius: 8,
      },
    ],
  };

  const deviceData = {
    labels: ['Mobile', 'Desktop', 'Tablet', 'CTV'],
    datasets: [{
      data: [55, 30, 10, 5],
      backgroundColor: ['#0066FF', '#00FF00', '#FFB800', '#FF4444'],
      borderWidth: 0,
    }],
  };

  const geoData = {
    labels: ['United States', 'Canada', 'United Kingdom', 'Germany', 'France'],
    datasets: [{
      label: 'Impressions',
      data: [45, 20, 15, 12, 8],
      backgroundColor: 'rgba(0, 102, 255, 0.8)',
      borderRadius: 8,
    }],
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top',
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: { color: 'rgba(0, 0, 0, 0.05)' },
      },
      y1: {
        type: 'linear',
        display: true,
        position: 'right',
        grid: { drawOnChartArea: false },
      },
      x: {
        grid: { display: false },
      },
    },
  };

  const barOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { position: 'top' },
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: { color: 'rgba(0, 0, 0, 0.05)' },
      },
      x: {
        grid: { display: false },
      },
    },
  };

  const doughnutOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { position: 'right' },
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
    { title: 'Impressions', value: stats?.impressions?.total?.toLocaleString(), change: stats?.impressions?.change, icon: Eye, color: 'blue' },
    { title: 'Clicks', value: stats?.clicks?.total?.toLocaleString(), change: stats?.clicks?.change, icon: MousePointer, color: 'green' },
    { title: 'Spend', value: `$${stats?.spend?.total?.toLocaleString()}`, change: stats?.spend?.change, icon: DollarSign, color: 'purple' },
    { title: 'CTR', value: `${stats?.ctr?.total}%`, change: stats?.ctr?.change, icon: Target, color: 'orange' },
    { title: 'Conversions', value: stats?.conversions?.total?.toLocaleString(), change: stats?.conversions?.change, icon: Target, color: 'pink' },
    { title: 'Revenue', value: `$${stats?.revenue?.total?.toLocaleString()}`, change: stats?.revenue?.change, icon: DollarSign, color: 'teal' },
  ];

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Analytics</h1>
          <p className="text-gray-500">Track your advertising performance</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 bg-white border border-gray-200 rounded-lg px-3 py-2">
            <Calendar size={16} className="text-gray-400" />
            <select
              value={dateRange}
              onChange={(e) => setDateRange(e.target.value)}
              className="text-sm focus:outline-none"
            >
              <option value="7d">Last 7 days</option>
              <option value="30d">Last 30 days</option>
              <option value="90d">Last 90 days</option>
            </select>
          </div>
          <button
            onClick={fetchAnalytics}
            className="p-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <RefreshCw size={18} className="text-gray-600" />
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-cyber-blue text-white rounded-lg hover:bg-blue-600 transition-colors">
            <Download size={16} />
            Export
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {statCards.map((card, index) => {
          const Icon = card.icon;
          const isPositive = card.change >= 0;
          return (
            <div key={index} className="bg-white rounded-xl p-4 shadow-sm">
              <div className="flex items-center justify-between mb-2">
                <Icon size={18} className="text-gray-400" />
                <div className={`flex items-center gap-1 text-xs ${isPositive ? 'text-green-600' : 'text-red-600'}`}>
                  {isPositive ? <TrendingUp size={12} /> : <TrendingDown size={12} />}
                  {Math.abs(card.change)}%
                </div>
              </div>
              <p className="text-xl font-bold text-gray-900">{card.value}</p>
              <p className="text-xs text-gray-500">{card.title}</p>
            </div>
          );
        })}
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Performance Trend</h3>
          <div className="h-80">
            <Line data={performanceData} options={chartOptions} />
          </div>
        </div>
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Revenue vs Spend</h3>
          <div className="h-80">
            <Bar data={revenueData} options={barOptions} />
          </div>
        </div>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Device Distribution</h3>
          <div className="h-64">
            <Doughnut data={deviceData} options={doughnutOptions} />
          </div>
        </div>
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Top Geos</h3>
          <div className="h-64">
            <Bar 
              data={geoData} 
              options={{
                ...barOptions,
                indexAxis: 'y',
                plugins: { legend: { display: false } }
              }} 
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default Analytics;
