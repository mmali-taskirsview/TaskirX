import { useState, useEffect, useRef } from 'react';
import { Activity, Zap, Clock, Server, TrendingUp, AlertCircle } from 'lucide-react';
import io from 'socket.io-client';

const RTBMonitor = () => {
  const [stats, setStats] = useState({
    qps: 0,
    avgLatency: 0,
    bidRate: 0,
    winRate: 0,
    totalBids: 0,
    activeBidders: 0
  });
  const [recentBids, setRecentBids] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);
  const socketRef = useRef(null);

  useEffect(() => {
    // Connect to NestJS Gateway
    socketRef.current = io('http://localhost:3000/metrics', {
      transports: ['websocket'],
      auth: {
        userId: 'admin-dashboard' // In prod, this would be a JWT
      }
    });

    socketRef.current.on('connect', () => {
      console.log('Connected to RTB Metrics Gateway');
      setIsConnected(true);
      setLoading(false);
      
      // Subscribe to real-time streams
      socketRef.current.emit('subscribe_bidding_metrics');
    });

    socketRef.current.on('disconnect', () => {
      console.log('Disconnected from RTB Metrics Gateway');
      setIsConnected(false);
    });

    // Listen for aggregate metrics
    socketRef.current.on('metrics_update', (data) => {
        if (!data) return;
        setStats(prev => ({
            qps: data.qps || prev.qps,
            avgLatency: data.latency || prev.avgLatency,
            bidRate: data.filled_rate || prev.bidRate,
            winRate: data.win_rate || prev.winRate,
            totalBids: data.total_bids || prev.totalBids,
            activeBidders: data.active_nodes || prev.activeBidders
        }));
    });

    // Listen for individual bid events (sampled)
    socketRef.current.on('bid_event', (event) => {
        setRecentBids(prev => {
            const newBid = {
                id: event.id || Date.now(),
                bidder: event.bidder || 'Unknown',
                amount: parseFloat(event.amount || 0).toFixed(2),
                latency: event.latency || 0,
                status: event.success ? 'won' : 'lost',
                timestamp: new Date().toISOString()
            };
            return [newBid, ...prev.slice(0, 19)];
        });
    });

    // Fallback/Demo mode if connection fails or no data for 2s
    // (Optional: keep simulation if backend silent for demo purposes)

    return () => {
      if (socketRef.current) {
        socketRef.current.disconnect();
      }
    };
  }, []);

  const statCards = [
    { title: 'Queries/Second', value: stats.qps.toLocaleString(), icon: Zap, color: 'bg-blue-500' },
    { title: 'Avg Latency', value: `${stats.avgLatency}ms`, icon: Clock, color: 'bg-green-500' },
    { title: 'Bid Rate', value: `${stats.bidRate}%`, icon: TrendingUp, color: 'bg-purple-500' },
    { title: 'Win Rate', value: `${stats.winRate}%`, icon: Activity, color: 'bg-orange-500' },
    { title: 'Total Bids Today', value: stats.totalBids.toLocaleString(), icon: Server, color: 'bg-pink-500' },
    { title: 'Active Bidders', value: stats.activeBidders, icon: Activity, color: 'bg-teal-500' }
  ];

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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">RTB Monitor</h1>
          <p className="text-gray-500">Real-time bidding activity</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 bg-green-100 text-green-800 rounded-full">
          <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
          <span className="text-sm font-medium">Live</span>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {statCards.map((card, index) => {
          const Icon = card.icon;
          return (
            <div key={index} className="bg-white rounded-xl p-4 shadow-sm">
              <div className={`${card.color} w-10 h-10 rounded-lg flex items-center justify-center mb-3`}>
                <Icon size={20} className="text-white" />
              </div>
              <p className="text-2xl font-bold text-gray-900">{card.value}</p>
              <p className="text-xs text-gray-500">{card.title}</p>
            </div>
          );
        })}
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Live Bid Stream */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm">
          <div className="p-4 border-b border-gray-100 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-gray-900">Live Bid Stream</h3>
            <span className="text-sm text-gray-500">Last 20 bids</span>
          </div>
          <div className="p-4 h-96 overflow-y-auto">
            <div className="space-y-2">
              {recentBids.map((bid) => (
                <div 
                  key={bid.id} 
                  className={`flex items-center justify-between p-3 rounded-lg transition-all ${
                    bid.status === 'won' ? 'bg-green-50' : 'bg-gray-50'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className={`w-2 h-2 rounded-full ${bid.status === 'won' ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                    <div>
                      <p className="font-medium text-gray-900">{bid.bidder}</p>
                      <p className="text-xs text-gray-500">
                        {new Date(bid.timestamp).toLocaleTimeString()}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-semibold text-gray-900">${bid.amount}</p>
                    <p className="text-xs text-gray-500">{bid.latency}ms</p>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs font-semibold ${
                    bid.status === 'won' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {bid.status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* System Health */}
        <div className="space-y-6">
          {/* Latency Distribution */}
          <div className="bg-white rounded-xl shadow-sm p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Latency Distribution</h3>
            <div className="space-y-3">
              {[
                { range: '0-50ms', percentage: 65, color: 'bg-green-500' },
                { range: '50-100ms', percentage: 25, color: 'bg-yellow-500' },
                { range: '100-150ms', percentage: 8, color: 'bg-orange-500' },
                { range: '>150ms', percentage: 2, color: 'bg-red-500' }
              ].map((item, index) => (
                <div key={index}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-gray-600">{item.range}</span>
                    <span className="font-medium">{item.percentage}%</span>
                  </div>
                  <div className="h-2 bg-gray-200 rounded-full">
                    <div 
                      className={`h-full ${item.color} rounded-full transition-all`}
                      style={{ width: `${item.percentage}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Alerts */}
          <div className="bg-white rounded-xl shadow-sm p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">System Alerts</h3>
            <div className="space-y-3">
              <div className="flex items-start gap-3 p-3 bg-yellow-50 rounded-lg">
                <AlertCircle size={18} className="text-yellow-600 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-yellow-800">High Latency Warning</p>
                  <p className="text-xs text-yellow-600">Bidder-5 avg latency: 120ms</p>
                </div>
              </div>
              <div className="flex items-start gap-3 p-3 bg-green-50 rounded-lg">
                <Activity size={18} className="text-green-600 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-green-800">All Systems Operational</p>
                  <p className="text-xs text-green-600">Last check: 30s ago</p>
                </div>
              </div>
            </div>
          </div>

          {/* Top Bidders */}
          <div className="bg-white rounded-xl shadow-sm p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Top Bidders</h3>
            <div className="space-y-3">
              {[
                { name: 'Bidder-3', bids: 12500, winRate: 32 },
                { name: 'Bidder-7', bids: 10200, winRate: 28 },
                { name: 'Bidder-1', bids: 8900, winRate: 35 }
              ].map((bidder, index) => (
                <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div>
                    <p className="font-medium text-gray-900">{bidder.name}</p>
                    <p className="text-xs text-gray-500">{bidder.bids.toLocaleString()} bids</p>
                  </div>
                  <span className="text-sm font-semibold text-cyber-blue">{bidder.winRate}% win</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RTBMonitor;
