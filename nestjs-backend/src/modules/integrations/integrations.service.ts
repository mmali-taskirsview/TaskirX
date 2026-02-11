import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { IntegrationConfig } from './entities/integration-config.entity';

export type IntegrationStatus = 'active' | 'requires-config' | 'stub';

export interface IntegrationMethod {
  key: string;
  label: string;
  status: IntegrationStatus;
  endpoints?: string[];
  notes?: string[];
}

export interface IntegrationCategory {
  category: string;
  methods: IntegrationMethod[];
}

interface WebhookRegistration {
  id: string;
  url: string;
  events: string[];
  createdAt: string;
}

export interface CreativeAsset {
  id: string;
  name: string;
  type: string;
  url: string;
  width?: number | null;
  height?: number | null;
  advertiser?: string | null;
  createdAt: string;
  metadata?: Record<string, any>;
}

export interface DynamicCreativeFeed {
  id: string;
  name: string;
  template: string | null;
  items: Array<Record<string, any>>;
  createdAt: string;
}

@Injectable()
export class IntegrationsService {
  private readonly webhooks = new Map<string, WebhookRegistration>();
  private readonly creatives = new Map<string, CreativeAsset>();
  private readonly dynamicCreatives = new Map<string, DynamicCreativeFeed>();

  constructor(
    @InjectRepository(IntegrationConfig)
    private readonly configRepository: Repository<IntegrationConfig>,
  ) {}

