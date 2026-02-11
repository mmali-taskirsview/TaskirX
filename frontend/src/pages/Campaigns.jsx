import { useState, useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { campaignsAPI } from '../services/api';
import {
  Plus,
  Search,
  Filter,
  MoreVertical,
  Play,
  Pause,
  Trash2,
  Edit,
  Eye,
  TrendingUp
} from 'lucide-react';
import toast from 'react-hot-toast';

const Campaigns = () => {
  const [searchParams] = useSearchParams();
  const [campaigns, setCampaigns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState(searchParams.get('search') || '');
  const [statusFilter, setStatusFilter] = useState('all');
  const [selectedCampaign, setSelectedCampaign] = useState(null);

  useEffect(() => {
    fetchCampaigns();
  }, [statusFilter]);

  const fetchCampaigns = async () => {
    try {
      const params = {};
      if (statusFilter !== 'all') params.status = statusFilter;
      
      const response = await campaignsAPI.getAll(params);
      setCampaigns(response.data?.campaigns || []);
    } catch (error) {
      console.error('Failed to fetch campaigns:', error);
      // Demo data
      setCampaigns([
        {
          _id: '1',
          name: 'Summer Sale Campaign',
          status: 'active',
          type: 'display',
          budget: { total: 10000, spent: 4500 },
          bidding: { strategy: 'cpm', maxBid: 5.0 },
          stats: { impressions: 125000, clicks: 4500, ctr: 3.6 },
          createdAt: new Date().toISOString()
        },
        {
          _id: '2',
          name: 'Brand Awareness Q1',
          status: 'active',
          type: 'video',
          budget: { total: 25000, spent: 12000 },
          bidding: { strategy: 'cpv', maxBid: 0.10 },
          stats: { impressions: 450000, clicks: 15000, ctr: 3.3 },
          createdAt: new Date().toISOString()
        },
        {
          _id: '3',
          name: 'Retargeting Campaign',
          status: 'paused',
          type: 'native',
          budget: { total: 5000, spent: 2800 },
          bidding: { strategy: 'cpc', maxBid: 1.50 },
          stats: { impressions: 85000, clicks: 3200, ctr: 3.8 },
          createdAt: new Date().toISOString()
        }
      ]);
    } finally {
      setLoading(false);
    }
  };

  const handlePause = async (id) => {
    try {
      await campaignsAPI.pause(id);
      toast.success('Campaign paused');
      fetchCampaigns();
    } catch (error) {
      toast.error('Failed to pause campaign');
    }
  };

  const handleResume = async (id) => {
    try {
      await campaignsAPI.resume(id);
      toast.success('Campaign resumed');
      fetchCampaigns();
    } catch (error) {
      toast.error('Failed to resume campaign');
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this campaign?')) return;
    
    try {
      await campaignsAPI.delete(id);
      toast.success('Campaign deleted');
      fetchCampaigns();
    } catch (error) {
      toast.error('Failed to delete campaign');
    }
  };

  const filteredCampaigns = campaigns.filter(campaign =>
    campaign.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusBadge = (status) => {
    const styles = {
      active: 'bg-green-100 text-green-800',
      paused: 'bg-yellow-100 text-yellow-800',
      completed: 'bg-gray-100 text-gray-800',
      draft: 'bg-blue-100 text-blue-800'
    };
    return styles[status] || styles.draft;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-cyber-blue"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campaigns</h1>
          <p className="text-gray-500">Manage your advertising campaigns</p>
        </div>
        <Link
          to="/campaigns/new"
          className="flex items-center gap-2 px-4 py-2 bg-cyber-blue text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          <Plus size={18} />
          New Campaign
        </Link>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          {/* Search */}
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
            <input
              type="text"
              placeholder="Search campaigns..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
            />
          </div>

          {/* Status Filter */}
          <div className="flex items-center gap-2">
            <Filter size={18} className="text-gray-400" />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="border border-gray-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-cyber-blue"
            >
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="paused">Paused</option>
              <option value="completed">Completed</option>
              <option value="draft">Draft</option>
            </select>
          </div>
        </div>
      </div>

      {/* Campaigns Table */}
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  Campaign
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  Budget
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  Impressions
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  CTR
                </th>
                <th className="px-6 py-4 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {filteredCampaigns.length > 0 ? (
                filteredCampaigns.map((campaign) => (
                  <tr key={campaign._id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4">
                      <div>
                        <p className="font-medium text-gray-900">{campaign.name}</p>
                        <p className="text-sm text-gray-500 capitalize">{campaign.type} • {campaign.bidding?.strategy?.toUpperCase()}</p>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 rounded-full text-xs font-semibold ${getStatusBadge(campaign.status)}`}>
                        {campaign.status}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div>
                        <p className="font-medium text-gray-900">${campaign.budget?.spent?.toLocaleString() || 0}</p>
                        <p className="text-sm text-gray-500">of ${campaign.budget?.total?.toLocaleString() || 0}</p>
                        <div className="w-24 h-1.5 bg-gray-200 rounded-full mt-1">
                          <div 
                            className="h-full bg-cyber-blue rounded-full"
                            style={{ width: `${(campaign.budget?.spent / campaign.budget?.total) * 100 || 0}%` }}
                          ></div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <Eye size={14} className="text-gray-400" />
                        <span className="font-medium text-gray-900">{campaign.stats?.impressions?.toLocaleString() || 0}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <TrendingUp size={14} className="text-green-500" />
                        <span className="font-medium text-gray-900">{campaign.stats?.ctr || 0}%</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <Link
                          to={`/campaigns/${campaign._id}`}
                          className="p-2 text-gray-400 hover:text-cyber-blue hover:bg-blue-50 rounded-lg transition-colors"
                          title="View"
                        >
                          <Eye size={16} />
                        </Link>
                        <Link
                          to={`/campaigns/${campaign._id}?edit=true`}
                          className="p-2 text-gray-400 hover:text-cyber-blue hover:bg-blue-50 rounded-lg transition-colors"
                          title="Edit"
                        >
                          <Edit size={16} />
                        </Link>
                        {campaign.status === 'active' ? (
                          <button
                            onClick={() => handlePause(campaign._id)}
                            className="p-2 text-gray-400 hover:text-yellow-600 hover:bg-yellow-50 rounded-lg transition-colors"
                            title="Pause"
                          >
                            <Pause size={16} />
                          </button>
                        ) : (
                          <button
                            onClick={() => handleResume(campaign._id)}
                            className="p-2 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                            title="Resume"
                          >
                            <Play size={16} />
                          </button>
                        )}
                        <button
                          onClick={() => handleDelete(campaign._id)}
                          className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                          title="Delete"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="6" className="px-6 py-12 text-center">
                    <p className="text-gray-500 mb-2">No campaigns found</p>
                    <Link to="/campaigns/new" className="text-cyber-blue hover:underline">
                      Create your first campaign
                    </Link>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Campaigns;
