/**
 * TaskirX - JavaScript SDK
 * Version: 1.0.0
 * For web publishers - banner, native, and video ads
 */

(function(window) {
  'use strict';

  const AdxSDK = {
    version: '1.0.0',
    config: {
      apiUrl: 'http://localhost:3000/api',
      publisherId: null,
      debug: false
    },
    
    ads: new Map(), // Track active ads
    
    /**
     * Initialize the SDK
     * @param {Object} options - Configuration options
     * @param {string} options.publisherId - Publisher ID
     * @param {string} options.apiUrl - API base URL (optional)
     * @param {boolean} options.debug - Enable debug logging (optional)
     */
    init: function(options) {
      if (!options.publisherId) {
        throw new Error('AdxSDK: publisherId is required');
      }
      
      this.config.publisherId = options.publisherId;
      if (options.apiUrl) this.config.apiUrl = options.apiUrl;
      if (options.debug) this.config.debug = options.debug;
      
      this.log('AdxSDK initialized', this.config);
      
      // Setup viewport tracking for viewability
      this.setupViewabilityTracking();
    },
    
    /**
     * Request and display a banner ad
     * @param {Object} options - Banner options
     * @param {string} options.containerId - DOM element ID to insert ad
     * @param {string} options.size - Ad size (e.g., '300x250', '728x90')
     * @param {Function} options.onAdLoaded - Callback when ad loads
     * @param {Function} options.onAdFailed - Callback when ad fails
     */
    showBanner: async function(options) {
      const { containerId, size = '300x250', onAdLoaded, onAdFailed } = options;
      const container = document.getElementById(containerId);
      
      if (!container) {
        this.error('Container not found:', containerId);
        if (onAdFailed) onAdFailed(new Error('Container not found'));
        return;
      }
      
      try {
        const [width, height] = size.split('x').map(Number);
        
        // Create ad request
        const bidRequest = this.createBidRequest({
          imp: [{
            id: '1',
            banner: { w: width, h: height },
            bidfloor: 0.50
          }]
        });
        
        // Request ad from server
        const response = await fetch(`${this.config.apiUrl}/rtb/bid-request`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(bidRequest)
        });
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        
        const bidResponse = await response.json();
        
        if (bidResponse.seatbid && bidResponse.seatbid.length > 0) {
          const bid = bidResponse.seatbid[0].bid[0];
          
          // Render the ad
          this.renderBanner(container, bid, { width, height });
          
          // Track impression
          this.trackImpression(bid.id);
          
          // Store ad reference
          this.ads.set(containerId, { bid, type: 'banner' });
          
          if (onAdLoaded) onAdLoaded(bid);
          this.log('Banner ad loaded', bid);
        } else {
          throw new Error('No ads available');
        }
      } catch (error) {
        this.error('Failed to load banner ad:', error);
        if (onAdFailed) onAdFailed(error);
        
        // Show placeholder
        container.innerHTML = `<div style="width:${size.split('x')[0]}px;height:${size.split('x')[1]}px;background:#f0f0f0;display:flex;align-items:center;justify-content:center;color:#999;">Ad</div>`;
      }
    },
    
    /**
     * Request and display a native ad
     * @param {Object} options - Native ad options
     * @param {string} options.containerId - DOM element ID
     * @param {Function} options.onAdLoaded - Callback when ad loads
     * @param {Function} options.onAdFailed - Callback when ad fails
     */
    showNative: async function(options) {
      const { containerId, onAdLoaded, onAdFailed } = options;
      const container = document.getElementById(containerId);
      
      if (!container) {
        this.error('Container not found:', containerId);
        if (onAdFailed) onAdFailed(new Error('Container not found'));
        return;
      }
      
      try {
        const bidRequest = this.createBidRequest({
          imp: [{
            id: '1',
            native: {
              request: JSON.stringify({
                ver: '1.2',
                assets: [
                  { id: 1, required: 1, title: { len: 90 } },
                  { id: 2, required: 1, img: { type: 3, w: 1200, h: 627 } },
                  { id: 3, required: 1, data: { type: 2, len: 140 } }
                ]
              })
            },
            bidfloor: 0.50
          }]
        });
        
        const response = await fetch(`${this.config.apiUrl}/rtb/bid-request`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(bidRequest)
        });
        
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        
        const bidResponse = await response.json();
        
        if (bidResponse.seatbid && bidResponse.seatbid.length > 0) {
          const bid = bidResponse.seatbid[0].bid[0];
          
          // Render native ad
          this.renderNative(container, bid);
          
          // Track impression
          this.trackImpression(bid.id);
          
          this.ads.set(containerId, { bid, type: 'native' });
          
          if (onAdLoaded) onAdLoaded(bid);
          this.log('Native ad loaded', bid);
        } else {
          throw new Error('No ads available');
        }
      } catch (error) {
        this.error('Failed to load native ad:', error);
        if (onAdFailed) onAdFailed(error);
      }
    },
    
    /**
     * Request and display a video ad
     * @param {Object} options - Video ad options
     * @param {string} options.containerId - DOM element ID
     * @param {string} options.size - Video size (e.g., '640x480')
     * @param {Function} options.onAdLoaded - Callback when ad loads
     * @param {Function} options.onAdFailed - Callback when ad fails
     * @param {Function} options.onAdCompleted - Callback when video completes
     */
    showVideo: async function(options) {
      const { containerId, size = '640x480', onAdLoaded, onAdFailed, onAdCompleted } = options;
      const container = document.getElementById(containerId);
      
      if (!container) {
        this.error('Container not found:', containerId);
        if (onAdFailed) onAdFailed(new Error('Container not found'));
        return;
      }
      
      try {
        const [width, height] = size.split('x').map(Number);
        
        const bidRequest = this.createBidRequest({
          imp: [{
            id: '1',
            video: {
              w: width,
              h: height,
              mimes: ['video/mp4', 'video/webm'],
              protocols: [2, 3, 5, 6],
              minduration: 5,
              maxduration: 30
            },
            bidfloor: 1.00
          }]
        });
        
        const response = await fetch(`${this.config.apiUrl}/rtb/bid-request`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(bidRequest)
        });
        
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        
        const bidResponse = await response.json();
        
        if (bidResponse.seatbid && bidResponse.seatbid.length > 0) {
          const bid = bidResponse.seatbid[0].bid[0];
          
          // Render video ad
          this.renderVideo(container, bid, { width, height, onAdCompleted });
          
          // Track impression
          this.trackImpression(bid.id);
          
          this.ads.set(containerId, { bid, type: 'video' });
          
          if (onAdLoaded) onAdLoaded(bid);
          this.log('Video ad loaded', bid);
        } else {
          throw new Error('No ads available');
        }
      } catch (error) {
        this.error('Failed to load video ad:', error);
        if (onAdFailed) onAdFailed(error);
      }
    },
    
    /**
     * Render banner ad HTML
     */
    renderBanner: function(container, bid, dimensions) {
      const clickUrl = `${this.config.apiUrl}/rtb/click/${bid.id}`;
      
      container.innerHTML = `
        <div class="adx-banner" style="width:${dimensions.width}px;height:${dimensions.height}px;position:relative;overflow:hidden;">
          <a href="${clickUrl}" target="_blank" rel="noopener" onclick="AdxSDK.trackClick('${bid.id}')">
            <img src="${bid.adm}" alt="Advertisement" style="width:100%;height:100%;object-fit:cover;" />
          </a>
          <div style="position:absolute;bottom:2px;right:2px;background:rgba(0,0,0,0.5);color:white;font-size:10px;padding:2px 4px;border-radius:2px;">Ad</div>
        </div>
      `;
    },
    
    /**
     * Render native ad HTML
     */
    renderNative: function(container, bid) {
      const clickUrl = `${this.config.apiUrl}/rtb/click/${bid.id}`;
      
      // Parse native response (simplified)
      container.innerHTML = `
        <div class="adx-native" style="border:1px solid #e0e0e0;border-radius:8px;padding:16px;max-width:600px;">
          <a href="${clickUrl}" target="_blank" rel="noopener" onclick="AdxSDK.trackClick('${bid.id}')" style="text-decoration:none;color:inherit;">
            <img src="${bid.adm}" alt="Advertisement" style="width:100%;border-radius:4px;margin-bottom:12px;" />
            <h3 style="margin:0 0 8px 0;font-size:18px;color:#333;">Sponsored Content</h3>
            <p style="margin:0;font-size:14px;color:#666;">Click to learn more about this offer.</p>
          </a>
          <div style="margin-top:8px;font-size:10px;color:#999;">Advertisement</div>
        </div>
      `;
    },
    
    /**
     * Render video ad HTML
     */
    renderVideo: function(container, bid, options) {
      const { width, height, onAdCompleted } = options;
      
      container.innerHTML = `
        <div class="adx-video" style="width:${width}px;height:${height}px;position:relative;">
          <video id="adx-video-${bid.id}" width="${width}" height="${height}" controls autoplay>
            <source src="${bid.adm}" type="video/mp4">
            Your browser does not support video.
          </video>
          <div style="position:absolute;top:4px;right:4px;background:rgba(0,0,0,0.7);color:white;font-size:10px;padding:4px 6px;border-radius:2px;">Ad</div>
        </div>
      `;
      
      // Track video completion
      const video = document.getElementById(`adx-video-${bid.id}`);
      if (video) {
        video.addEventListener('ended', () => {
          this.log('Video ad completed', bid.id);
          if (onAdCompleted) onAdCompleted(bid);
        });
        
        video.addEventListener('click', () => {
          this.trackClick(bid.id);
        });
      }
    },
    
    /**
     * Create OpenRTB bid request
     */
    createBidRequest: function(options) {
      return {
        id: this.generateId(),
        imp: options.imp,
        site: {
          id: this.config.publisherId,
          domain: window.location.hostname,
          page: window.location.href,
          ref: document.referrer
        },
        device: {
          ua: navigator.userAgent,
          ip: '', // Server will detect
          devicetype: this.getDeviceType(),
          geo: {}
        },
        user: {
          id: this.getUserId()
        },
        at: 2, // Second price auction
        tmax: 120,
        cur: ['USD']
      };
    },
    
    /**
     * Track impression
     */
    trackImpression: function(bidId) {
      const pixel = new Image(1, 1);
      pixel.src = `${this.config.apiUrl}/rtb/impression/${bidId}`;
      this.log('Impression tracked', bidId);
    },
    
    /**
     * Track click
     */
    trackClick: function(bidId) {
      fetch(`${this.config.apiUrl}/rtb/click/${bidId}`);
      this.log('Click tracked', bidId);
    },
    
    /**
     * Setup viewability tracking
     */
    setupViewabilityTracking: function() {
      if ('IntersectionObserver' in window) {
        const observer = new IntersectionObserver((entries) => {
          entries.forEach(entry => {
            if (entry.isIntersecting && entry.intersectionRatio >= 0.5) {
              this.log('Ad viewable', entry.target.id);
              // Could send viewability event here
            }
          });
        }, { threshold: 0.5 });
        
        // Observe all ad containers
        document.addEventListener('DOMContentLoaded', () => {
          document.querySelectorAll('[id^="ad-"]').forEach(el => {
            observer.observe(el);
          });
        });
      }
    },
    
    /**
     * Helper: Get device type
     */
    getDeviceType: function() {
      const ua = navigator.userAgent;
      if (/Mobile|Android|iPhone|iPad|iPod/i.test(ua)) {
        return /iPad|Tablet/i.test(ua) ? 5 : 4; // 5=tablet, 4=phone
      }
      return 2; // 2=PC
    },
    
    /**
     * Helper: Get or create user ID
     */
    getUserId: function() {
      let userId = localStorage.getItem('adx_user_id');
      if (!userId) {
        userId = this.generateId();
        localStorage.setItem('adx_user_id', userId);
      }
      return userId;
    },
    
    /**
     * Helper: Generate unique ID
     */
    generateId: function() {
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
      });
    },
    
    /**
     * Helper: Debug logging
     */
    log: function(...args) {
      if (this.config.debug) {
        console.log('[AdxSDK]', ...args);
      }
    },
    
    /**
     * Helper: Error logging
     */
    error: function(...args) {
      console.error('[AdxSDK]', ...args);
    },
    
    /**
     * Destroy ad and cleanup
     */
    destroyAd: function(containerId) {
      const container = document.getElementById(containerId);
      if (container) {
        container.innerHTML = '';
      }
      this.ads.delete(containerId);
      this.log('Ad destroyed', containerId);
    }
  };
  
  // Expose SDK globally
  window.AdxSDK = AdxSDK;
  
})(window);