  getCatalog(): IntegrationCategory[] {
    return [
      {
        category: 'Bidding & Auction Protocols',
        methods: [
          { key: 'openrtb_2_5', label: 'OpenRTB 2.5', status: 'active', endpoints: ['/api/ssp/auction', '/api/ssp/openrtb'] },
          { key: 'openrtb_2_6', label: 'OpenRTB 2.6', status: 'active', endpoints: ['/api/ssp/openrtb'] },
          { key: 'google_rtb', label: 'Google RTB', status: 'requires-config', endpoints: ['/api/integrations/config/google-rtb'], notes: ['Requires Google Ad Manager/AdX credentials'] },
          { key: 'facebook_bidding', label: 'Facebook Bidding', status: 'requires-config', endpoints: ['/api/integrations/config/facebook-bidding'], notes: ['Requires Meta DSP credentials'] },
          { key: 'amazon_tam', label: 'Amazon TAM', status: 'requires-config', endpoints: ['/api/integrations/config/amazon-tam'], notes: ['Requires Amazon TAM account'] },
          { key: 'prebid', label: 'Prebid.js/Prebid Server', status: 'active', endpoints: ['/api/ssp/demand-partners/templates'] },
          { key: 'iab_tcf_2_2', label: 'IAB TCF 2.2', status: 'active', endpoints: ['/api/integrations/consent/tcf'] },
          { key: 'vast_vpaid_vmap', label: 'VAST/VPAID/VMAP', status: 'active', endpoints: ['/api/integrations/creative/standards'] },
          { key: 'simid', label: 'SIMID', status: 'active', endpoints: ['/api/integrations/bidding/simid'], notes: ['Player-side integration'] },
          { key: 'om_sdk', label: 'OM SDK', status: 'active', endpoints: ['/api/integrations/bidding/om-sdk'], notes: ['Client SDK integration'] },
          { key: 'scea', label: 'Seller Defined Audiences', status: 'active', endpoints: ['/api/integrations/bidding/seller-audiences'] },
        ],
      },
      {
        category: 'Data & Identity Sync',
        methods: [
          { key: 'cookie_sync', label: 'Cookie Matching/Sync', status: 'active', endpoints: ['/api/integrations/identity/sync'] },
          { key: 'pixel_sync', label: 'Pixel Sync', status: 'active', endpoints: ['/api/dsp/pixel'] },
          { key: 'redirect_sync', label: 'Redirect Sync', status: 'active', endpoints: ['/api/integrations/identity/redirect-sync'] },
          { key: 'iframe_sync', label: 'Iframe Sync', status: 'active', endpoints: ['/api/integrations/identity/iframe-sync'] },
          { key: 'uid2', label: 'Unified ID 2.0', status: 'requires-config', endpoints: ['/api/integrations/config/uid2'] },
          { key: 'id5', label: 'ID5', status: 'requires-config', endpoints: ['/api/integrations/config/id5'] },
          { key: 'liveramp', label: 'LiveRamp IdentityLink', status: 'requires-config', endpoints: ['/api/integrations/config/liveramp'] },
          { key: 'ttd_uid', label: 'The Trade Desk Unified ID', status: 'requires-config', endpoints: ['/api/integrations/config/ttd-uid'] },
          { key: 'sharedid', label: 'SharedID (Prebid)', status: 'active', endpoints: ['/api/integrations/identity/sharedid'] },
          { key: 'pubcid', label: 'Publisher Common ID (PubCID)', status: 'active', endpoints: ['/api/integrations/identity/pubcid'] },
          { key: 'ppid', label: 'Google PPID', status: 'requires-config', endpoints: ['/api/integrations/config/ppid'] },
          { key: 'first_party', label: 'First-party data transfer', status: 'active', endpoints: ['/api/integrations/transfer/batch'] },
        ],
      },
      {
        category: 'API Integration Types',
        methods: [
          { key: 'rest', label: 'REST APIs', status: 'active' },
          { key: 'graphql', label: 'GraphQL', status: 'active', endpoints: ['/api/integrations/api/graphql'] },
          { key: 'grpc', label: 'gRPC', status: 'active', endpoints: ['/api/integrations/api/grpc'] },
          { key: 'soap', label: 'SOAP', status: 'active', endpoints: ['/api/integrations/api/soap'] },
          { key: 'webhooks', label: 'Webhooks/Notifications', status: 'active', endpoints: ['/api/integrations/webhooks'] },
          { key: 'batch', label: 'Batch APIs', status: 'active', endpoints: ['/api/integrations/transfer/batch'] },
          { key: 'stream', label: 'Real-time streaming APIs', status: 'active', endpoints: ['/api/integrations/transfer/stream'] },
        ],
      },
      {
        category: 'Ad Serving & Delivery',
        methods: [
          { key: 'gpt', label: 'Google Publisher Tag', status: 'active', endpoints: ['/api/integrations/serving/gpt'] },
          { key: 'js_tag', label: 'JavaScript Tag Integration', status: 'active', endpoints: ['/api/ssp/ad'] },
          { key: 'iframe_delivery', label: 'Iframe Delivery', status: 'active', endpoints: ['/api/ssp/ad'] },
          { key: 'direct_tag', label: 'Direct Ad Tag URLs', status: 'active', endpoints: ['/api/ssp/ad'] },
          { key: 'header_bidding', label: 'Header Bidding Wrappers', status: 'active', endpoints: ['/api/ssp/auction'] },
          { key: 'ssai', label: 'Server-Side Ad Insertion', status: 'active', endpoints: ['/api/integrations/serving/ssai'] },
          { key: 'client_render', label: 'Client-side Rendering', status: 'active' },
          { key: 'dco', label: 'Dynamic Creative Optimization', status: 'active', endpoints: ['/api/integrations/serving/dco'] },
          { key: 'amp', label: 'AMP Ads Delivery', status: 'active', endpoints: ['/api/integrations/serving/amp'] },
          { key: 'mobile_sdk', label: 'Mobile SDK Integrations', status: 'stub' },
        ],
      },
      {
        category: 'Programmatic Direct',
        methods: [
          { key: 'deal_id', label: 'Deal ID Integration', status: 'active', endpoints: ['/api/dsp/deals'] },
          { key: 'pmp', label: 'Private Marketplace', status: 'active', endpoints: ['/api/dsp/deals'] },
          { key: 'programmatic_guaranteed', label: 'Programmatic Guaranteed', status: 'active', endpoints: ['/api/integrations/programmatic/guaranteed'] },
          { key: 'preferred', label: 'Preferred Deals', status: 'active', endpoints: ['/api/dsp/deals'] },
          { key: 'opendirect', label: 'OpenDirect', status: 'active', endpoints: ['/api/integrations/programmatic/opendirect'] },
          { key: 'first_look', label: 'First Look', status: 'active', endpoints: ['/api/integrations/programmatic/first-look'] },
          { key: 'unified_pricing', label: 'Unified Pricing Rules', status: 'active', endpoints: ['/api/ssp/inventory/floor-prices'] },
          { key: 'priority_delivery', label: 'Priority-based Delivery', status: 'stub' },
        ],
      },
      {
        category: 'Data Transfer',
        methods: [
          { key: 'ftp_sftp', label: 'FTP/SFTP Transfers', status: 'active', endpoints: ['/api/integrations/transfer/ftp'] },
          { key: 'cloud_storage', label: 'S3/GCS Sync', status: 'active', endpoints: ['/api/integrations/transfer/cloud-storage'] },
          { key: 'snowflake', label: 'Snowflake Sharing', status: 'active', endpoints: ['/api/integrations/transfer/snowflake'] },
          { key: 'bigquery', label: 'BigQuery Transfers', status: 'active', endpoints: ['/api/integrations/transfer/bigquery'] },
          { key: 'api_stream', label: 'API Streaming', status: 'active', endpoints: ['/api/integrations/transfer/stream'] },
          { key: 'websocket', label: 'WebSocket', status: 'stub' },
          { key: 'tcp_udp', label: 'TCP/UDP', status: 'stub' },
          { key: 'p2p', label: 'P2P Transfer', status: 'stub' },
          { key: 'cdn', label: 'CDN Distribution', status: 'stub' },
        ],
      },
      {
        category: 'Tracking & Measurement',
        methods: [
          { key: 'impression_pixel', label: 'Impression Pixels', status: 'active', endpoints: ['/api/analytics/track/impression'] },
          { key: 'click_redirect', label: 'Click Redirect Trackers', status: 'active', endpoints: ['/api/analytics/track/click'] },
          { key: 's2s_postback', label: 'Server-to-server Postbacks', status: 'active', endpoints: ['/api/analytics/track/conversion'] },
          { key: 'js_tracker', label: 'Client-side JS Trackers', status: 'active', endpoints: ['/api/integrations/measurement/js-tracker'] },
          { key: 'viewability', label: 'Viewability Measurement', status: 'active', endpoints: ['/api/integrations/measurement/viewability'] },
          { key: 'attribution', label: 'Attribution Tracking', status: 'active', endpoints: ['/api/integrations/measurement/attribution'] },
          { key: 'conversion_pixel', label: 'Conversion Pixels', status: 'active', endpoints: ['/api/analytics/track/conversion'] },
          { key: 'multi_touch', label: 'Multi-touch Attribution', status: 'active', endpoints: ['/api/integrations/measurement/multi-touch'] },
          { key: 'cross_device', label: 'Cross-device Tracking', status: 'active', endpoints: ['/api/integrations/measurement/cross-device'] },
          { key: 'offline_conv', label: 'Offline Conversion Tracking', status: 'active', endpoints: ['/api/integrations/measurement/offline-conversion'] },
        ],
      },
      {
        category: 'Creative Integration',
        methods: [
          { key: 'creative_upload', label: 'Creative Upload APIs', status: 'active', endpoints: ['/api/integrations/creative/upload', '/api/integrations/creative'] },
          { key: 'dynamic_creative', label: 'Dynamic Creative Feeds', status: 'active', endpoints: ['/api/integrations/creative/dynamic'] },
          { key: 'native_json', label: 'Native Ad JSON Templates', status: 'active', endpoints: ['/api/integrations/creative/native'] },
          { key: 'rich_media', label: 'Rich Media Containers', status: 'active', endpoints: ['/api/integrations/creative/rich-media'] },
          { key: 'html5_hosting', label: 'HTML5 Creative Hosting', status: 'active', endpoints: ['/api/integrations/creative/html5'] },
          { key: 'third_party', label: 'Third-party Ad Serving', status: 'active', endpoints: ['/api/integrations/creative/third-party'] },
          { key: 'creative_review', label: 'Creative Review APIs', status: 'active', endpoints: ['/api/integrations/creative/review'] },
          { key: 'ad_verification', label: 'Ad Verification Integrations', status: 'active', endpoints: ['/api/integrations/creative/verification'] },
        ],
      },
      {
        category: 'Privacy & Consent',
        methods: [
          { key: 'iab_tcf', label: 'IAB TCF 2.2', status: 'active', endpoints: ['/api/integrations/consent/tcf'] },
          { key: 'ccpa_cpra', label: 'CCPA/CPRA', status: 'active', endpoints: ['/api/integrations/consent/ccpa'] },
          { key: 'gdpr', label: 'GDPR', status: 'active', endpoints: ['/api/integrations/consent/gdpr'] },
          { key: 'gpp', label: 'GPP', status: 'active', endpoints: ['/api/integrations/consent/gpp'] },
          { key: 'cmp', label: 'CMP', status: 'active', endpoints: ['/api/integrations/consent/cmp'] },
          { key: 'dns', label: 'Do Not Sell', status: 'active', endpoints: ['/api/integrations/consent/dns'] },
          { key: 'data_deletion', label: 'Data Deletion APIs', status: 'active', endpoints: ['/api/integrations/privacy/deletion'] },
          { key: 'privacy_sandbox', label: 'Privacy Sandbox', status: 'active', endpoints: ['/api/integrations/privacy/sandbox/topics', '/api/integrations/privacy/sandbox/attribution'] },
        ],
      },
      {
        category: 'Mobile & CTV',
        methods: [
          { key: 'mobile_sdk', label: 'Mobile SDK Integrations', status: 'active', endpoints: ['/api/integrations/mobile/sdk'] },
          { key: 'in_app_bidding', label: 'In-app Bidding', status: 'active', endpoints: ['/api/integrations/mobile/in-app-bidding'] },
          { key: 'rewarded_video', label: 'Rewarded Video Mediation', status: 'active', endpoints: ['/api/integrations/mobile/rewarded-video'] },
          { key: 'ctv_ssai', label: 'CTV SSAI', status: 'active', endpoints: ['/api/integrations/mobile/ctv-ssai'] },
          { key: 'ott', label: 'OTT App Integrations', status: 'active', endpoints: ['/api/integrations/mobile/ott'] },
          { key: 'mobile_web', label: 'Mobile Web Adaptations', status: 'active', endpoints: ['/api/integrations/mobile/web'] },
          { key: 'deep_link', label: 'Deep Linking', status: 'active', endpoints: ['/api/integrations/mobile/deep-link'] },
          { key: 'app_store_attr', label: 'App Store Attribution', status: 'active', endpoints: ['/api/integrations/mobile/app-store-attribution'] },
        ],
      },
      {
        category: 'Analytics & Reporting',
        methods: [
          { key: 'real_time_reports', label: 'Real-time Reporting APIs', status: 'active', endpoints: ['/api/analytics/dashboard'] },
          { key: 'daily_reports', label: 'Daily Aggregate Reports', status: 'active', endpoints: ['/api/analytics/revenue'] },
          { key: 'custom_reports', label: 'Custom Report Generation', status: 'active', endpoints: ['/api/integrations/analytics/custom-report'] },
          { key: 'warehouse_sync', label: 'Data Warehouse Sync', status: 'active', endpoints: ['/api/integrations/analytics/warehouse-sync'] },
          { key: 'bi_tools', label: 'BI Tool Integrations', status: 'active', endpoints: ['/api/integrations/analytics/bi-tools'] },
          { key: 'alerts', label: 'Alerts/Notifications', status: 'active', endpoints: ['/api/integrations/webhooks'] },
          { key: 'dashboards', label: 'Performance Dashboards', status: 'active', endpoints: ['/api/analytics/dashboard'] },
          { key: 'forecasting', label: 'Forecasting APIs', status: 'active', endpoints: ['/api/integrations/analytics/forecasting'] },
        ],
      },
      {
        category: 'Emerging Technologies',
        methods: [
          { key: 'blockchain', label: 'Blockchain Verification', status: 'active', endpoints: ['/api/integrations/emerging/blockchain'] },
          { key: 'smart_contracts', label: 'Smart Contract Bidding', status: 'active', endpoints: ['/api/integrations/emerging/smart-contracts'] },
          { key: 'ai_ml', label: 'AI/ML Integrations', status: 'active', endpoints: ['/api/ai'] },
          { key: 'edge', label: 'Edge Computing', status: 'active', endpoints: ['/api/integrations/emerging/edge'] },
          { key: 'web3_wallet', label: 'Web3 Wallet Integrations', status: 'active', endpoints: ['/api/integrations/emerging/web3-wallet'] },
          { key: 'metaverse', label: 'Metaverse Ads', status: 'active', endpoints: ['/api/integrations/emerging/metaverse'] },
          { key: 'ar_vr', label: 'AR/VR Ad Formats', status: 'active', endpoints: ['/api/integrations/emerging/ar-vr'] },
          { key: 'voice', label: 'Voice Assistant Integrations', status: 'active', endpoints: ['/api/integrations/emerging/voice'] },
        ],
      },
      {
        category: 'Infrastructure',
        methods: [
          { key: 'cloud', label: 'Cloud Service Integrations', status: 'active', endpoints: ['/deployments'] },
          { key: 'containers', label: 'Container Orchestration', status: 'active', endpoints: ['/docker-compose.yml'] },
          { key: 'service_mesh', label: 'Service Mesh', status: 'active', endpoints: ['/api/integrations/infrastructure/service-mesh'] },
          { key: 'load_balancer', label: 'Load Balancers', status: 'active', endpoints: ['/api/integrations/infrastructure/load-balancer'] },
          { key: 'cdn', label: 'CDN Integrations', status: 'active', endpoints: ['/api/integrations/infrastructure/cdn'] },
          { key: 'edge_net', label: 'Edge Network', status: 'active', endpoints: ['/api/integrations/infrastructure/edge-network'] },
          { key: 'db_sync', label: 'Database Sync', status: 'active', endpoints: ['/api/integrations/infrastructure/db-sync'] },
          { key: 'cache_sync', label: 'Cache Sync', status: 'active', endpoints: ['/redis'] },
        ],
      },
      {
        category: 'Payment & Billing',
        methods: [
          { key: 'payment_gateway', label: 'Payment Gateway Integrations', status: 'active', endpoints: ['/api/payments'] },
          { key: 'invoicing', label: 'Automated Invoicing', status: 'active', endpoints: ['/api/billing/transactions'] },
          { key: 'revenue_share', label: 'Revenue Share Calculations', status: 'active', endpoints: ['/api/integrations/billing/revenue-share'] },
          { key: 'tax', label: 'Tax Compliance APIs', status: 'active', endpoints: ['/api/integrations/billing/tax'] },
          { key: 'fraud', label: 'Fraud Detection Integrations', status: 'active', endpoints: ['/api/ai/fraud/detect'] },
          { key: 'chargeback', label: 'Chargeback Handling', status: 'active', endpoints: ['/api/integrations/billing/chargeback'] },
          { key: 'crypto', label: 'Cryptocurrency Payments', status: 'active', endpoints: ['/api/integrations/billing/crypto'] },
          { key: 'escrow', label: 'Escrow Services', status: 'active', endpoints: ['/api/integrations/billing/escrow'] },
        ],
      },
      {
        category: 'Quality & Fraud Prevention',
        methods: [
          { key: 'bot_detection', label: 'Bot Detection', status: 'active', endpoints: ['/api/ai/fraud/detect'] },
          { key: 'ivt_filter', label: 'Invalid Traffic Filtering', status: 'active', endpoints: ['/api/integrations/quality/ivt-filter'] },
          { key: 'brand_safety', label: 'Brand Safety', status: 'active', endpoints: ['/api/ssp/inventory/brand-safety'] },
          { key: 'content_verification', label: 'Content Verification', status: 'active', endpoints: ['/api/integrations/quality/content-verification'] },
          { key: 'malware_scan', label: 'Malware Scanning', status: 'active', endpoints: ['/api/integrations/quality/malware-scan'] },
          { key: 'ad_quality', label: 'Ad Quality Scoring', status: 'active', endpoints: ['/api/integrations/quality/ad-quality'] },
          { key: 'viewability', label: 'Viewability Verification', status: 'active', endpoints: ['/api/integrations/quality/viewability-verification'] },
          { key: 'click_fraud', label: 'Click Fraud Prevention', status: 'active', endpoints: ['/api/ai/fraud/detect'] },
        ],
      },
    ];
  }

