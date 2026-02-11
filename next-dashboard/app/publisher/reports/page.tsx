'use client';

import React, { useState, useEffect } from 'react';
import {
  FileText,
  Download,
  Calendar,
  Filter,
  TrendingUp,
  DollarSign,
  BarChart3,
  PieChart,
  Clock,
  CheckCircle,
  AlertCircle,
  ChevronDown,
  RefreshCw,
  Mail,
  Plus,
  Loader2
} from 'lucide-react';
import { api } from '@/lib/api';

interface DashboardMetrics {
  totalAdUnits: number;
  activeAdUnits: number;
  totalPlacements: number;
  activeDemandPartners: number;
  metrics: {
    totalImpressions: number;
    totalRequests: number;
    totalRevenue: number;
    fillRate: string;
    ecpm: string;
  };
}

interface Report {
  id: string;
  name: string;
  type: string;
  dateRange: string;
  status: string;
  createdAt: string;
  size: string;
  data?: any;
}

export default function PublisherReportsPage() {
  const [dateRange, setDateRange] = useState('last_7_days');
  const [reportType, setReportType] = useState('all');
  const [loading, setLoading] = useState(true);
  const [dashboard, setDashboard] = useState<DashboardMetrics | null>(null);
  const [reports, setReports] = useState<Report[]>([]);
  const [generating, setGenerating] = useState(false);

  // Fetch dashboard data on mount
  useEffect(() => {
    async function fetchData() {
      try {
        setLoading(true);
        const response = await api.getSSPDashboard();
        setDashboard(response.data);
        
        // Generate reports from real metrics
        if (response.data) {
          const now = new Date();
          const generatedReports: Report[] = [
            {
              id: 'rpt_live',
              name: 'Live Performance Report',
              type: 'Performance',
              dateRange: `As of ${now.toLocaleDateString()}`,
              status: 'ready',
              createdAt: now.toISOString(),
              size: '1.2 KB',
              data: response.data.metrics
            },
            {
              id: 'rpt_revenue',
              name: `Revenue Summary - ${now.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}`,
              type: 'Revenue',
              dateRange: `${now.toLocaleDateString('en-US', { month: 'short' })} 1 - ${now.toLocaleDateString()}`,
              status: 'ready',
              createdAt: now.toISOString(),
              size: '2.1 KB',
              data: { revenue: response.data.metrics.totalRevenue, ecpm: response.data.metrics.ecpm }
            },
            {
              id: 'rpt_inventory',
              name: 'Inventory Utilization Report',
              type: 'Inventory',
              dateRange: 'Current',
              status: 'ready',
              createdAt: now.toISOString(),
              size: '0.8 KB',
              data: { 
                totalAdUnits: response.data.totalAdUnits,
                activeAdUnits: response.data.activeAdUnits,
                fillRate: response.data.metrics.fillRate
              }
            },
            {
              id: 'rpt_partners',
              name: 'Demand Partner Performance',
              type: 'Partners',
              dateRange: 'Current',
              status: 'ready',
              createdAt: now.toISOString(),
              size: '1.5 KB',
              data: { activeDemandPartners: response.data.activeDemandPartners }
            }
          ];
          setReports(generatedReports);
        }
      } catch (err) {
        console.error('Error fetching data:', err);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  // Generate a new report
  const handleGenerateReport = async (templateId: string) => {
    setGenerating(true);
    try {
      // Refresh dashboard data and create new report
      const response = await api.getSSPDashboard();
      const now = new Date();
      const newReport: Report = {
        id: `rpt_${Date.now()}`,
        name: `${templateId.charAt(0).toUpperCase() + templateId.slice(1)} Report - ${now.toLocaleString()}`,
        type: templateId.charAt(0).toUpperCase() + templateId.slice(1),
        dateRange: now.toLocaleDateString(),
        status: 'ready',
        createdAt: now.toISOString(),
        size: '1.0 KB',
        data: response.data?.metrics
      };
      setReports([newReport, ...reports]);
    } catch (err) {
      console.error('Error generating report:', err);
    } finally {
      setGenerating(false);
    }
  };

  // Download report as JSON
  const handleDownload = (report: Report) => {
    const dataStr = JSON.stringify(report.data || report, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${report.name.replace(/\s+/g, '_')}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const scheduledReports = [
    {
      id: 1,
      name: 'Weekly Revenue Summary',
      frequency: 'Weekly (Monday)',
      nextRun: 'Feb 10, 2026',
      format: 'JSON',
      recipients: ['publisher@example.com']
    },
    {
      id: 2,
      name: 'Monthly Performance Report',
      frequency: 'Monthly (1st)',
      nextRun: 'Mar 1, 2026',
      format: 'JSON',
      recipients: ['publisher@example.com', 'finance@example.com']
    }
  ];

  const reportTemplates = [
    { id: 'revenue', name: 'Revenue Report', icon: DollarSign, description: 'Detailed revenue breakdown by ad unit, partner, and time' },
    { id: 'performance', name: 'Performance Report', icon: TrendingUp, description: 'Impressions, fill rate, eCPM, and viewability metrics' },
    { id: 'partners', name: 'Demand Partner Report', icon: BarChart3, description: 'Partner performance comparison and optimization insights' },
    { id: 'inventory', name: 'Inventory Report', icon: PieChart, description: 'Ad unit utilization and inventory availability analysis' },
  ];

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ready':
        return (
          <span className="flex items-center gap-1 px-2 py-1 bg-green-100 text-green-700 rounded-full text-xs font-medium">
            <CheckCircle className="w-3 h-3" />
            Ready
          </span>
        );
      case 'processing':
        return (
          <span className="flex items-center gap-1 px-2 py-1 bg-yellow-100 text-yellow-700 rounded-full text-xs font-medium">
            <RefreshCw className="w-3 h-3 animate-spin" />
            Processing
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1 px-2 py-1 bg-gray-100 text-gray-700 rounded-full text-xs font-medium">
            <AlertCircle className="w-3 h-3" />
            Unknown
          </span>
        );
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
        <span className="ml-2 text-gray-600">Loading reports...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Reports</h1>
          <p className="text-gray-600 mt-1">Generate and download detailed reports from real data</p>
        </div>
        {dashboard && (
          <div className="text-right text-sm text-gray-500">
            <p>Total Revenue: <span className="font-semibold text-emerald-600">${Number(dashboard.metrics.totalRevenue).toLocaleString()}</span></p>
            <p>eCPM: <span className="font-semibold">${parseFloat(dashboard.metrics.ecpm).toFixed(2)}</span></p>
          </div>
        )}
      </div>

      {/* Quick Report Templates */}
      <div className="grid grid-cols-4 gap-4">
        {reportTemplates.map((template) => (
          <button
            key={template.id}
            onClick={() => handleGenerateReport(template.id)}
            disabled={generating}
            className="p-4 bg-white border border-gray-200 rounded-xl hover:border-emerald-300 hover:bg-emerald-50 transition-all text-left group disabled:opacity-50"
          >
            <div className="w-10 h-10 bg-emerald-100 rounded-lg flex items-center justify-center mb-3 group-hover:bg-emerald-200">
              <template.icon className="w-5 h-5 text-emerald-600" />
            </div>
            <h3 className="font-medium text-gray-900">{template.name}</h3>
            <p className="text-sm text-gray-500 mt-1">{template.description}</p>
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl border border-gray-200 p-4">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Calendar className="w-5 h-5 text-gray-400" />
            <select
              value={dateRange}
              onChange={(e) => setDateRange(e.target.value)}
              className="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="last_7_days">Last 7 Days</option>
              <option value="last_30_days">Last 30 Days</option>
              <option value="last_90_days">Last 90 Days</option>
              <option value="this_month">This Month</option>
              <option value="last_month">Last Month</option>
              <option value="custom">Custom Range</option>
            </select>
          </div>
          <div className="flex items-center gap-2">
            <Filter className="w-5 h-5 text-gray-400" />
            <select
              value={reportType}
              onChange={(e) => setReportType(e.target.value)}
              className="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Report Types</option>
              <option value="revenue">Revenue</option>
              <option value="performance">Performance</option>
              <option value="partners">Partners</option>
              <option value="inventory">Inventory</option>
            </select>
          </div>
          <div className="flex-1" />
          <button 
            onClick={() => window.location.reload()}
            className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:bg-gray-50 rounded-lg transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
            Refresh
          </button>
        </div>
      </div>

      {/* Recent Reports */}
      <div className="bg-white rounded-xl border border-gray-200">
        <div className="p-4 border-b border-gray-200">
          <h2 className="font-semibold text-gray-900">Generated Reports ({reports.length})</h2>
        </div>
        {reports.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <FileText className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p>No reports generated yet</p>
            <p className="text-sm mt-1">Click a template above to generate your first report</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-600">Report Name</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-600">Type</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-600">Date Range</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-600">Status</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-600">Size</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-600">Created</th>
                  <th className="text-right px-4 py-3 text-sm font-medium text-gray-600">Actions</th>
                </tr>
              </thead>
              <tbody>
                {reports.map((report) => (
                  <tr key={report.id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <FileText className="w-5 h-5 text-gray-400" />
                        <span className="font-medium text-gray-900">{report.name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm">{report.type}</span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{report.dateRange}</td>
                    <td className="px-4 py-3">{getStatusBadge(report.status)}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{report.size}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{new Date(report.createdAt).toLocaleDateString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-2">
                        {report.status === 'ready' && (
                          <>
                            <button 
                              onClick={() => handleDownload(report)}
                              className="p-2 text-gray-400 hover:text-emerald-600 hover:bg-emerald-50 rounded-lg"
                              title="Download JSON"
                            >
                              <Download className="w-4 h-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Scheduled Reports */}
      <div className="bg-white rounded-xl border border-gray-200">
        <div className="p-4 border-b border-gray-200 flex items-center justify-between">
          <div>
            <h2 className="font-semibold text-gray-900">Scheduled Reports</h2>
            <p className="text-sm text-gray-500">Automatically generated and delivered reports</p>
          </div>
          <button className="flex items-center gap-2 px-3 py-1.5 text-sm text-emerald-600 border border-emerald-200 rounded-lg hover:bg-emerald-50">
            <Plus className="w-4 h-4" />
            Add Schedule
          </button>
        </div>
        <div className="p-4 space-y-4">
          {scheduledReports.map((report) => (
            <div key={report.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 bg-emerald-100 rounded-lg flex items-center justify-center">
                  <Clock className="w-5 h-5 text-emerald-600" />
                </div>
                <div>
                  <h4 className="font-medium text-gray-900">{report.name}</h4>
                  <p className="text-sm text-gray-500">{report.frequency} • Next: {report.nextRun}</p>
                </div>
              </div>
              <div className="flex items-center gap-4">
                <span className="px-2 py-1 bg-white border border-gray-200 rounded text-sm text-gray-600">
                  {report.format}
                </span>
                <div className="flex items-center gap-1">
                  <Mail className="w-4 h-4 text-gray-400" />
                  <span className="text-sm text-gray-500">{report.recipients.length} recipient(s)</span>
                </div>
                <button className="text-sm text-emerald-600 hover:text-emerald-700">Edit</button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Export Options */}
      <div className="bg-gradient-to-r from-emerald-50 to-teal-50 rounded-xl border border-emerald-200 p-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">Bulk Export</h3>
            <p className="text-sm text-gray-600 mt-1">Download all your data for external analysis or backup</p>
          </div>
          <div className="flex items-center gap-3">
            <button className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 text-sm font-medium">
              <Download className="w-4 h-4" />
              Export CSV
            </button>
            <button className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 text-sm font-medium">
              <Download className="w-4 h-4" />
              Export PDF
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
