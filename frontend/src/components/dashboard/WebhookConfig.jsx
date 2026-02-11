import React, { useState, useEffect } from 'react';
import { Plus, Edit2, Trash2, Copy, Send, Eye, EyeOff } from 'lucide-react';

const WebhookConfig = () => {
  const [webhooks, setWebhooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [showLogs, setShowLogs] = useState(null);
  const [logs, setLogs] = useState([]);
  const [testLoading, setTestLoading] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    events: [],
    active: true,
    retryPolicy: 'exponential',
    maxRetries: 3,
    timeout: 30
  });

  const eventTypes = [
    'campaign.created',
    'campaign.updated',
    'campaign.paused',
    'campaign.resumed',
    'campaign.deleted',
    'bid.submitted',
    'bid.accepted',
    'bid.rejected',
    'impression.tracked',
    'click.tracked',
    'conversion.tracked',
    'analytics.updated'
  ];

  useEffect(() => {
    fetchWebhooks();
  }, []);

  const fetchWebhooks = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/webhooks', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch webhooks');

      const data = await response.json();
      setWebhooks(data.webhooks || []);
      setError(null);
    } catch (err) {
      console.error('Error fetching webhooks:', err);
      setError(err.message);
      setWebhooks([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/webhooks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(formData)
      });

      if (!response.ok) throw new Error('Failed to create webhook');

      setShowModal(false);
      resetForm();
      fetchWebhooks();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`/api/webhooks/${editingId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(formData)
      });

      if (!response.ok) throw new Error('Failed to update webhook');

      setShowModal(false);
      resetForm();
      fetchWebhooks();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this webhook?')) return;

    try {
      const response = await fetch(`/api/webhooks/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to delete webhook');

      fetchWebhooks();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleTest = async (webhookId) => {
    try {
      setTestLoading(webhookId);
      const response = await fetch(`/api/webhooks/${webhookId}/test`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Test failed');

      alert('Test webhook sent successfully!');
    } catch (err) {
      setError(err.message);
    } finally {
      setTestLoading(null);
    }
  };

  const handleFetchLogs = async (webhookId) => {
    try {
      const response = await fetch(`/api/webhooks/${webhookId}/logs`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch logs');

      const data = await response.json();
      setLogs(data.logs || []);
      setShowLogs(webhookId);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleEdit = (webhook) => {
    setEditingId(webhook.id);
    setFormData({
      name: webhook.name,
      url: webhook.url,
      events: webhook.events,
      active: webhook.active,
      retryPolicy: webhook.retryPolicy,
      maxRetries: webhook.maxRetries,
      timeout: webhook.timeout
    });
    setShowModal(true);
  };

  const resetForm = () => {
    setFormData({
      name: '',
      url: '',
      events: [],
      active: true,
      retryPolicy: 'exponential',
      maxRetries: 3,
      timeout: 30
    });
    setEditingId(null);
  };

  const toggleEvent = (event) => {
    setFormData(prev => ({
      ...prev,
      events: prev.events.includes(event)
        ? prev.events.filter(e => e !== event)
        : [...prev.events, event]
    }));
  };

  const copyUrl = (url) => {
    navigator.clipboard.writeText(url);
    alert('Webhook URL copied to clipboard!');
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
        <h2 className="text-2xl font-bold text-gray-900">🔗 Webhook Configuration</h2>
        <button
          onClick={() => {
            resetForm();
            setShowModal(true);
          }}
          className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded flex items-center gap-2 transition-colors"
        >
          <Plus size={20} />
          Add Webhook
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
          {error}
        </div>
      )}

      {/* Webhooks List */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {webhooks.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            No webhooks configured. Create one to receive event notifications!
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {webhooks.map(webhook => (
              <div key={webhook.id} className="p-6 hover:bg-gray-50 transition-colors">
                <div className="flex justify-between items-start mb-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h3 className="text-lg font-bold text-gray-900">{webhook.name}</h3>
                      <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                        webhook.active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                      }`}>
                        {webhook.active ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                    <p className="text-gray-600 text-sm mt-1 font-mono break-all">{webhook.url}</p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleFetchLogs(webhook.id)}
                      className="text-blue-600 hover:text-blue-900 p-2"
                      title="View Logs"
                    >
                      <Eye size={18} />
                    </button>
                    <button
                      onClick={() => handleTest(webhook.id)}
                      disabled={testLoading === webhook.id}
                      className="text-purple-600 hover:text-purple-900 p-2 disabled:opacity-50"
                      title="Test Webhook"
                    >
                      <Send size={18} />
                    </button>
                    <button
                      onClick={() => copyUrl(webhook.url)}
                      className="text-gray-600 hover:text-gray-900 p-2"
                      title="Copy URL"
                    >
                      <Copy size={18} />
                    </button>
                    <button
                      onClick={() => handleEdit(webhook)}
                      className="text-blue-600 hover:text-blue-900 p-2"
                      title="Edit Webhook"
                    >
                      <Edit2 size={18} />
                    </button>
                    <button
                      onClick={() => handleDelete(webhook.id)}
                      className="text-red-600 hover:text-red-900 p-2"
                      title="Delete Webhook"
                    >
                      <Trash2 size={18} />
                    </button>
                  </div>
                </div>

                <div className="mb-4">
                  <p className="text-sm font-medium text-gray-700 mb-2">Subscribed Events:</p>
                  <div className="flex flex-wrap gap-2">
                    {webhook.events.map(event => (
                      <span key={event} className="inline-block px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                        {event}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <p className="text-gray-600">Retry Policy</p>
                    <p className="font-medium text-gray-900">{webhook.retryPolicy}</p>
                  </div>
                  <div>
                    <p className="text-gray-600">Max Retries</p>
                    <p className="font-medium text-gray-900">{webhook.maxRetries}</p>
                  </div>
                  <div>
                    <p className="text-gray-600">Timeout (s)</p>
                    <p className="font-medium text-gray-900">{webhook.timeout}</p>
                  </div>
                </div>

                {showLogs === webhook.id && (
                  <div className="mt-6 pt-6 border-t border-gray-200">
                    <div className="flex justify-between items-center mb-4">
                      <h4 className="font-medium text-gray-900">Delivery Logs</h4>
                      <button
                        onClick={() => setShowLogs(null)}
                        className="text-gray-600 hover:text-gray-900"
                      >
                        <EyeOff size={18} />
                      </button>
                    </div>
                    <div className="space-y-2 max-h-60 overflow-y-auto">
                      {logs.length === 0 ? (
                        <p className="text-gray-500 text-sm">No logs available</p>
                      ) : (
                        logs.map((log, idx) => (
                          <div key={idx} className="text-sm p-3 bg-gray-50 rounded">
                            <div className="flex justify-between items-start">
                              <span className={`font-medium ${
                                log.status === 'success' ? 'text-green-600' : 'text-red-600'
                              }`}>
                                {log.status === 'success' ? '✓' : '✗'} {log.event}
                              </span>
                              <span className="text-gray-500 text-xs">
                                {new Date(log.timestamp).toLocaleString()}
                              </span>
                            </div>
                            {log.error && (
                              <p className="text-red-600 text-xs mt-1">{log.error}</p>
                            )}
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-200">
              <h3 className="text-xl font-bold text-gray-900">
                {editingId ? 'Edit Webhook' : 'Create New Webhook'}
              </h3>
            </div>

            <form onSubmit={editingId ? handleUpdate : handleCreate} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Webhook Name *</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({...formData, name: e.target.value})}
                  className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g., Campaign Events"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Webhook URL *</label>
                <input
                  type="url"
                  required
                  value={formData.url}
                  onChange={(e) => setFormData({...formData, url: e.target.value})}
                  className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://example.com/webhook"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-3">Subscribe to Events *</label>
                <div className="grid grid-cols-2 gap-3 max-h-48 overflow-y-auto p-3 border border-gray-300 rounded">
                  {eventTypes.map(event => (
                    <label key={event} className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={formData.events.includes(event)}
                        onChange={() => toggleEvent(event)}
                        className="w-4 h-4 rounded"
                      />
                      <span className="text-sm text-gray-700">{event}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Retry Policy</label>
                  <select
                    value={formData.retryPolicy}
                    onChange={(e) => setFormData({...formData, retryPolicy: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="exponential">Exponential</option>
                    <option value="linear">Linear</option>
                    <option value="constant">Constant</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Max Retries</label>
                  <input
                    type="number"
                    min="0"
                    max="10"
                    value={formData.maxRetries}
                    onChange={(e) => setFormData({...formData, maxRetries: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Timeout (seconds)</label>
                  <input
                    type="number"
                    min="5"
                    max="300"
                    value={formData.timeout}
                    onChange={(e) => setFormData({...formData, timeout: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={formData.active}
                  onChange={(e) => setFormData({...formData, active: e.target.checked})}
                  className="w-4 h-4 rounded"
                />
                <label className="text-sm font-medium text-gray-700">Active</label>
              </div>

              <div className="flex justify-end gap-3 pt-6 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    resetForm();
                  }}
                  className="px-4 py-2 text-gray-700 border border-gray-300 rounded hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                >
                  {editingId ? 'Update Webhook' : 'Create Webhook'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default WebhookConfig;