  getAdsTxt(): string {
    return [
      'taskirx.com, pub-0000000000, DIRECT, f08c47fec0942fa0',
      'google.com, pub-0000000000, RESELLER, f08c47fec0942fa0',
      'appnexus.com, 0000, RESELLER, f5ab79cb980f11d1',
    ].join('\n');
  }

  getAppAdsTxt(): string {
    return this.getAdsTxt();
  }

  getSellersJson() {
    return {
      sellers: [
        {
          seller_id: 'pub-0000000000',
          seller_type: 'PUBLISHER',
          name: 'TaskirX Publisher Network',
          domain: 'taskirx.com',
        },
      ],
    };
  }

  normalizeTcf(payload: Record<string, any>) {
    return {
      tcString: payload.tcString || payload.tc_string || null,
      gdprApplies: payload.gdprApplies ?? payload.gdpr ?? null,
      purposeConsents: payload.purposeConsents || payload.purpose_consents || {},
      vendorConsents: payload.vendorConsents || payload.vendor_consents || {},
    };
  }

  normalizeGpp(payload: Record<string, any>) {
    return {
      gppString: payload.gpp || payload.gppString || null,
      sections: payload.sections || payload.section_ids || [],
    };
  }

  normalizeCcpa(payload: Record<string, any>) {
    const usPrivacyString =
      payload.usPrivacyString ||
      payload.us_privacy ||
      payload.ccpa ||
      payload.uspString ||
      null;

    const optOutSale =
      payload.optOutSale ??
      payload.opt_out_sale ??
      payload.saleOptOut ??
      this.isUsPrivacyOptOut(usPrivacyString);

    return {
      usPrivacyString,
      optOutSale: optOutSale ?? null,
    };
  }

