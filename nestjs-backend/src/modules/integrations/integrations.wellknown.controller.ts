import { Controller, Get, Header } from '@nestjs/common';
import { IntegrationsService } from './integrations.service';

@Controller()
export class IntegrationsWellKnownController {
  constructor(private readonly integrationsService: IntegrationsService) {}

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
}
