import { useState } from 'react';
import { Download, Calendar, FileText, Filter, RefreshCw } from 'lucide-react';
import toast from 'react-hot-toast';

const Reports = () => {
  const [reportType, setReportType] = useState('performance');
  const [dateRange, setDateRange] = useState('7d');
  const [generating, setGenerating] = useState(false);

  const reportTypes = [
    { id: 'performance', name: 'Performance Report', description: 'Impressions, clicks, CTR, spend' },
    { id: 'campaign', name: 'Campaign Report', description: 'Detailed campaign metrics' },
    { id: 'revenue', name: 'Revenue Report', description: 'Revenue and earnings breakdown' },
    { id: 'geo', name: 'Geographic Report', description: 'Performance by location' },
    { id: 'device', name: 'Device Report', description: 'Performance by device type' },
    { id: 'bidding', name: 'Bidding Report', description: 'RTB auction analytics' }
  ];

  const recentReports = [
    { id: 1, name: 'Performance Report - Jan 2026', type: 'performance', date: '2026-01-27', size: '2.4 MB' },
    { id: 2, name: 'Campaign Report - Week 4', type: 'campaign', date: '2026-01-25', size: '1.8 MB' },
    { id: 3, name: 'Revenue Report - Q1', type: 'revenue', date: '2026-01-20', size: '3.2 MB' },
    { id: 4, name: 'Geographic Report - US', type: 'geo', date: '2026-01-18', size: '1.1 MB' }
  ];

  const handleGenerate = async () => {
    setGenerating(true);
    // Simulate report generation
    await new Promise(resolve => setTimeout(resolve, 2000));
    setGenerating(false);
    toast.success('Report generated successfully!');
  };

  const handleDownload = (report) => {
    toast.success(`Downloading ${report.name}...`);
  };

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Reports</h1>
        <p className="text-gray-500">Generate and download analytics reports</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Report Generator */}
        <div className="lg:col-span-2 space-y-6">
          {/* Report Type Selection */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Generate New Report</h3>
            
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
              {reportTypes.map((type) => (
                <button
                  key={type.id}
                  onClick={() => setReportType(type.id)}
                  className={`p-4 border rounded-lg text-left transition-all ${
                    reportType === type.id
                      ? 'border-cyber-blue bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <p className="font-medium text-gray-900">{type.name}</p>
                  <p className="text-xs text-gray-500 mt-1">{type.description}</p>
                </button>
              ))}
            </div>

            {/* Date Range */}
            <div className="flex flex-col sm:flex-row gap-4 mb-6">
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 mb-2">Date Range</label>
                <div className="flex items-center gap-2 border border-gray-200 rounded-lg px-3 py-2">
                  <Calendar size={18} className="text-gray-400" />
                  <select
                    value={dateRange}
                    onChange={(e) => setDateRange(e.target.value)}
                    className="flex-1 focus:outline-none"
                  >
                    <option value="7d">Last 7 days</option>
                    <option value="30d">Last 30 days</option>
                    <option value="90d">Last 90 days</option>
                    <option value="ytd">Year to date</option>
                    <option value="custom">Custom range</option>
                  </select>
                </div>
              </div>
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 mb-2">Format</label>
                <select className="w-full border border-gray-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-cyber-blue">
                  <option value="csv">CSV</option>
                  <option value="xlsx">Excel (XLSX)</option>
                  <option value="pdf">PDF</option>
                  <option value="json">JSON</option>
                </select>
              </div>
            </div>

            {/* Generate Button */}
            <button
              onClick={handleGenerate}
              disabled={generating}
              className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-cyber-blue text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50"
            >
              {generating ? (
                <>
                  <RefreshCw size={18} className="animate-spin" />
                  Generating Report...
                </>
              ) : (
                <>
                  <FileText size={18} />
                  Generate Report
                </>
              )}
            </button>
          </div>

          {/* Recent Reports */}
          <div className="bg-white rounded-xl shadow-sm">
            <div className="p-4 border-b border-gray-100 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900">Recent Reports</h3>
              <button className="text-sm text-cyber-blue hover:underline">View all</button>
            </div>
            <div className="divide-y divide-gray-100">
              {recentReports.map((report) => (
                <div key={report.id} className="p-4 flex items-center justify-between hover:bg-gray-50">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                      <FileText size={20} className="text-cyber-blue" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{report.name}</p>
                      <p className="text-sm text-gray-500">{report.date} • {report.size}</p>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDownload(report)}
                    className="flex items-center gap-2 px-3 py-1.5 text-cyber-blue hover:bg-blue-50 rounded-lg transition-colors"
                  >
                    <Download size={16} />
                    Download
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Scheduled Reports */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Scheduled Reports</h3>
            <div className="space-y-3">
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center justify-between mb-1">
                  <p className="font-medium text-gray-900">Weekly Performance</p>
                  <span className="text-xs text-green-600 bg-green-100 px-2 py-0.5 rounded">Active</span>
                </div>
                <p className="text-sm text-gray-500">Every Monday at 9:00 AM</p>
              </div>
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center justify-between mb-1">
                  <p className="font-medium text-gray-900">Monthly Revenue</p>
                  <span className="text-xs text-green-600 bg-green-100 px-2 py-0.5 rounded">Active</span>
                </div>
                <p className="text-sm text-gray-500">1st of each month</p>
              </div>
            </div>
            <button className="w-full mt-4 py-2 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 transition-colors">
              + Add Schedule
            </button>
          </div>

          {/* Quick Stats */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">This Month</h3>
            <div className="space-y-4">
              <div className="flex justify-between">
                <span className="text-gray-500">Reports Generated</span>
                <span className="font-semibold">24</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Total Downloads</span>
                <span className="font-semibold">156</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Storage Used</span>
                <span className="font-semibold">48.2 MB</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Reports;