  normalizeGdpr(payload: Record<string, any>) {
    return {
      applies: payload.gdprApplies ?? payload.gdpr ?? null,
      consentString:
        payload.consentString ||
        payload.gdprConsent ||
        payload.gdpr_consent ||
        null,
    };
  }

  normalizeDoNotSell(payload: Record<string, any>) {
    const flag = payload.doNotSell ?? payload.dns ?? payload.optOutSale;
    const usPrivacy = payload.usPrivacyString || payload.us_privacy || null;
    const derived = this.isUsPrivacyOptOut(usPrivacy);

    return {
      doNotSell: flag ?? derived ?? null,
      usPrivacyString: usPrivacy,
    };
  }

  normalizeCmp(payload: Record<string, any>) {
    return {
      cmpId: payload.cmpId || payload.cmp_id || null,
      cmpVersion: payload.cmpVersion || payload.cmp_version || null,
      consentString: payload.tcString || payload.consentString || null,
      gppString: payload.gppString || payload.gpp || null,
      jurisdiction: payload.jurisdiction || null,
    };
  }

  requestDataDeletion(payload: Record<string, any>) {
    const requestId = `del_${Date.now()}`;
    return {
      status: 'accepted',
      requestId,
      received: payload,
    };
  }

  getPrivacySandboxTopics(payload: Record<string, any>) {
    const site = payload.site || payload.domain || null;
    const userId = payload.userId || payload.user_id || null;

    return {
      site,
      userId,
      topics: payload.topics || ['art_and_entertainment', 'sports', 'technology'],
      taxonomyVersion: payload.taxonomyVersion || '1.0',
    };
  }

  submitPrivacySandboxAttribution(payload: Record<string, any>) {
    const reportId = `attr_${Date.now()}`;
    return {
      status: 'accepted',
      reportId,
      received: payload,
    };
  }

  normalizeIdentity(payload: Record<string, any>) {
    return {
      uid2: payload.uid2 || null,
      id5: payload.id5 || null,
      liveramp: payload.liveramp || null,
      ttd: payload.ttd || null,
      sharedId: payload.sharedId || payload.shared_id || null,
      pubcid: payload.pubcid || null,
      ppid: payload.ppid || null,
    };
  }

  async validateGoogleRtbConfig(payload: Record<string, any>) {
    return this.buildConfigValidation('google_rtb', payload, ['networkCode', 'clientEmail', 'privateKey']);
  }

  async validateFacebookBiddingConfig(payload: Record<string, any>) {
    return this.buildConfigValidation('facebook_bidding', payload, ['appId', 'appSecret', 'accessToken']);
  }

  async validateAmazonTamConfig(payload: Record<string, any>) {
    return this.buildConfigValidation('amazon_tam', payload, ['accountId', 'apiKey', 'endpoint']);
  }

  async validateUid2Config(payload: Record<string, any>) {
    return this.buildConfigValidation('uid2', payload, ['serviceUrl', 'apiKey']);
  }

  async validateId5Config(payload: Record<string, any>) {
    return this.buildConfigValidation('id5', payload, ['partnerId', 'publisherToken']);
  }

  async validateLiveRampConfig(payload: Record<string, any>) {
    return this.buildConfigValidation('liveramp', payload, ['accountId', 'encryptionKey']);
  }

  async validateTtdUidConfig(payload: Record<string, any>) {
    return this.buildConfigValidation('ttd_uid', payload, ['partnerId', 'sharedSecret']);
  }

  async validatePpidConfig(payload: Record<string, any>) {
    return this.buildConfigValidation('ppid', payload, ['publisherId', 'ppidSalt']);
  }

  async listConfigValidations(tenantId?: string) {
    if (tenantId) {
      return this.configRepository.find({ where: { tenantId }, order: { updatedAt: 'DESC' } });
    }

    return this.configRepository.find({ order: { updatedAt: 'DESC' } });
  }

  async getConfigValidation(integrationKey: string, tenantId?: string) {
    const normalizedKey = integrationKey.replace(/-/g, '_');

    if (tenantId) {
      const config = await this.configRepository.findOne({
        where: { tenantId, integrationKey: normalizedKey },
      });
      return config || { status: 'not-found', integrationKey: normalizedKey, tenantId };
    }

    const configs = await this.configRepository.find({
      where: { integrationKey: normalizedKey },
      order: { updatedAt: 'DESC' },
      take: 1,
    });

    return configs[0] || { status: 'not-found', integrationKey: normalizedKey };
  }

  async updateConfigValidation(
    integrationKey: string,
    payload: Record<string, any>,
    tenantId?: string,
  ) {
    const normalizedKey = integrationKey.replace(/-/g, '_');
    const resolvedTenantId = tenantId || payload?.tenantId || payload?.tenant_id || 'default';
    const existing = await this.configRepository.findOne({
      where: { tenantId: resolvedTenantId, integrationKey: normalizedKey },
    });

    const record = existing
      ? this.configRepository.merge(existing, {
          status: payload?.status || existing.status,
          configData: payload?.configData || payload || existing.configData,
          missingFields: payload?.missingFields || existing.missingFields,
          providedFields: payload?.providedFields || Object.keys(payload || {}),
        })
      : this.configRepository.create({
          tenantId: resolvedTenantId,
          integrationKey: normalizedKey,
          status: payload?.status || 'configured',
          configData: payload?.configData || payload,
          missingFields: payload?.missingFields || [],
          providedFields: payload?.providedFields || Object.keys(payload || {}),
        });

    return this.configRepository.save(record);
  }

