import {
  Body,
  Controller,
  Delete,
  Get,
  Header,
  HttpCode,
  HttpStatus,
  Param,
  Post,
  Put,
  Query,
  Request,
  UseGuards,
} from '@nestjs/common';
import { IntegrationsService } from './integrations.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';

@Controller('integrations')
export class IntegrationsController {
  constructor(private readonly integrationsService: IntegrationsService) {}

  @Get('catalog')
  getCatalog() {
    return this.integrationsService.getCatalog();
  }

  @Get('ads.txt')
  @Header('Content-Type', 'text/plain')
  getAdsTxt() {
    return this.integrationsService.getAdsTxt();
  }

  @Get('app-ads.txt')
  @Header('Content-Type', 'text/plain')
  getAppAdsTxt() {
    return this.integrationsService.getAppAdsTxt();
  }

  @Get('sellers.json')
  getSellersJson() {
    return this.integrationsService.getSellersJson();
  }

  @Post('consent/tcf')
  @HttpCode(HttpStatus.OK)
  normalizeTcf(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      normalized: this.integrationsService.normalizeTcf(payload),
    };
  }

  @Post('consent/gpp')
  @HttpCode(HttpStatus.OK)
  normalizeGpp(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      normalized: this.integrationsService.normalizeGpp(payload),
    };
  }

  @Post('consent/ccpa')
  @HttpCode(HttpStatus.OK)
  normalizeCcpa(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      normalized: this.integrationsService.normalizeCcpa(payload),
    };
  }

  @Post('consent/gdpr')
  @HttpCode(HttpStatus.OK)
  normalizeGdpr(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      normalized: this.integrationsService.normalizeGdpr(payload),
    };
  }

  @Post('consent/dns')
  @HttpCode(HttpStatus.OK)
  normalizeDoNotSell(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      normalized: this.integrationsService.normalizeDoNotSell(payload),
    };
  }

  @Post('consent/cmp')
  @HttpCode(HttpStatus.OK)
  normalizeCmp(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      normalized: this.integrationsService.normalizeCmp(payload),
    };
  }

  @Post('privacy/deletion')
  @HttpCode(HttpStatus.ACCEPTED)
  requestDataDeletion(@Body() payload: Record<string, any>) {
    return this.integrationsService.requestDataDeletion(payload);
  }

  @Post('privacy/sandbox/topics')
  @HttpCode(HttpStatus.OK)
  getPrivacySandboxTopics(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      topics: this.integrationsService.getPrivacySandboxTopics(payload),
    };
  }

  @Post('privacy/sandbox/attribution')
  @HttpCode(HttpStatus.ACCEPTED)
  submitPrivacySandboxAttribution(@Body() payload: Record<string, any>) {
    return this.integrationsService.submitPrivacySandboxAttribution(payload);
  }

  @Post('identity/sync')
  @HttpCode(HttpStatus.OK)
  normalizeIdentity(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      ids: this.integrationsService.normalizeIdentity(payload),
    };
  }

  @Post('identity/redirect-sync')
  @HttpCode(HttpStatus.OK)
  redirectSync(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerRedirectSync(payload);
  }

  @Post('identity/iframe-sync')
  @HttpCode(HttpStatus.OK)
  iframeSync(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerIframeSync(payload);
  }

  @Post('identity/sharedid')
  @HttpCode(HttpStatus.OK)
  registerSharedId(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerSharedId(payload);
  }

  @Post('identity/pubcid')
  @HttpCode(HttpStatus.OK)
  registerPubcid(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerPubcid(payload);
  }

  @Post('config/google-rtb')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateGoogleRtb(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateGoogleRtbConfig({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/facebook-bidding')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateFacebookBidding(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateFacebookBiddingConfig({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/amazon-tam')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateAmazonTam(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateAmazonTamConfig({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/uid2')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateUid2(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateUid2Config({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/id5')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateId5(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateId5Config({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/liveramp')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateLiveRamp(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateLiveRampConfig({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/ttd-uid')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validateTtdUid(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validateTtdUidConfig({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Post('config/ppid')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  validatePpid(@Request() req, @Body() payload: Record<string, any>) {
    return this.integrationsService.validatePpidConfig({
      ...payload,
      tenantId: payload.tenantId || req.user?.tenantId,
    });
  }

  @Get('config')
  @UseGuards(JwtAuthGuard)
  listConfig(@Request() req, @Query('tenantId') tenantId?: string) {
    return this.integrationsService.listConfigValidations(tenantId || req.user?.tenantId);
  }

  @Get('config/:integrationKey')
  @UseGuards(JwtAuthGuard)
  getConfig(
    @Request() req,
    @Param('integrationKey') integrationKey: string,
    @Query('tenantId') tenantId?: string,
  ) {
    return this.integrationsService.getConfigValidation(
      integrationKey,
      tenantId || req.user?.tenantId,
    );
  }

  @Put('config/:integrationKey')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  updateConfig(
    @Request() req,
    @Param('integrationKey') integrationKey: string,
    @Body() payload: Record<string, any>,
    @Query('tenantId') tenantId?: string,
  ) {
    return this.integrationsService.updateConfigValidation(
      integrationKey,
      payload,
      tenantId || req.user?.tenantId,
    );
  }

  @Delete('config/:integrationKey')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  deleteConfig(
    @Request() req,
    @Param('integrationKey') integrationKey: string,
    @Query('tenantId') tenantId?: string,
  ) {
    return this.integrationsService.deleteConfigValidation(
      integrationKey,
      tenantId || req.user?.tenantId,
    );
  }

  @Post('transfer/batch')
  @HttpCode(HttpStatus.ACCEPTED)
  acceptBatch(@Body() payload: Record<string, any>) {
    return {
      status: 'accepted',
      jobId: `job_${Date.now()}`,
      received: payload,
    };
  }

  @Post('transfer/ftp')
  @HttpCode(HttpStatus.ACCEPTED)
  registerFtpTransfer(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerFtpTransfer(payload);
  }

  @Post('transfer/cloud-storage')
  @HttpCode(HttpStatus.ACCEPTED)
  registerCloudStorage(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerCloudStorageSync(payload);
  }

  @Post('transfer/snowflake')
  @HttpCode(HttpStatus.CREATED)
  registerSnowflake(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerSnowflakeShare(payload);
  }

  @Post('transfer/bigquery')
  @HttpCode(HttpStatus.ACCEPTED)
  registerBigQuery(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerBigQueryTransfer(payload);
  }

  @Post('transfer/stream')
  @HttpCode(HttpStatus.ACCEPTED)
  acceptStream(@Body() payload: Record<string, any>) {
    return {
      status: 'accepted',
      streamId: `stream_${Date.now()}`,
      received: payload,
    };
  }

  @Post('api/graphql')
  @HttpCode(HttpStatus.OK)
  handleGraphql(@Body() payload: Record<string, any>) {
    return this.integrationsService.handleGraphql(payload);
  }

  @Post('api/grpc')
  @HttpCode(HttpStatus.OK)
  handleGrpc(@Body() payload: Record<string, any>) {
    return this.integrationsService.handleGrpc(payload);
  }

  @Post('api/soap')
  @HttpCode(HttpStatus.OK)
  handleSoap(@Body() payload: Record<string, any>) {
    return this.integrationsService.handleSoap(payload);
  }

  @Post('webhooks')
  @HttpCode(HttpStatus.CREATED)
  registerWebhook(@Body() payload: { url: string; events?: string[] }) {
    const events = payload.events?.length ? payload.events : ['auction.win', 'conversion'];
    return this.integrationsService.registerWebhook(payload.url, events);
  }

  @Get('webhooks')
  listWebhooks(): Array<Record<string, any>> {
    return this.integrationsService.listWebhooks();
  }

  @Get('creative/standards')
  getCreativeStandards() {
    return {
      standards: ['VAST 4.2', 'VPAID 2.1', 'VMAP 1.0'],
      simid: 'requires-player',
    };
  }

  @Post('creative/upload')
  @HttpCode(HttpStatus.CREATED)
  uploadCreative(@Body() payload: Record<string, any>) {
    return this.integrationsService.createCreative(payload);
  }

  @Post('creative/dynamic')
  @HttpCode(HttpStatus.ACCEPTED)
  submitDynamicCreative(@Body() payload: Record<string, any>) {
    return this.integrationsService.submitDynamicCreative(payload);
  }

  @Post('creative/native')
  @HttpCode(HttpStatus.OK)
  createNativeTemplate(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      template: this.integrationsService.createNativeTemplate(payload),
    };
  }

  @Post('creative/rich-media')
  @HttpCode(HttpStatus.CREATED)
  createRichMedia(@Body() payload: Record<string, any>) {
    return this.integrationsService.createRichMedia(payload);
  }

  @Post('creative/html5')
  @HttpCode(HttpStatus.CREATED)
  createHtml5Creative(@Body() payload: Record<string, any>) {
    return this.integrationsService.createHtml5Creative(payload);
  }

  @Post('creative/third-party')
  @HttpCode(HttpStatus.CREATED)
  registerThirdPartyCreative(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerThirdPartyCreative(payload);
  }

  @Post('creative/review')
  @HttpCode(HttpStatus.ACCEPTED)
  submitCreativeReview(@Body() payload: Record<string, any>) {
    return this.integrationsService.submitCreativeReview(payload);
  }

  @Post('creative/verification')
  @HttpCode(HttpStatus.OK)
  verifyCreative(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      verification: this.integrationsService.verifyCreative(payload),
    };
  }

  @Post('bidding/simid')
  @HttpCode(HttpStatus.OK)
  registerSimid(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerSimid(payload);
  }

  @Post('bidding/om-sdk')
  @HttpCode(HttpStatus.OK)
  registerOmSdk(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerOmSdk(payload);
  }

  @Post('bidding/seller-audiences')
  @HttpCode(HttpStatus.OK)
  registerSellerAudiences(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerSellerAudiences(payload);
  }

  @Post('serving/gpt')
  @HttpCode(HttpStatus.OK)
  registerGpt(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerGptTag(payload);
  }

  @Post('serving/ssai')
  @HttpCode(HttpStatus.OK)
  registerSsai(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerSsai(payload);
  }

  @Post('serving/dco')
  @HttpCode(HttpStatus.OK)
  registerDco(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerDco(payload);
  }

  @Post('serving/amp')
  @HttpCode(HttpStatus.OK)
  registerAmp(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerAmpDelivery(payload);
  }

  @Post('billing/tax')
  @HttpCode(HttpStatus.OK)
  calculateTax(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      tax: this.integrationsService.calculateTax(payload),
    };
  }

  @Post('billing/revenue-share')
  @HttpCode(HttpStatus.OK)
  calculateRevenueShare(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      revenueShare: this.integrationsService.calculateRevenueShare(payload),
    };
  }

  @Post('billing/chargeback')
  @HttpCode(HttpStatus.ACCEPTED)
  createChargeback(@Body() payload: Record<string, any>) {
    return this.integrationsService.createChargeback(payload);
  }

  @Post('billing/crypto')
  @HttpCode(HttpStatus.ACCEPTED)
  createCryptoPayment(@Body() payload: Record<string, any>) {
    return this.integrationsService.createCryptoPayment(payload);
  }

  @Post('billing/escrow')
  @HttpCode(HttpStatus.CREATED)
  createEscrow(@Body() payload: Record<string, any>) {
    return this.integrationsService.createEscrow(payload);
  }

  @Get('creative/dynamic')
  listDynamicCreative() {
    return {
      feeds: this.integrationsService.listDynamicCreatives(),
    };
  }

  @Get('creative')
  listCreatives() {
    return {
      creatives: this.integrationsService.listCreatives(),
    };
  }

  @Post('measurement/viewability')
  @HttpCode(HttpStatus.OK)
  measureViewability(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      measurement: this.integrationsService.measureViewability(payload),
    };
  }

  @Post('measurement/js-tracker')
  @HttpCode(HttpStatus.OK)
  measureJsTracker(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      tracker: this.integrationsService.measureJsTracker(payload),
    };
  }

  @Post('measurement/attribution')
  @HttpCode(HttpStatus.OK)
  measureAttribution(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      attribution: this.integrationsService.measureAttribution(payload),
    };
  }

  @Post('measurement/multi-touch')
  @HttpCode(HttpStatus.OK)
  measureMultiTouch(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      attribution: this.integrationsService.measureMultiTouch(payload),
    };
  }

  @Post('measurement/cross-device')
  @HttpCode(HttpStatus.OK)
  measureCrossDevice(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      mapping: this.integrationsService.measureCrossDevice(payload),
    };
  }

  @Post('measurement/offline-conversion')
  @HttpCode(HttpStatus.OK)
  measureOfflineConversion(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      conversion: this.integrationsService.measureOfflineConversion(payload),
    };
  }

  @Post('quality/ivt-filter')
  @HttpCode(HttpStatus.OK)
  evaluateIvt(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      result: this.integrationsService.evaluateIvt(payload),
    };
  }

  @Post('quality/content-verification')
  @HttpCode(HttpStatus.OK)
  verifyContent(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      verification: this.integrationsService.verifyContent(payload),
    };
  }

  @Post('quality/malware-scan')
  @HttpCode(HttpStatus.OK)
  scanMalware(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      scan: this.integrationsService.scanMalware(payload),
    };
  }

  @Post('quality/ad-quality')
  @HttpCode(HttpStatus.OK)
  scoreAdQuality(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      quality: this.integrationsService.scoreAdQuality(payload),
    };
  }

  @Post('quality/viewability-verification')
  @HttpCode(HttpStatus.OK)
  verifyViewability(@Body() payload: Record<string, any>) {
    return {
      status: 'ok',
      verification: this.integrationsService.verifyViewability(payload),
    };
  }

  @Post('mobile/sdk')
  @HttpCode(HttpStatus.OK)
  registerMobileSdk(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerMobileSdk(payload);
  }

  @Post('mobile/in-app-bidding')
  @HttpCode(HttpStatus.OK)
  registerInAppBidding(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerInAppBidding(payload);
  }

  @Post('mobile/rewarded-video')
  @HttpCode(HttpStatus.OK)
  registerRewardedVideo(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerRewardedVideo(payload);
  }

  @Post('mobile/ctv-ssai')
  @HttpCode(HttpStatus.OK)
  registerCtvSsai(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerCtvSsai(payload);
  }

  @Post('mobile/ott')
  @HttpCode(HttpStatus.OK)
  registerOtt(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerOtt(payload);
  }

  @Post('mobile/web')
  @HttpCode(HttpStatus.OK)
  registerMobileWeb(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerMobileWeb(payload);
  }

  @Post('mobile/deep-link')
  @HttpCode(HttpStatus.OK)
  registerDeepLink(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerDeepLink(payload);
  }

  @Post('mobile/app-store-attribution')
  @HttpCode(HttpStatus.OK)
  registerAppStoreAttribution(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerAppStoreAttribution(payload);
  }

  @Post('analytics/custom-report')
  @HttpCode(HttpStatus.CREATED)
  createCustomReport(@Body() payload: Record<string, any>) {
    return this.integrationsService.createCustomReport(payload);
  }

  @Post('analytics/warehouse-sync')
  @HttpCode(HttpStatus.ACCEPTED)
  syncWarehouse(@Body() payload: Record<string, any>) {
    return this.integrationsService.syncWarehouse(payload);
  }

  @Post('analytics/bi-tools')
  @HttpCode(HttpStatus.OK)
  registerBiTool(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerBiTool(payload);
  }

  @Post('analytics/forecasting')
  @HttpCode(HttpStatus.OK)
  forecastRevenue(@Body() payload: Record<string, any>) {
    return this.integrationsService.forecastRevenue(payload);
  }

  @Post('infrastructure/service-mesh')
  @HttpCode(HttpStatus.OK)
  registerServiceMesh(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerServiceMesh(payload);
  }

  @Post('infrastructure/load-balancer')
  @HttpCode(HttpStatus.OK)
  registerLoadBalancer(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerLoadBalancer(payload);
  }

  @Post('infrastructure/cdn')
  @HttpCode(HttpStatus.OK)
  registerCdn(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerCdn(payload);
  }

  @Post('infrastructure/edge-network')
  @HttpCode(HttpStatus.OK)
  registerEdgeNetwork(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerEdgeNetwork(payload);
  }

  @Post('infrastructure/db-sync')
  @HttpCode(HttpStatus.ACCEPTED)
  registerDbSync(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerDbSync(payload);
  }

  @Post('emerging/blockchain')
  @HttpCode(HttpStatus.OK)
  verifyBlockchain(@Body() payload: Record<string, any>) {
    return this.integrationsService.verifyBlockchain(payload);
  }

  @Post('emerging/smart-contracts')
  @HttpCode(HttpStatus.OK)
  registerSmartContract(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerSmartContract(payload);
  }

  @Post('emerging/web3-wallet')
  @HttpCode(HttpStatus.OK)
  registerWeb3Wallet(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerWeb3Wallet(payload);
  }

  @Post('emerging/metaverse')
  @HttpCode(HttpStatus.OK)
  registerMetaverse(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerMetaverse(payload);
  }

  @Post('emerging/ar-vr')
  @HttpCode(HttpStatus.OK)
  registerArVr(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerArVr(payload);
  }

  @Post('emerging/voice')
  @HttpCode(HttpStatus.OK)
  registerVoice(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerVoice(payload);
  }

  @Post('emerging/edge')
  @HttpCode(HttpStatus.OK)
  registerEdgeCompute(@Body() payload: Record<string, any>) {
    return this.integrationsService.registerEdgeCompute(payload);
  }

  @Post('programmatic/guaranteed')
  @HttpCode(HttpStatus.CREATED)
  createProgrammaticGuaranteed(@Body() payload: Record<string, any>) {
    return this.integrationsService.createProgrammaticGuaranteed(payload);
  }

  @Post('programmatic/opendirect')
  @HttpCode(HttpStatus.CREATED)
  createOpenDirect(@Body() payload: Record<string, any>) {
    return this.integrationsService.createOpenDirect(payload);
  }

  @Post('programmatic/first-look')
  @HttpCode(HttpStatus.CREATED)
  createFirstLook(@Body() payload: Record<string, any>) {
    return this.integrationsService.createFirstLook(payload);
  }
}
