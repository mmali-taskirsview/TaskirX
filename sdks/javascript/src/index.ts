/**
 * TaskirX JavaScript SDK v1.0.0
 * Modern TypeScript SDK for web advertising
 * 
 * Features:
 * - ES2022+ syntax
 * - TypeScript for type safety
 * - Intersection Observer for viewability
 * - Fetch API for network requests
 * - Web Components for ad rendering
 * - Automatic impression/click tracking
 */

export interface AdxConfig {
  publisherId: string;
  apiEndpoint?: string;
  enableDebug?: boolean;
  autoTrack?: boolean;
}

export interface AdPlacement {
  placementId: string;
  format: 'banner' | 'native' | 'video' | 'interstitial';
  size?: { width: number; height: number };
  targeting?: Record<string, any>;
}

export interface AdResponse {
  id: string;
  creative: {
    type: string;
    url: string;
    width: number;
    height: number;
    clickUrl: string;
  };
  impressionUrl: string;
  clickUrl: string;
}

class AdxSDK {
  private config: Required<AdxConfig>;
  private initialized = false;
  private observedAds = new Map<string, IntersectionObserver>();

  constructor() {
    this.config = {
      publisherId: '',
      apiEndpoint: 'http://localhost:3000/api',
      enableDebug: false,
      autoTrack: true
    };
  }

  /**
   * Initialize the SDK
   */
  public init(config: AdxConfig): void {
    if (this.initialized) {
      this.log('warn', 'SDK already initialized');
      return;
    }

    this.config = { ...this.config, ...config };
    this.initialized = true;
    this.log('info', `ADX SDK initialized for publisher: ${this.config.publisherId}`);
  }

  /**
   * Request and display a banner ad
   */
  public async showBanner(options: {
    containerId: string;
    placementId: string;
    size?: { width: number; height: number };
  }): Promise<void> {
    this.ensureInitialized();

    const container = document.getElementById(options.containerId);
    if (!container) {
      throw new Error(`Container element '${options.containerId}' not found`);
    }

    try {
      const ad = await this.requestAd({
        placementId: options.placementId,
        format: 'banner',
        size: options.size
      });

      this.renderBanner(container, ad);
      
      if (this.config.autoTrack) {
        this.setupViewabilityTracking(container, ad);
      }
    } catch (error) {
      this.log('error', 'Failed to show banner', error);
      throw error;
    }
  }

  /**
   * Request and display a native ad
   */
  public async showNative(options: {
    containerId: string;
    placementId: string;
    template?: (ad: AdResponse) => string;
  }): Promise<void> {
    this.ensureInitialized();

    const container = document.getElementById(options.containerId);
    if (!container) {
      throw new Error(`Container element '${options.containerId}' not found`);
    }

    try {
      const ad = await this.requestAd({
        placementId: options.placementId,
        format: 'native'
      });

      this.renderNative(container, ad, options.template);
      
      if (this.config.autoTrack) {
        this.setupViewabilityTracking(container, ad);
      }
    } catch (error) {
      this.log('error', 'Failed to show native ad', error);
      throw error;
    }
  }

  /**
   * Request and display a video ad
   */
  public async showVideo(options: {
    containerId: string;
    placementId: string;
    autoplay?: boolean;
    muted?: boolean;
  }): Promise<void> {
    this.ensureInitialized();

    const container = document.getElementById(options.containerId);
    if (!container) {
      throw new Error(`Container element '${options.containerId}' not found`);
    }

    try {
      const ad = await this.requestAd({
        placementId: options.placementId,
        format: 'video'
      });

      this.renderVideo(container, ad, options);
      
      if (this.config.autoTrack) {
        this.setupViewabilityTracking(container, ad);
      }
    } catch (error) {
      this.log('error', 'Failed to show video ad', error);
      throw error;
    }
  }