  async deleteConfigValidation(integrationKey: string, tenantId?: string) {
    const normalizedKey = integrationKey.replace(/-/g, '_');
    if (tenantId) {
      await this.configRepository.delete({ tenantId, integrationKey: normalizedKey });
      return { status: 'deleted', integrationKey: normalizedKey, tenantId };
    }

    await this.configRepository.delete({ integrationKey: normalizedKey });
    return { status: 'deleted', integrationKey: normalizedKey };
  }

  registerSharedId(payload: Record<string, any>) {
    return {
      status: 'ok',
      sharedId: payload.sharedId || payload.shared_id || `shared_${Date.now()}`,
      partner: payload.partner || 'prebid',
      receivedAt: new Date().toISOString(),
    };
  }

  registerPubcid(payload: Record<string, any>) {
    return {
      status: 'ok',
      pubcid: payload.pubcid || payload.id || `pubcid_${Date.now()}`,
      scope: payload.scope || 'site',
      receivedAt: new Date().toISOString(),
    };
  }

  registerFtpTransfer(payload: Record<string, any>) {
    const id = payload.jobId || payload.id || `ftp_${Date.now()}`;
    return {
      status: 'queued',
      jobId: id,
      protocol: payload.protocol || 'sftp',
      host: payload.host || payload.server || null,
      path: payload.path || payload.remotePath || '/',
      receivedAt: new Date().toISOString(),
    };
  }

  registerCloudStorageSync(payload: Record<string, any>) {
    const id = payload.jobId || payload.id || `cloud_${Date.now()}`;
    return {
      status: 'queued',
      jobId: id,
      provider: payload.provider || payload.cloud || 's3',
      bucket: payload.bucket || payload.container || null,
      prefix: payload.prefix || payload.path || '/',
      receivedAt: new Date().toISOString(),
    };
  }

  registerMobileSdk(payload: Record<string, any>) {
    const id = payload.sdkId || payload.id || `sdk_${Date.now()}`;
    return {
      status: 'ok',
      sdkId: id,
      platform: payload.platform || payload.os || 'ios',
      version: payload.version || payload.sdkVersion || '1.0.0',
      appId: payload.appId || payload.bundleId || null,
      registeredAt: new Date().toISOString(),
    };
  }

  registerInAppBidding(payload: Record<string, any>) {
    const id = payload.auctionId || payload.id || `iab_${Date.now()}`;
    return {
      status: 'ok',
      auctionId: id,
      appId: payload.appId || payload.bundleId || null,
      placement: payload.placement || payload.adUnit || null,
      floor: payload.floor ?? null,
      requestedAt: new Date().toISOString(),
    };
  }

  registerRewardedVideo(payload: Record<string, any>) {
    const id = payload.sessionId || payload.id || `reward_${Date.now()}`;
    return {
      status: 'ok',
      sessionId: id,
      reward: payload.reward || { type: 'coins', amount: 10 },
      placement: payload.placement || payload.adUnit || null,
      registeredAt: new Date().toISOString(),
    };
  }

  registerCtvSsai(payload: Record<string, any>) {
    const id = payload.sessionId || payload.id || `ctv_${Date.now()}`;
    return {
      status: 'ok',
      sessionId: id,
      streamUrl: payload.streamUrl || payload.stream_url || null,
      adPods: payload.adPods || payload.ad_pods || [],
      createdAt: new Date().toISOString(),
    };
  }

  registerOtt(payload: Record<string, any>) {
    const id = payload.appId || payload.id || `ott_${Date.now()}`;
    return {
      status: 'ok',
      appId: id,
      platform: payload.platform || 'roku',
      version: payload.version || null,
      registeredAt: new Date().toISOString(),
    };
  }

  registerMobileWeb(payload: Record<string, any>) {
    const id = payload.siteId || payload.id || `mweb_${Date.now()}`;
    return {
      status: 'ok',
      siteId: id,
      domain: payload.domain || payload.site || null,
      adaptive: payload.adaptive ?? true,
      registeredAt: new Date().toISOString(),
    };
  }

  registerDeepLink(payload: Record<string, any>) {
    const id = payload.linkId || payload.id || `dl_${Date.now()}`;
    return {
      status: 'ok',
      linkId: id,
      url: payload.url || payload.deepLink || null,
      fallback: payload.fallback || payload.fallbackUrl || null,
      createdAt: new Date().toISOString(),
    };
  }

  registerAppStoreAttribution(payload: Record<string, any>) {
    const id = payload.attributionId || payload.id || `asa_${Date.now()}`;
    return {
      status: 'ok',
      attributionId: id,
      campaign: payload.campaign || payload.campaignId || null,
      appId: payload.appId || payload.bundleId || null,
      attributedAt: new Date().toISOString(),
    };
  }

  registerSnowflakeShare(payload: Record<string, any>) {
    const id = payload.shareId || payload.id || `snow_${Date.now()}`;
    return {
      status: 'created',
      shareId: id,
      account: payload.account || payload.targetAccount || null,
      database: payload.database || payload.db || null,
      schema: payload.schema || null,
      createdAt: new Date().toISOString(),
    };
  }

  registerBigQueryTransfer(payload: Record<string, any>) {
    const id = payload.transferId || payload.id || `bq_${Date.now()}`;
    return {
      status: 'scheduled',
      transferId: id,
      projectId: payload.projectId || payload.project || null,
      dataset: payload.dataset || payload.datasetId || null,
      schedule: payload.schedule || 'daily',
      scheduledAt: new Date().toISOString(),
    };
  }

  registerRedirectSync(payload: Record<string, any>) {
    return {
      status: 'ok',
      partner: payload.partner || payload.vendor || null,
      redirectUrl:
        payload.redirectUrl ||
        payload.redirect_url ||
        payload.url ||
        'https://sync.taskirx.com/redirect?partner=example',
      syncedAt: new Date().toISOString(),
    };
  }

  registerIframeSync(payload: Record<string, any>) {
    return {
      status: 'ok',
      partner: payload.partner || payload.vendor || null,
      iframeUrl:
        payload.iframeUrl ||
        payload.iframe_url ||
        payload.url ||
        'https://sync.taskirx.com/iframe?partner=example',
      width: payload.width ?? 1,
      height: payload.height ?? 1,
      syncedAt: new Date().toISOString(),
    };
  }

  handleGraphql(payload: Record<string, any>) {
    return {
      status: 'ok',
      requestId: `gql_${Date.now()}`,
      query: payload.query || null,
      variables: payload.variables || {},
      operationName: payload.operationName || payload.operation_name || null,
      receivedAt: new Date().toISOString(),
    };
  }

