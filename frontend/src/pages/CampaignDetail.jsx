import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { campaignsAPI } from '../services/api';
import {
  ArrowLeft,
  Edit,
  Trash2,
  Play,
  Pause,
  Eye,
  MousePointer,
  DollarSign,
  TrendingUp,
  Calendar,
  Target,
  Globe,
  Smartphone
} from 'lucide-react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import toast from 'react-hot-toast';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const CampaignDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [campaign, setCampaign] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCampaign();
  }, [id]);

  const fetchCampaign = async () => {
    try {
      const response = await campaignsAPI.getById(id);
      setCampaign(response.data);
    } catch (error) {
      console.error('Failed to fetch campaign:', error);
      // Demo data
      setCampaign({
        _id: id,
        name: 'Summer Sale Campaign',
        status: 'active',
        type: 'display',
        budget: { total: 10000, spent: 4500, daily: 500 },
        bidding: { strategy: 'cpm', maxBid: 5.0 },
        targeting: {
          geo: { countries: ['US', 'CA', 'UK'], cities: ['New York', 'Los Angeles'] },
          device: { types: ['mobile', 'desktop'] },
          demographics: { ageRange: '18-45' }
        },
        schedule: {
          startDate: '2026-01-01',
          endDate: '2026-02-28'
        },
        stats: {
          impressions: 125000,
          clicks: 4500,
          conversions: 320,
          ctr: 3.6,
          cpc: 1.00,
          cpm: 36.00,
          spend: 4500
        },
        createdAt: new Date().toISOString()
      });
    } finally {
      setLoading(false);
    }
  };

  const handlePause = async () => {
    try {
      await campaignsAPI.pause(id);
      toast.success('Campaign paused');
      fetchCampaign();
    } catch (error) {
      toast.error('Failed to pause campaign');
    }
  };

  const handleResume = async () => {
    try {
      await campaignsAPI.resume(id);
      toast.success('Campaign resumed');
      fetchCampaign();
    } catch (error) {
      toast.error('Failed to resume campaign');
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this campaign?')) return;
    
    try {
      await campaignsAPI.delete(id);
      toast.success('Campaign deleted');
      navigate('/campaigns');
    } catch (error) {
      toast.error('Failed to delete campaign');
    }
  };

  // Performance chart data
  const chartData = {
    labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
    datasets: [
      {
        label: 'Impressions',
        data: [25000, 35000, 40000, 25000],
        borderColor: '#0066FF',
        backgroundColor: 'rgba(0, 102, 255, 0.1)',
        fill: true,
        tension: 0.4,
        yAxisID: 'y',
      },
      {
        label: 'Clicks',
        data: [900, 1300, 1500, 800],
        borderColor: '#00FF00',
        backgroundColor: 'rgba(0, 255, 0, 0.1)',
        fill: true,
        tension: 0.4,
        yAxisID: 'y1',
      },
    ],
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index',
      intersect: false,
    },
    plugins: {
      legend: {
        position: 'top',
      },
    },
    scales: {
      y: {
        type: 'linear',
        display: true,
        position: 'left',
        grid: {
          color: 'rgba(0, 0, 0, 0.05)',
        },
      },
      y1: {
        type: 'linear',
        display: true,
        position: 'right',
        grid: {
          drawOnChartArea: false,
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

  if (!campaign) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">Campaign not found</p>
        <Link to="/campaigns" className="text-cyber-blue hover:underline">
          Back to campaigns
        </Link>
      </div>
    );
  }

  const getStatusBadge = (status) => {
    const styles = {
      active: 'bg-green-100 text-green-800',
      paused: 'bg-yellow-100 text-yellow-800',
      completed: 'bg-gray-100 text-gray-800',
      draft: 'bg-blue-100 text-blue-800'
    };
    return styles[status] || styles.draft;
  };

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link
            to="/campaigns"
            className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
          >
            <ArrowLeft size={20} className="text-gray-600" />
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-gray-900">{campaign.name}</h1>
              <span className={`px-2 py-1 rounded-full text-xs font-semibold ${getStatusBadge(campaign.status)}`}>
                {campaign.status}
              </span>
            </div>
            <p className="text-gray-500 capitalize">{campaign.type} Campaign • {campaign.bidding?.strategy?.toUpperCase()}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {campaign.status === 'active' ? (
            <button
              onClick={handlePause}
              className="flex items-center gap-2 px-4 py-2 border border-yellow-500 text-yellow-600 rounded-lg hover:bg-yellow-50 transition-colors"
            >
              <Pause size={16} />
              Pause
            </button>
          ) : (
            <button
              onClick={handleResume}
              className="flex items-center gap-2 px-4 py-2 border border-green-500 text-green-600 rounded-lg hover:bg-green-50 transition-colors"
            >
              <Play size={16} />
              Resume
            </button>
          )}
          <button className="flex items-center gap-2 px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
            <Edit size={16} />
            Edit
          </button>
          <button
            onClick={handleDelete}
            className="flex items-center gap-2 px-4 py-2 border border-red-200 text-red-600 rounded-lg hover:bg-red-50 transition-colors"
          >
            <Trash2 size={16} />
            Delete
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        <div className="bg-white rounded-xl p-4 shadow-sm">
          <div className="flex items-center gap-2 text-gray-500 mb-1">
            <Eye size={14} />
            <span className="text-xs">Impressions</span>
          </div>
          <p className="text-xl font-bold text-gray-900">{campaign.stats?.impressions?.toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-xl p-4 shadow-sm">
          <div className="flex items-center gap-2 text-gray-500 mb-1">
            <MousePointer size={14} />
            <span className="text-xs">Clicks</span>
          </div>
          <p className="text-xl font-bold text-gray-900">{campaign.stats?.clicks?.toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-xl p-4 shadow-sm">
          <div className="flex items-center gap-2 text-gray-500 mb-1">
            <TrendingUp size={14} />
            <span className="text-xs">CTR</span>
          </div>
          <p className="text-xl font-bold text-gray-900">{campaign.stats?.ctr}%</p>
        </div>
        <div className="bg-white rounded-xl p-4 shadow-sm">
          <div className="flex items-center gap-2 text-gray-500 mb-1">
            <DollarSign size={14} />
            <span className="text-xs">CPC</span>
          </div>
          <p className="text-xl font-bold text-gray-900">${campaign.stats?.cpc?.toFixed(2)}</p>
        </div>
        <div className="bg-white rounded-xl p-4 shadow-sm">
          <div className="flex items-center gap-2 text-gray-500 mb-1">
            <DollarSign size={14} />
            <span className="text-xs">CPM</span>
          </div>
          <p className="text-xl font-bold text-gray-900">${campaign.stats?.cpm?.toFixed(2)}</p>
        </div>
        <div className="bg-white rounded-xl p-4 shadow-sm">
          <div className="flex items-center gap-2 text-gray-500 mb-1">
            <Target size={14} />
            <span className="text-xs">Conversions</span>
          </div>
          <p className="text-xl font-bold text-gray-900">{campaign.stats?.conversions?.toLocaleString()}</p>
        </div>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Performance Chart */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Performance Over Time</h3>
          <div className="h-80">
            <Line data={chartData} options={chartOptions} />
          </div>
        </div>

        {/* Campaign Details */}
        <div className="space-y-6">
          {/* Budget */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Budget</h3>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-gray-500">Spent</span>
                  <span className="font-medium">${campaign.budget?.spent?.toLocaleString()} / ${campaign.budget?.total?.toLocaleString()}</span>
                </div>
                <div className="h-2 bg-gray-200 rounded-full">
                  <div 
                    className="h-full bg-cyber-blue rounded-full"
                    style={{ width: `${(campaign.budget?.spent / campaign.budget?.total) * 100}%` }}
                  ></div>
                </div>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Daily Budget</span>
                <span className="font-medium">${campaign.budget?.daily?.toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Max Bid</span>
                <span className="font-medium">${campaign.bidding?.maxBid?.toFixed(2)}</span>
              </div>
            </div>
          </div>

          {/* Targeting */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Targeting</h3>
            <div className="space-y-4">
              <div>
                <div className="flex items-center gap-2 text-gray-500 text-sm mb-2">
                  <Globe size={14} />
                  <span>Geo</span>
                </div>
                <div className="flex flex-wrap gap-1">
                  {campaign.targeting?.geo?.countries?.map((country) => (
                    <span key={country} className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs">
                      {country}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <div className="flex items-center gap-2 text-gray-500 text-sm mb-2">
                  <Smartphone size={14} />
                  <span>Devices</span>
                </div>
                <div className="flex flex-wrap gap-1">
                  {campaign.targeting?.device?.types?.map((device) => (
                    <span key={device} className="px-2 py-1 bg-green-50 text-green-700 rounded text-xs capitalize">
                      {device}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <div className="flex items-center gap-2 text-gray-500 text-sm mb-2">
                  <Calendar size={14} />
                  <span>Schedule</span>
                </div>
                <p className="text-sm">
                  {campaign.schedule?.startDate} → {campaign.schedule?.endDate}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CampaignDetail;
