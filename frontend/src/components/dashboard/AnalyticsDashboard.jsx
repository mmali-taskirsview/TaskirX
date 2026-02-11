import React, { useState, useEffect } from 'react';
import { Calendar, Download, TrendingUp, BarChart3 } from 'lucide-react';

const AnalyticsDashboard = ({ stats }) => {
  const [dateRange, setDateRange] = useState('7days');
  const [analyticsData, setAnalyticsData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedMetrics, setSelectedMetrics] = useState({
    impressions: true,
    clicks: true,
    conversions: true,
    spend: true
  });

  useEffect(() => {
    fetchAnalytics();
  }, [dateRange]);

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/analytics?range=${dateRange}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch analytics');

      const data = await response.json();
      setAnalyticsData(data);
      setError(null);
    } catch (err) {
      console.error('Error fetching analytics:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleExport = async () => {
    try {
      const response = await fetch(`/api/analytics/export?range=${dateRange}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to export data');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `analytics-${dateRange}-${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      setError(err.message);
    }
  };

  const toggleMetric = (metric) => {
    setSelectedMetrics(prev => ({
      ...prev,
      [metric]: !prev[metric]
    }));
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="animate-spin inline-block">
          <svg className="h-12 w-12 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          </svg>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">📈 Analytics Dashboard</h2>
        <div className="flex gap-3">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="24h">Last 24 Hours</option>
            <option value="7days">Last 7 Days</option>
            <option value="30days">Last 30 Days</option>
            <option value="90days">Last 90 Days</option>
            <option value="all">All Time</option>
          </select>
          <button
            onClick={handleExport}
            className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded flex items-center gap-2 transition-colors"
          >
            <Download size={20} />
            Export
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
          {error}
        </div>
      )}

      {/* Key Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard
          title="Total Impressions"
          value={analyticsData?.impressions?.toLocaleString() || '0'}
          change={analyticsData?.impressionChange}
          icon="👁️"
          checked={selectedMetrics.impressions}
          onToggle={() => toggleMetric('impressions')}
        />
        <MetricCard
          title="Total Clicks"
          value={analyticsData?.clicks?.toLocaleString() || '0'}
          change={analyticsData?.clickChange}
          icon="🖱️"
          checked={selectedMetrics.clicks}
          onToggle={() => toggleMetric('clicks')}
        />
        <MetricCard
          title="Total Conversions"
          value={analyticsData?.conversions?.toLocaleString() || '0'}
          change={analyticsData?.conversionChange}
          icon="✅"
          checked={selectedMetrics.conversions}
          onToggle={() => toggleMetric('conversions')}
        />
        <MetricCard
          title="Total Spend"
          value={`$${(analyticsData?.spend / 100 || 0).toFixed(2)}`}
          change={analyticsData?.spendChange}
          icon="💰"
          checked={selectedMetrics.spend}
          onToggle={() => toggleMetric('spend')}
        />
      </div>

      {/* Performance Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* CTR and Conversion Rate */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-bold text-gray-900 mb-4">📊 Performance Metrics</h3>
          <div className="space-y-4">
            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-gray-700 font-medium">Click-Through Rate (CTR)</span>
                <span className="text-2xl font-bold text-blue-600">
                  {analyticsData?.ctr ? (analyticsData.ctr * 100).toFixed(2) : '0.00'}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full"
                  style={{ width: `${Math.min((analyticsData?.ctr || 0) * 1000, 100)}%` }}
                ></div>
              </div>
            </div>

            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-gray-700 font-medium">Conversion Rate</span>
                <span className="text-2xl font-bold text-green-600">
                  {analyticsData?.conversionRate ? (analyticsData.conversionRate * 100).toFixed(2) : '0.00'}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-green-600 h-2 rounded-full"
                  style={{ width: `${Math.min((analyticsData?.conversionRate || 0) * 1000, 100)}%` }}
                ></div>
              </div>
            </div>

            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-gray-700 font-medium">Cost Per Click (CPC)</span>
                <span className="text-2xl font-bold text-purple-600">
                  ${((analyticsData?.spend || 0) / (analyticsData?.clicks || 1) / 100).toFixed(4)}
                </span>
              </div>
            </div>

            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-gray-700 font-medium">Cost Per Conversion (CPA)</span>
                <span className="text-2xl font-bold text-orange-600">
                  ${((analyticsData?.spend || 0) / (analyticsData?.conversions || 1) / 100).toFixed(2)}
                </span>
              </div>
            </div>

            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-gray-700 font-medium">Return on Ad Spend (ROAS)</span>
                <span className="text-2xl font-bold text-indigo-600">
                  {analyticsData?.roas ? analyticsData.roas.toFixed(2) : '0.00'}x
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Campaign Breakdown */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-bold text-gray-900 mb-4">🎯 Top Campaigns</h3>
          <div className="space-y-3">
            {analyticsData?.topCampaigns && analyticsData.topCampaigns.length > 0 ? (
              analyticsData.topCampaigns.map((campaign, idx) => (
                <div key={idx} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                  <div className="flex-1">
                    <p className="font-medium text-gray-900">{campaign.name}</p>
                    <p className="text-sm text-gray-500">{campaign.clicks} clicks</p>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-gray-900">{campaign.conversionRate}%</p>
                    <p className="text-sm text-gray-500">conversion</p>
                  </div>
                </div>
              ))
            ) : (
              <p className="text-gray-500">No campaign data available</p>
            )}
          </div>
        </div>
      </div>

      {/* Detailed Stats Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="p-6 border-b border-gray-200">
          <h3 className="text-lg font-bold text-gray-900">📋 Detailed Statistics</h3>
        </div>
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Date</th>
              {selectedMetrics.impressions && (
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Impressions</th>
              )}
              {selectedMetrics.clicks && (
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Clicks</th>
              )}
              {selectedMetrics.conversions && (
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Conversions</th>
              )}
              {selectedMetrics.spend && (
                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Spend</th>
              )}
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">CTR</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {analyticsData?.dailyStats && analyticsData.dailyStats.length > 0 ? (
              analyticsData.dailyStats.map((day, idx) => (
                <tr key={idx} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-gray-900 font-medium">
                    {new Date(day.date).toLocaleDateString()}
                  </td>
                  {selectedMetrics.impressions && (
                    <td className="px-6 py-4 text-gray-900">{day.impressions.toLocaleString()}</td>
                  )}
                  {selectedMetrics.clicks && (
                    <td className="px-6 py-4 text-gray-900">{day.clicks.toLocaleString()}</td>
                  )}
                  {selectedMetrics.conversions && (
                    <td className="px-6 py-4 text-gray-900">{day.conversions.toLocaleString()}</td>
                  )}
                  {selectedMetrics.spend && (
                    <td className="px-6 py-4 text-gray-900">${(day.spend / 100).toFixed(2)}</td>
                  )}
                  <td className="px-6 py-4 text-gray-900">{((day.clicks / day.impressions) * 100).toFixed(2)}%</td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="6" className="px-6 py-8 text-center text-gray-500">
                  No data available for this period
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

const MetricCard = ({ title, value, change, icon, checked, onToggle }) => (
  <div className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow cursor-pointer" onClick={onToggle}>
    <div className="flex items-start justify-between">
      <div className="flex-1">
        <p className="text-gray-600 text-sm font-medium">{title}</p>
        <p className="text-3xl font-bold text-gray-900 mt-2">{value}</p>
        {change !== undefined && (
          <div className={`text-sm font-medium mt-2 flex items-center gap-1 ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
            <TrendingUp size={16} className={change < 0 ? 'rotate-180' : ''} />
            {change >= 0 ? '+' : ''}{change}%
          </div>
        )}
      </div>
      <div className="flex items-center gap-2">
        <span className="text-3xl">{icon}</span>
        <input
          type="checkbox"
          checked={checked}
          onChange={onToggle}
          onClick={(e) => e.stopPropagation()}
          className="w-4 h-4 rounded"
        />
      </div>
    </div>
  </div>
);

export default AnalyticsDashboard;
