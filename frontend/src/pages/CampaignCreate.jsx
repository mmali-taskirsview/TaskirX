import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { campaignsAPI } from '../services/api';
import {
  ArrowLeft,
  Save,
  Plus,
  X,
  DollarSign,
  Target,
  Globe,
  Smartphone,
  Calendar
} from 'lucide-react';
import toast from 'react-hot-toast';

const CampaignCreate = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(1);
  
  const [formData, setFormData] = useState({
    name: '',
    type: 'display',
    status: 'draft',
    bidding: {
      strategy: 'cpm',
      maxBid: 5.00
    },
    budget: {
      total: 1000,
      daily: 100
    },
    targeting: {
      geo: {
        countries: ['US'],
        cities: []
      },
      device: {
        types: ['mobile', 'desktop']
      }
    },
    schedule: {
      startDate: new Date().toISOString().split('T')[0],
      endDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
    },
    creative: {
      sizes: ['300x250'],
      assets: []
    }
  });

  const handleChange = (path, value) => {
    setFormData(prev => {
      const newData = { ...prev };
      const keys = path.split('.');
      let current = newData;
      
      for (let i = 0; i < keys.length - 1; i++) {
        current = current[keys[i]];
      }
      current[keys[keys.length - 1]] = value;
      
      return newData;
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      await campaignsAPI.create(formData);
      toast.success('Campaign created successfully!');
      navigate('/campaigns');
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to create campaign');
    } finally {
      setLoading(false);
    }
  };

  const countries = ['US', 'CA', 'UK', 'DE', 'FR', 'AU', 'JP', 'BR', 'IN', 'MX'];
  const adSizes = ['300x250', '728x90', '160x600', '320x50', '300x600', '970x250'];
  const deviceTypes = ['mobile', 'desktop', 'tablet', 'ctv'];

  return (
    <div className="max-w-4xl mx-auto animate-fadeIn">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate('/campaigns')}
          className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
        >
          <ArrowLeft size={20} className="text-gray-600" />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Create Campaign</h1>
          <p className="text-gray-500">Set up a new advertising campaign</p>
        </div>
      </div>

      {/* Progress Steps */}
      <div className="flex items-center justify-between mb-8">
        {['Basic Info', 'Budget & Bidding', 'Targeting', 'Creative'].map((label, index) => (
          <div key={index} className="flex items-center">
            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
              step > index + 1 ? 'bg-green-500 text-white' :
              step === index + 1 ? 'bg-cyber-blue text-white' :
              'bg-gray-200 text-gray-500'
            }`}>
              {step > index + 1 ? '✓' : index + 1}
            </div>
            <span className={`ml-2 text-sm ${step === index + 1 ? 'text-gray-900 font-medium' : 'text-gray-500'}`}>
              {label}
            </span>
            {index < 3 && <div className={`w-16 h-0.5 mx-4 ${step > index + 1 ? 'bg-green-500' : 'bg-gray-200'}`}></div>}
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit}>
        <div className="bg-white rounded-xl shadow-sm p-6">
          {/* Step 1: Basic Info */}
          {step === 1 && (
            <div className="space-y-6">
              <h2 className="text-lg font-semibold text-gray-900">Basic Information</h2>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Campaign Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleChange('name', e.target.value)}
                  placeholder="e.g., Summer Sale 2026"
                  required
                  className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Campaign Type</label>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                  {['display', 'video', 'native', 'audio'].map((type) => (
                    <button
                      key={type}
                      type="button"
                      onClick={() => handleChange('type', type)}
                      className={`p-4 border rounded-lg text-center transition-all ${
                        formData.type === type
                          ? 'border-cyber-blue bg-blue-50 text-cyber-blue'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <span className="capitalize font-medium">{type}</span>
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Schedule</label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">Start Date</label>
                    <input
                      type="date"
                      value={formData.schedule.startDate}
                      onChange={(e) => handleChange('schedule.startDate', e.target.value)}
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">End Date</label>
                    <input
                      type="date"
                      value={formData.schedule.endDate}
                      onChange={(e) => handleChange('schedule.endDate', e.target.value)}
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Step 2: Budget & Bidding */}
          {step === 2 && (
            <div className="space-y-6">
              <h2 className="text-lg font-semibold text-gray-900">Budget & Bidding</h2>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Bidding Strategy</label>
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { value: 'cpm', label: 'CPM', desc: 'Cost per 1,000 impressions' },
                    { value: 'cpc', label: 'CPC', desc: 'Cost per click' },
                    { value: 'cpa', label: 'CPA', desc: 'Cost per action' }
                  ].map((strategy) => (
                    <button
                      key={strategy.value}
                      type="button"
                      onClick={() => handleChange('bidding.strategy', strategy.value)}
                      className={`p-4 border rounded-lg text-left transition-all ${
                        formData.bidding.strategy === strategy.value
                          ? 'border-cyber-blue bg-blue-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <span className="font-semibold block">{strategy.label}</span>
                      <span className="text-xs text-gray-500">{strategy.desc}</span>
                    </button>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Max Bid ($)</label>
                  <div className="relative">
                    <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                    <input
                      type="number"
                      step="0.01"
                      min="0.01"
                      value={formData.bidding.maxBid}
                      onChange={(e) => handleChange('bidding.maxBid', parseFloat(e.target.value))}
                      className="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Total Budget ($)</label>
                  <div className="relative">
                    <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                    <input
                      type="number"
                      min="100"
                      value={formData.budget.total}
                      onChange={(e) => handleChange('budget.total', parseInt(e.target.value))}
                      className="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Daily Budget ($)</label>
                <div className="relative">
                  <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                  <input
                    type="number"
                    min="10"
                    value={formData.budget.daily}
                    onChange={(e) => handleChange('budget.daily', parseInt(e.target.value))}
                    className="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyber-blue"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Step 3: Targeting */}
          {step === 3 && (
            <div className="space-y-6">
              <h2 className="text-lg font-semibold text-gray-900">Targeting</h2>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <Globe size={14} className="inline mr-1" />
                  Countries
                </label>
                <div className="flex flex-wrap gap-2">
                  {countries.map((country) => (
                    <button
                      key={country}
                      type="button"
                      onClick={() => {
                        const current = formData.targeting.geo.countries;
                        const updated = current.includes(country)
                          ? current.filter(c => c !== country)
                          : [...current, country];
                        handleChange('targeting.geo.countries', updated);
                      }}
                      className={`px-3 py-1.5 border rounded-lg text-sm transition-all ${
                        formData.targeting.geo.countries.includes(country)
                          ? 'border-cyber-blue bg-blue-50 text-cyber-blue'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      {country}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <Smartphone size={14} className="inline mr-1" />
                  Devices
                </label>
                <div className="flex flex-wrap gap-2">
                  {deviceTypes.map((device) => (
                    <button
                      key={device}
                      type="button"
                      onClick={() => {
                        const current = formData.targeting.device.types;
                        const updated = current.includes(device)
                          ? current.filter(d => d !== device)
                          : [...current, device];
                        handleChange('targeting.device.types', updated);
                      }}
                      className={`px-3 py-1.5 border rounded-lg text-sm capitalize transition-all ${
                        formData.targeting.device.types.includes(device)
                          ? 'border-cyber-blue bg-blue-50 text-cyber-blue'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      {device}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Step 4: Creative */}
          {step === 4 && (
            <div className="space-y-6">
              <h2 className="text-lg font-semibold text-gray-900">Creative</h2>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Ad Sizes</label>
                <div className="flex flex-wrap gap-2">
                  {adSizes.map((size) => (
                    <button
                      key={size}
                      type="button"
                      onClick={() => {
                        const current = formData.creative.sizes;
                        const updated = current.includes(size)
                          ? current.filter(s => s !== size)
                          : [...current, size];
                        handleChange('creative.sizes', updated);
                      }}
                      className={`px-3 py-1.5 border rounded-lg text-sm transition-all ${
                        formData.creative.sizes.includes(size)
                          ? 'border-cyber-blue bg-blue-50 text-cyber-blue'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      {size}
                    </button>
                  ))}
                </div>
              </div>

              <div className="p-8 border-2 border-dashed border-gray-200 rounded-lg text-center">
                <div className="text-gray-400 mb-2">
                  <Plus size={32} className="mx-auto" />
                </div>
                <p className="text-gray-600 font-medium">Upload Creative Assets</p>
                <p className="text-sm text-gray-400">PNG, JPG, GIF up to 5MB</p>
              </div>

              {/* Summary */}
              <div className="bg-gray-50 rounded-lg p-4">
                <h4 className="font-medium text-gray-900 mb-3">Campaign Summary</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-500">Name:</span>
                    <span className="ml-2 font-medium">{formData.name || 'Not set'}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Type:</span>
                    <span className="ml-2 font-medium capitalize">{formData.type}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Budget:</span>
                    <span className="ml-2 font-medium">${formData.budget.total}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Bidding:</span>
                    <span className="ml-2 font-medium">{formData.bidding.strategy.toUpperCase()} @ ${formData.bidding.maxBid}</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Navigation Buttons */}
          <div className="flex justify-between mt-8 pt-6 border-t border-gray-100">
            <button
              type="button"
              onClick={() => setStep(Math.max(1, step - 1))}
              disabled={step === 1}
              className="px-6 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            
            {step < 4 ? (
              <button
                type="button"
                onClick={() => setStep(Math.min(4, step + 1))}
                className="px-6 py-2 bg-cyber-blue text-white rounded-lg hover:bg-blue-600 transition-colors"
              >
                Next
              </button>
            ) : (
              <button
                type="submit"
                disabled={loading || !formData.name}
                className="flex items-center gap-2 px-6 py-2 bg-gradient-to-r from-cyber-blue to-blue-600 text-white rounded-lg hover:from-blue-600 hover:to-cyber-blue transition-all disabled:opacity-50"
              >
                {loading ? (
                  <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                ) : (
                  <Save size={18} />
                )}
                Create Campaign
              </button>
            )}
          </div>
        </div>
      </form>
    </div>
  );
};

export default CampaignCreate;