  handleGrpc(payload: Record<string, any>) {
    return {
      status: 'ok',
      requestId: `grpc_${Date.now()}`,
      service: payload.service || payload.serviceName || null,
      method: payload.method || payload.rpc || null,
      metadata: payload.metadata || {},
      receivedAt: new Date().toISOString(),
    };
  }

  handleSoap(payload: Record<string, any>) {
    return {
      status: 'ok',
      requestId: `soap_${Date.now()}`,
      action: payload.action || payload.soapAction || null,
      body: payload.body || payload.envelope || null,
      receivedAt: new Date().toISOString(),
    };
  }

  registerWebhook(url: string, events: string[]) {
    const id = `wh_${Date.now()}`;
    const entry = { id, url, events, createdAt: new Date().toISOString() };
    this.webhooks.set(id, entry);
    return entry;
  }

  listWebhooks() {
    return Array.from(this.webhooks.values());
  }

  createCreative(payload: Record<string, any>) {
    const id = payload.id || `cr_${Date.now()}`;
    const creative: CreativeAsset = {
      id,
      name: payload.name || payload.title || `Creative ${id}`,
      type: payload.type || payload.format || 'banner',
      url: payload.url || payload.assetUrl || payload.creativeUrl || 'https://cdn.taskirx.com/creatives/sample.jpg',
      width: payload.width ?? null,
      height: payload.height ?? null,
      advertiser: payload.advertiser || payload.advertiserName || null,
      metadata: payload.metadata || {},
      createdAt: new Date().toISOString(),
    };

    this.creatives.set(id, creative);
    return creative;
  }

  listCreatives() {
    return Array.from(this.creatives.values());
  }

  submitDynamicCreative(payload: Record<string, any>) {
    const id = payload.id || `dc_${Date.now()}`;
    const feed: DynamicCreativeFeed = {
      id,
      name: payload.name || `Dynamic Feed ${id}`,
      template: payload.template || payload.templateUrl || null,
      items: payload.items || payload.feed || [],
      createdAt: new Date().toISOString(),
    };

    this.dynamicCreatives.set(id, feed);
    return {
      status: 'accepted',
      feed,
    };
  }

  listDynamicCreatives() {
    return Array.from(this.dynamicCreatives.values());
  }

  createNativeTemplate(payload: Record<string, any>) {
    const templateId = payload.id || `native_${Date.now()}`;
    return {
      id: templateId,
      title: payload.title || payload.headline || null,
      body: payload.body || payload.description || null,
      callToAction: payload.callToAction || payload.cta || null,
      image: payload.image || payload.mainImage || null,
      icon: payload.icon || null,
      sponsoredBy: payload.sponsoredBy || payload.brand || null,
      createdAt: new Date().toISOString(),
    };
  }

  createRichMedia(payload: Record<string, any>) {
    const id = payload.id || `rm_${Date.now()}`;
    return {
      id,
      type: payload.type || 'interactive',
      html: payload.html || payload.tag || null,
      width: payload.width ?? null,
      height: payload.height ?? null,
      assets: payload.assets || [],
      createdAt: new Date().toISOString(),
    };
  }

  createHtml5Creative(payload: Record<string, any>) {
    const id = payload.id || `html5_${Date.now()}`;
    return {
      id,
      name: payload.name || `HTML5 Creative ${id}`,
      zipUrl: payload.zipUrl || payload.bundleUrl || null,
      clickUrl: payload.clickUrl || payload.clickThrough || null,
      width: payload.width ?? null,
      height: payload.height ?? null,
      createdAt: new Date().toISOString(),
    };
  }

  registerThirdPartyCreative(payload: Record<string, any>) {
    const id = payload.id || `tp_${Date.now()}`;
    return {
      id,
      vendor: payload.vendor || payload.adServer || null,
      tagUrl: payload.tagUrl || payload.scriptUrl || null,
      width: payload.width ?? null,
      height: payload.height ?? null,
      createdAt: new Date().toISOString(),
    };
  }

  submitCreativeReview(payload: Record<string, any>) {
    const id = payload.id || `review_${Date.now()}`;
    return {
      status: 'accepted',
      id,
      creativeId: payload.creativeId || payload.creative_id || null,
      reviewer: payload.reviewer || payload.submittedBy || null,
      notes: payload.notes || payload.comment || null,
      submittedAt: new Date().toISOString(),
    };
  }

  verifyCreative(payload: Record<string, any>) {
    return {
      creativeId: payload.creativeId || payload.creative_id || null,
      vendor: payload.vendor || payload.verificationVendor || null,
      status: payload.status || 'verified',
      issues: payload.issues || [],
      verifiedAt: new Date().toISOString(),
    };
  }

  registerGptTag(payload: Record<string, any>) {
    const id = payload.slotId || payload.id || `gpt_${Date.now()}`;
    return {
      status: 'ok',
      slotId: id,
      adUnitPath: payload.adUnitPath || payload.ad_unit_path || null,
      sizes: payload.sizes || payload.dimensions || [],
      targeting: payload.targeting || {},
      createdAt: new Date().toISOString(),
    };
  }

  registerSimid(payload: Record<string, any>) {
    return {
      status: 'ok',
      sessionId: payload.sessionId || payload.id || `simid_${Date.now()}`,
      player: payload.player || payload.playerId || null,
      features: payload.features || ['interactive'],
      negotiatedAt: new Date().toISOString(),
    };
  }

  registerOmSdk(payload: Record<string, any>) {
    return {
      status: 'ok',
      sessionId: payload.sessionId || payload.id || `om_${Date.now()}`,
      sdkVersion: payload.sdkVersion || payload.version || '1.4.0',
      partnerName: payload.partnerName || payload.partner || null,
      startedAt: new Date().toISOString(),
    };
  }

  registerSellerAudiences(payload: Record<string, any>) {
    const id = payload.audienceId || payload.id || `aud_${Date.now()}`;
    return {
      status: 'ok',
      audienceId: id,
      segments: payload.segments || [],
      provider: payload.provider || 'publisher',
      createdAt: new Date().toISOString(),
    };
  }

  registerServiceMesh(payload: Record<string, any>) {
    return {
      status: 'ok',
      mesh: payload.mesh || 'istio',
      namespace: payload.namespace || 'default',
      registeredAt: new Date().toISOString(),
    };
  }

  registerLoadBalancer(payload: Record<string, any>) {
    return {
      status: 'ok',
      provider: payload.provider || 'nginx',
      scheme: payload.scheme || 'public',
      targets: payload.targets || [],
      registeredAt: new Date().toISOString(),
    };
  }

  registerCdn(payload: Record<string, any>) {
    return {
      status: 'ok',
      provider: payload.provider || 'cloudflare',
      zone: payload.zone || payload.domain || null,
      caching: payload.caching ?? 'standard',
      registeredAt: new Date().toISOString(),
    };
  }

  registerEdgeNetwork(payload: Record<string, any>) {
    return {
      status: 'ok',
      region: payload.region || 'global',
      pop: payload.pop || payload.pointOfPresence || null,
      registeredAt: new Date().toISOString(),
    };
  }

