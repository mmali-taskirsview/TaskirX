import React, { useState, useEffect } from 'react';

const PublisherDemo = () => {
    const [auctionStatus, setAuctionStatus] = useState('Idle'); // Idle, Bidding, Ended
    const [bids, setBids] = useState([]);
    const [winner, setWinner] = useState(null);
    const [logs, setLogs] = useState([]);

    const log = (msg) => setLogs(prev => [...prev, `[${new Date().toLocaleTimeString()}] ${msg}`]);

    const runAuction = async () => {
        setAuctionStatus('Bidding');
        setBids([]);
        setWinner(null);
        setLogs([]);
        log("Starting Header Bidding Auction...");

        // 1. Define Bidders
        const externalBidderPromise = new Promise((resolve) => {
            const latency = Math.random() * 500 + 200; // 200-700ms random latency
            setTimeout(() => {
                const bid = {
                    bidder: 'Competitor-X (Simulated)',
                    price: (Math.random() * 5 + 1).toFixed(2), // Random $1-$6
                    ad: '<div>Competitor Ad Content</div>'
                };
                log(`Competitor-X responded in ${Math.floor(latency)}ms`);
                resolve(bid);
            }, latency);
        });

        // 2. Call Our Go Bidding Engine
        // Note: For this to work, we need CORS enabled on the Go engine or a proxy
        const taskirxBidderPromise = new Promise(async (resolve) => {
            const start = performance.now();
            try {
                // Simulating a request payload
                const bidRequest = {
                    id: `req-${Date.now()}`,
                    app: { id: "demo-publisher-site" },
                    device: { type: "desktop", ip: "127.0.0.1" },
                    user: { 
                        id: "demo-user", 
                        country: "US", // Ensure this isn't blocked by your geo-rules!
                        categories: ["tech", "demo"] 
                    }
                };

                const response = await fetch('http://localhost:8080/bid', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(bidRequest)
                });

                const data = await response.json();
                const latency = performance.now() - start;
                log(`TaskirX Engine responded in ${Math.floor(latency)}ms`);

                if (response.ok && data.bid_price) {
                    resolve({
                        bidder: 'TaskirX Engine',
                        price: data.bid_price.toFixed(2),
                        ad: data.creative_html || '<div>TaskirX Ad</div>'
                    });
                } else {
                    log(`TaskirX No Bid: ${response.status}`);
                    resolve(null);
                }
            } catch (err) {
                log(`TaskirX Error: ${err.message}`);
                resolve(null);
            }
        });

        // 3. Race/Wait for all (simple implementation: wait all with timeout)
        // In real Prebid, we'd have a hard timeout.
        const TIMEOUT = 1000;
        
        const timeoutPromise = new Promise((_, reject) => 
            setTimeout(() => reject(new Error("Auction Timeout")), TIMEOUT)
        );

        try {
            const results = await Promise.allSettled([
                Promise.race([externalBidderPromise, timeoutPromise]),
                Promise.race([taskirxBidderPromise, timeoutPromise])
            ]);

            const validBids = results
                .filter(r => r.status === 'fulfilled' && r.value !== null)
                .map(r => r.value)
                .sort((a, b) => parseFloat(b.price) - parseFloat(a.price));

            setBids(validBids);
             setAuctionStatus('Ended');

            if (validBids.length > 0) {
                setWinner(validBids[0]);
                log(`Winner determined: ${validBids[0].bidder} at $${validBids[0].price}`);
            } else {
                log("Auction ended with no bids.");
            }

        } catch (error) {
            log(`Auction Critical Error: ${error.message}`);
            setAuctionStatus('Error');
        }
    };

    return (
        <div className="p-8 max-w-4xl mx-auto">
            <h1 className="text-3xl font-bold mb-6 text-gray-800">Publisher Header Bidding Demo</h1>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Control Panel */}
                <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
                    <h2 className="text-xl font-semibold mb-4">Auction Control</h2>
                    <p className="mb-4 text-gray-600">
                        Simulate a client-side header bidding auction. This browser will request bids from:
                        <ul className="list-disc ml-6 mt-2">
                            <li><strong>Competitor-X (Simulated)</strong>: Random $1-$6, Random Latency</li>
                            <li><strong>TaskirX (Real)</strong>: Calls local Go Engine (:8080)</li>
                        </ul>
                    </p>
                    <button 
                        onClick={runAuction}
                        disabled={auctionStatus === 'Bidding'}
                        className={`w-full py-3 px-6 rounded font-bold text-white transition-colors ${
                            auctionStatus === 'Bidding' 
                            ? 'bg-gray-400 cursor-not-allowed' 
                            : 'bg-blue-600 hover:bg-blue-700'
                        }`}
                    >
                        {auctionStatus === 'Bidding' ? 'Running Auction...' : 'Run Auction'}
                    </button>
                    
                    <div className="mt-6">
                        <h3 className="font-bold text-gray-700 mb-2">Live Logs:</h3>
                        <div className="bg-gray-900 text-green-400 p-4 rounded h-48 overflow-y-auto font-mono text-xs">
                            {logs.map((l, i) => <div key={i}>{l}</div>)}
                        </div>
                    </div>
                </div>

                {/* Results Panel */}
                <div className="bg-white p-6 rounded-lg shadow border border-gray-200 flex flex-col">
                    <h2 className="text-xl font-semibold mb-4">Auction Results</h2>
                    
                    {winner ? (
                        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded text-center">
                            <div className="text-sm text-green-800 font-bold uppercase tracking-wide">Winning Bid</div>
                            <div className="text-4xl font-bold text-green-600 my-2">${winner.price}</div>
                            <div className="text-gray-700">by {winner.bidder}</div>
                        </div>
                    ) : (
                        <div className="mb-6 p-12 bg-gray-50 border border-dashed border-gray-300 rounded text-center text-gray-400">
                            {auctionStatus === 'Ended' ? 'No Bids Received' : 'Waiting for Auction...'}
                        </div>
                    )}

                    <div className="flex-1">
                        <h3 className="font-bold text-gray-700 mb-2">Bid Landscape:</h3>
                        <div className="space-y-2">
                            {bids.map((bid, i) => (
                                <div key={i} className={`flex justify-between items-center p-3 rounded ${i === 0 ? 'bg-green-100 border border-green-300' : 'bg-gray-100'}`}>
                                    <span className="font-medium">{bid.bidder}</span>
                                    <span className="font-mono">${bid.price}</span>
                                </div>
                            ))}
                            {bids.length === 0 && auctionStatus === 'Ended' && (
                                <div className="text-sm text-gray-500 italic">No valid bids returned.</div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default PublisherDemo;