  /**
   * Request an ad from the server
   */
  private async requestAd(placement: AdPlacement): Promise<AdResponse> {
    const response = await fetch(`${this.config.apiEndpoint}/rtb/bid-request`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: this.generateId(),
        imp: [{
          id: '1',
          [placement.format === 'banner' ? 'banner' : placement.format]: {
            w: placement.size?.width || 300,
            h: placement.size?.height || 250
          }
        }],
        site: {
          id: this.config.publisherId,
          domain: window.location.hostname,
          page: window.location.href
        },
        device: {
          ua: navigator.userAgent,
          language: navigator.language,
          devicetype: this.getDeviceType()
        },
        user: {
          id: this.getUserId()
        }
      })
    });

    if (!response.ok) {
      throw new Error(`Ad request failed: ${response.statusText}`);
    }

    const data = await response.json();
    
    // Transform OpenRTB response to our format
    return this.transformBidResponse(data);
  }

  /**
   * Render banner ad
   */
  private renderBanner(container: HTMLElement, ad: AdResponse): void {
    const adElement = document.createElement('div');
    adElement.className = 'adx-banner-ad';
    adElement.style.cssText = `
      width: ${ad.creative.width}px;
      height: ${ad.creative.height}px;
      overflow: hidden;
      position: relative;
    `;

    const link = document.createElement('a');
    link.href = '#';
    link.onclick = (e) => {
      e.preventDefault();
      this.trackClick(ad.clickUrl);
      window.open(ad.creative.clickUrl, '_blank');
    };

    const img = document.createElement('img');
    img.src = ad.creative.url;
    img.style.cssText = 'width: 100%; height: 100%; object-fit: contain;';
    img.alt = 'Advertisement';

    link.appendChild(img);
    adElement.appendChild(link);
    container.appendChild(adElement);
  }

  /**
   * Render native ad
   */
  private renderNative(
    container: HTMLElement, 
    ad: AdResponse, 
    template?: (ad: AdResponse) => string
  ): void {
    const html = template ? template(ad) : this.getDefaultNativeTemplate(ad);
    
    const adElement = document.createElement('div');
    adElement.className = 'adx-native-ad';
    adElement.innerHTML = html;

    // Add click handler
    const clickableElements = adElement.querySelectorAll('[data-adx-click]');
    clickableElements.forEach(el => {
      el.addEventListener('click', (e) => {
        e.preventDefault();
        this.trackClick(ad.clickUrl);
        window.open(ad.creative.clickUrl, '_blank');
      });
    });

    container.appendChild(adElement);
  }

  /**
   * Render video ad
   */
  private renderVideo(
    container: HTMLElement,
    ad: AdResponse,
    options: { autoplay?: boolean; muted?: boolean }
  ): void {
    const videoElement = document.createElement('video');
    videoElement.className = 'adx-video-ad';
    videoElement.style.cssText = 'width: 100%; height: 100%;';
    videoElement.src = ad.creative.url;
    videoElement.controls = true;
    videoElement.autoplay = options.autoplay ?? false;
    videoElement.muted = options.muted ?? true;

    // Track video events
    videoElement.addEventListener('play', () => this.log('info', 'Video ad started'));
    videoElement.addEventListener('ended', () => this.log('info', 'Video ad completed'));

    // Click handler
    videoElement.addEventListener('click', () => {
      this.trackClick(ad.clickUrl);
      window.open(ad.creative.clickUrl, '_blank');
    });

    container.appendChild(videoElement);
  }

  /**
   * Setup viewability tracking using Intersection Observer
   */
  private setupViewabilityTracking(container: HTMLElement, ad: AdResponse): void {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting && entry.intersectionRatio >= 0.5) {
            this.trackImpression(ad.impressionUrl);
            observer.disconnect();
            this.observedAds.delete(ad.id);
          }
        });
      },
      {
        threshold: 0.5 // 50% viewability
      }
    );

    observer.observe(container);
    this.observedAds.set(ad.id, observer);
  }

  /**
   * Track impression
   */
  private trackImpression(impressionUrl: string): void {
    const img = new Image(1, 1);
    img.src = impressionUrl;
    this.log('info', 'Impression tracked');
  }

  /**
   * Track click
   */
  private trackClick(clickUrl: string): void {
    fetch(clickUrl, { method: 'GET', mode: 'no-cors' });
    this.log('info', 'Click tracked');
  }

  /**
   * Helper methods
   */
  private ensureInitialized(): void {
    if (!this.initialized) {
      throw new Error('SDK not initialized. Call init() first.');
    }
  }

  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private getUserId(): string {
    let userId = localStorage.getItem('adx_user_id');
    if (!userId) {
      userId = this.generateId();
      localStorage.setItem('adx_user_id', userId);
    }
    return userId;
  }

  private getDeviceType(): number {
    const ua = navigator.userAgent;
    if (/mobile/i.test(ua)) return 1; // Mobile
    if (/tablet/i.test(ua)) return 5; // Tablet
    return 2; // Desktop
  }

  private transformBidResponse(data: any): AdResponse {
    // Transform OpenRTB bid response to our format
    const seatbid = data.seatbid?.[0];
    const bid = seatbid?.bid?.[0];

    if (!bid) {
      throw new Error('No bid returned');
    }

    return {
      id: bid.id || this.generateId(),
      creative: {
        type: bid.creative?.type || 'banner',
        url: bid.creative?.url || '',
        width: bid.creative?.width || 300,
        height: bid.creative?.height || 250,
        clickUrl: bid.creative?.clickUrl || ''
      },
      impressionUrl: `${this.config.apiEndpoint}/rtb/impression/${bid.id}`,
      clickUrl: `${this.config.apiEndpoint}/rtb/click/${bid.id}`
    };
  }

  private getDefaultNativeTemplate(ad: AdResponse): string {
    return `
      <div class="adx-native-container" style="padding: 16px; border: 1px solid #e0e0e0; border-radius: 8px;">
        <div class="adx-native-image" style="margin-bottom: 12px;">
          <img src="${ad.creative.url}" alt="Ad" style="width: 100%; border-radius: 4px;" data-adx-click />
        </div>
        <div class="adx-native-sponsored" style="font-size: 12px; color: #999; margin-bottom: 4px;">
          Sponsored
        </div>
      </div>
    `;
  }

  private log(level: 'info' | 'warn' | 'error', message: string, ...args: any[]): void {
    if (this.config.enableDebug) {
      console[level](`[AdxSDK] ${message}`, ...args);
    }
  }
}

// Export singleton instance
const adxSDK = new AdxSDK();

// Export new TaskirX Client
export { TaskirXClient as default } from './client';
export { TaskirXClient } from './client';
export * from './types';
export * from './services/AuthService';
export * from './services/CampaignService';
export * from './services/AnalyticsService';
export * from './services/BiddingService';
export * from './services/AdService';
export * from './services/WebhookService';
export { RequestManager } from './services/RequestManager';
export { Logger } from './utils/Logger';
export { TaskirXError, ErrorHandler } from './utils/ErrorHandler';

// Legacy ADX SDK export
export { adxSDK };