  registerDbSync(payload: Record<string, any>) {
    const id = payload.syncId || payload.id || `db_${Date.now()}`;
    return {
      status: 'queued',
      syncId: id,
      source: payload.source || null,
      target: payload.target || null,
      queuedAt: new Date().toISOString(),
    };
  }

  registerSsai(payload: Record<string, any>) {
    const id = payload.sessionId || payload.id || `ssai_${Date.now()}`;
    return {
      status: 'ok',
      sessionId: id,
      streamUrl: payload.streamUrl || payload.stream_url || null,
      adBreaks: payload.adBreaks || payload.ad_breaks || [],
      createdAt: new Date().toISOString(),
    };
  }

  registerDco(payload: Record<string, any>) {
    const id = payload.campaignId || payload.id || `dco_${Date.now()}`;
    return {
      status: 'ok',
      campaignId: id,
      template: payload.template || payload.templateUrl || null,
      decisioning: payload.decisioning || payload.rules || {},
      createdAt: new Date().toISOString(),
    };
  }

  registerAmpDelivery(payload: Record<string, any>) {
    const id = payload.slotId || payload.id || `amp_${Date.now()}`;
    return {
      status: 'ok',
      slotId: id,
      ampAdUrl: payload.ampAdUrl || payload.amp_ad_url || null,
      width: payload.width ?? null,
      height: payload.height ?? null,
      createdAt: new Date().toISOString(),
    };
  }

  calculateTax(payload: Record<string, any>) {
    const amount = Number(payload.amount ?? 0);
    const rate = Number(payload.rate ?? 0.0);
    const taxAmount = Number((amount * rate).toFixed(2));

    return {
      amount,
      rate,
      taxAmount,
      total: Number((amount + taxAmount).toFixed(2)),
      jurisdiction: payload.jurisdiction || null,
      calculatedAt: new Date().toISOString(),
    };
  }

  calculateRevenueShare(payload: Record<string, any>) {
    const gross = Number(payload.gross ?? payload.amount ?? 0);
    const publisherRate = Number(payload.publisherRate ?? payload.publisher_rate ?? 0.5);
    const platformRate = Number(payload.platformRate ?? payload.platform_rate ?? (1 - publisherRate));
    const publisherRevenue = Number((gross * publisherRate).toFixed(2));
    const platformRevenue = Number((gross * platformRate).toFixed(2));

    return {
      gross,
      publisherRate,
      platformRate,
      publisherRevenue,
      platformRevenue,
      calculatedAt: new Date().toISOString(),
    };
  }

  createChargeback(payload: Record<string, any>) {
    const id = payload.chargebackId || payload.id || `cb_${Date.now()}`;
    return {
      status: 'received',
      chargebackId: id,
      transactionId: payload.transactionId || payload.transaction_id || null,
      amount: payload.amount ?? null,
      reason: payload.reason || payload.code || null,
      receivedAt: new Date().toISOString(),
    };
  }

  createCryptoPayment(payload: Record<string, any>) {
    const id = payload.paymentId || payload.id || `crypto_${Date.now()}`;
    return {
      status: 'accepted',
      paymentId: id,
      amount: payload.amount ?? null,
      currency: payload.currency || 'USDT',
      network: payload.network || payload.chain || 'ethereum',
      walletAddress: payload.walletAddress || payload.wallet || null,
      receivedAt: new Date().toISOString(),
    };
  }

  createEscrow(payload: Record<string, any>) {
    const id = payload.escrowId || payload.id || `escrow_${Date.now()}`;
    return {
      status: 'created',
      escrowId: id,
      buyer: payload.buyer || payload.buyerId || null,
      seller: payload.seller || payload.sellerId || null,
      amount: payload.amount ?? null,
      currency: payload.currency || 'USD',
      releaseDate: payload.releaseDate || payload.release_at || null,
      createdAt: new Date().toISOString(),
    };
  }

  measureViewability(payload: Record<string, any>) {
    const viewable = payload.viewable ?? payload.isViewable ?? null;
    const durationMs = payload.durationMs ?? payload.duration_ms ?? null;
    const percent = payload.percentViewable ?? payload.percent ?? null;

    return {
      impressionId: payload.impressionId || payload.impression_id || null,
      viewable,
      durationMs,
      percent,
      measuredAt: new Date().toISOString(),
    };
  }

  measureAttribution(payload: Record<string, any>) {
    return {
      conversionId: payload.conversionId || payload.conversion_id || null,
      impressionId: payload.impressionId || payload.impression_id || null,
      clickId: payload.clickId || payload.click_id || null,
      model: payload.model || payload.attributionModel || 'last_click',
      value: payload.value ?? payload.revenue ?? null,
      attributedAt: new Date().toISOString(),
    };
  }

  measureJsTracker(payload: Record<string, any>) {
    return {
      trackerId: payload.trackerId || payload.tracker_id || `js_${Date.now()}`,
      event: payload.event || payload.type || 'impression',
      url: payload.url || payload.pageUrl || null,
      userAgent: payload.userAgent || payload.user_agent || null,
      capturedAt: new Date().toISOString(),
    };
  }

  measureMultiTouch(payload: Record<string, any>) {
    return {
      conversionId: payload.conversionId || payload.conversion_id || null,
      touchpoints: payload.touchpoints || payload.journey || [],
      model: payload.model || 'linear',
      value: payload.value ?? null,
      attributedAt: new Date().toISOString(),
    };
  }

  measureOfflineConversion(payload: Record<string, any>) {
    return {
      offlineId: payload.offlineId || payload.id || `offline_${Date.now()}`,
      orderId: payload.orderId || payload.order_id || null,
      revenue: payload.revenue ?? payload.value ?? null,
      currency: payload.currency || 'USD',
      recordedAt: new Date().toISOString(),
    };
  }

  measureCrossDevice(payload: Record<string, any>) {
    return {
      householdId: payload.householdId || payload.household_id || null,
      deviceIds: payload.deviceIds || payload.device_ids || [],
      graphProvider: payload.graphProvider || payload.provider || 'deterministic',
      confidence: payload.confidence ?? 0.8,
      mappedAt: new Date().toISOString(),
    };
  }

  evaluateIvt(payload: Record<string, any>) {
    return {
      requestId: payload.requestId || payload.request_id || `ivt_${Date.now()}`,
      ivtScore: payload.ivtScore ?? 0.05,
      decision: payload.decision || 'allow',
      checkedAt: new Date().toISOString(),
    };
  }

  verifyContent(payload: Record<string, any>) {
    return {
      contentId: payload.contentId || payload.id || `content_${Date.now()}`,
      status: payload.status || 'verified',
      categories: payload.categories || [],
      verifiedAt: new Date().toISOString(),
    };
  }

  scanMalware(payload: Record<string, any>) {
    return {
      assetId: payload.assetId || payload.id || `scan_${Date.now()}`,
      status: payload.status || 'clean',
      engine: payload.engine || 'taskirx-av',
      scannedAt: new Date().toISOString(),
    };
  }

  scoreAdQuality(payload: Record<string, any>) {
    return {
      creativeId: payload.creativeId || payload.creative_id || null,
      score: payload.score ?? 0.92,
      issues: payload.issues || [],
      scoredAt: new Date().toISOString(),
    };
  }

  verifyViewability(payload: Record<string, any>) {
    return {
      impressionId: payload.impressionId || payload.impression_id || null,
      viewable: payload.viewable ?? true,
      vendor: payload.vendor || 'taskirx',
      verifiedAt: new Date().toISOString(),
    };
  }

  createCustomReport(payload: Record<string, any>) {
    const id = payload.reportId || payload.id || `report_${Date.now()}`;
    return {
      status: 'created',
      reportId: id,
      dimensions: payload.dimensions || [],
      metrics: payload.metrics || [],
      filters: payload.filters || {},
      createdAt: new Date().toISOString(),
    };
  }

  syncWarehouse(payload: Record<string, any>) {
    const id = payload.syncId || payload.id || `wh_${Date.now()}`;
    return {
      status: 'queued',
      syncId: id,
      provider: payload.provider || 'snowflake',
      dataset: payload.dataset || payload.schema || null,
      scheduledAt: new Date().toISOString(),
    };
  }

  registerBiTool(payload: Record<string, any>) {
    const id = payload.integrationId || payload.id || `bi_${Date.now()}`;
    return {
      status: 'connected',
      integrationId: id,
      tool: payload.tool || payload.vendor || 'tableau',
      workspace: payload.workspace || payload.site || null,
      connectedAt: new Date().toISOString(),
    };
  }

  verifyBlockchain(payload: Record<string, any>) {
    const id = payload.txId || payload.transactionId || `tx_${Date.now()}`;
    return {
      status: 'ok',
      transactionId: id,
      chain: payload.chain || 'ethereum',
      verifiedAt: new Date().toISOString(),
    };
  }

  registerSmartContract(payload: Record<string, any>) {
    const id = payload.contractId || payload.id || `sc_${Date.now()}`;
    return {
      status: 'ok',
      contractId: id,
      network: payload.network || 'polygon',
      address: payload.address || null,
      registeredAt: new Date().toISOString(),
    };
  }

  registerWeb3Wallet(payload: Record<string, any>) {
    const id = payload.walletId || payload.id || `wallet_${Date.now()}`;
    return {
      status: 'connected',
      walletId: id,
      address: payload.address || null,
      chain: payload.chain || 'ethereum',
      connectedAt: new Date().toISOString(),
    };
  }

  registerMetaverse(payload: Record<string, any>) {
    const id = payload.campaignId || payload.id || `meta_${Date.now()}`;
    return {
      status: 'ok',
      campaignId: id,
      world: payload.world || payload.platform || null,
      format: payload.format || '3d',
      createdAt: new Date().toISOString(),
    };
  }

  registerArVr(payload: Record<string, any>) {
    const id = payload.experienceId || payload.id || `arvr_${Date.now()}`;
    return {
      status: 'ok',
      experienceId: id,
      mode: payload.mode || 'ar',
      asset: payload.asset || null,
      createdAt: new Date().toISOString(),
    };
  }

  registerVoice(payload: Record<string, any>) {
    const id = payload.integrationId || payload.id || `voice_${Date.now()}`;
    return {
      status: 'ok',
      integrationId: id,
      assistant: payload.assistant || 'alexa',
      locale: payload.locale || 'en-US',
      registeredAt: new Date().toISOString(),
    };
  }

  registerEdgeCompute(payload: Record<string, any>) {
    const id = payload.edgeId || payload.id || `edge_${Date.now()}`;
    return {
      status: 'ok',
      edgeId: id,
      provider: payload.provider || 'fastly',
      region: payload.region || 'global',
      registeredAt: new Date().toISOString(),
    };
  }

  forecastRevenue(payload: Record<string, any>) {
    return {
      status: 'ok',
      horizonDays: payload.horizonDays ?? 30,
      expectedRevenue: payload.expectedRevenue ?? 125000,
      confidence: payload.confidence ?? 0.7,
      generatedAt: new Date().toISOString(),
    };
  }

  private async buildConfigValidation(
    integrationKey: string,
    payload: Record<string, any>,
    requiredFields: string[],
  ) {
    const missing = requiredFields.filter((field) => !payload?.[field]);
    const tenantId = payload?.tenantId || payload?.tenant_id || 'default';
    const status = missing.length ? 'requires-config' : 'configured';
    const providedFields = Object.keys(payload || {});

    const existing = await this.configRepository.findOne({
      where: { tenantId, integrationKey },
    });

    const record = existing
      ? this.configRepository.merge(existing, {
          status,
          configData: payload,
          missingFields: missing,
          providedFields,
        })
      : this.configRepository.create({
          tenantId,
          integrationKey,
          status,
          configData: payload,
          missingFields: missing,
          providedFields,
        });

    await this.configRepository.save(record);

    return {
      status,
      integrationKey,
      tenantId,
      missingFields: missing,
      providedFields,
      checkedAt: new Date().toISOString(),
    };
  }

  createProgrammaticGuaranteed(payload: Record<string, any>) {
    const id = payload.id || `pg_${Date.now()}`;
    return {
      status: 'created',
      id,
      buyer: payload.buyer || payload.advertiser || null,
      seller: payload.seller || payload.publisher || null,
      cpm: payload.cpm ?? payload.price ?? null,
      startsAt: payload.startsAt || payload.start || null,
      endsAt: payload.endsAt || payload.end || null,
      inventory: payload.inventory || payload.adUnits || [],
    };
  }

  createOpenDirect(payload: Record<string, any>) {
    const id = payload.id || `od_${Date.now()}`;
    return {
      status: 'created',
      id,
      buyer: payload.buyer || payload.advertiser || null,
      seller: payload.seller || payload.publisher || null,
      floor: payload.floor ?? payload.cpm ?? null,
      inventory: payload.inventory || payload.adUnits || [],
      terms: payload.terms || null,
    };
  }

  createFirstLook(payload: Record<string, any>) {
    const id = payload.id || `fl_${Date.now()}`;
    return {
      status: 'created',
      id,
      buyer: payload.buyer || payload.advertiser || null,
      seller: payload.seller || payload.publisher || null,
      priority: payload.priority ?? 1,
      inventory: payload.inventory || payload.adUnits || [],
      terms: payload.terms || null,
    };
  }

  private isUsPrivacyOptOut(usPrivacyString: string | null): boolean | null {
    if (!usPrivacyString || usPrivacyString.length < 3) {
      return null;
    }

    const signal = usPrivacyString[2]?.toUpperCase();
    if (!signal) {
      return null;
    }

    return signal === 'Y';
  }
}
