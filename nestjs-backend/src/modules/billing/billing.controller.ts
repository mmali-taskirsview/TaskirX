import {
  Controller,
  Get,
  Post,
  Body,
  Query,
  UseGuards,
  Request,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
import { BillingService } from './billing.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';

@ApiTags('billing')
@Controller('billing')
@UseGuards(JwtAuthGuard)
@ApiBearerAuth()
export class BillingController {
  constructor(private readonly billingService: BillingService) {}

  @Get('wallet')
  @ApiOperation({ summary: 'Get user wallet' })
  async getWallet(@Request() req) {
    await this.billingService.createWallet(req.user.id, req.user.tenantId);
    return this.billingService.getWallet(req.user.id);
  }

  @Get('balance')
  @ApiOperation({ summary: 'Get wallet balance' })
  async getBalance(@Request() req) {
    await this.billingService.createWallet(req.user.id, req.user.tenantId);
    const balance = await this.billingService.getBalance(req.user.id);
    return { balance };
  }

  @Post('deposit')
  @ApiOperation({ summary: 'Deposit funds to wallet' })
  async deposit(
    @Body() depositDto: { amount: number; description?: string },
    @Request() req,
  ) {
    return this.billingService.deposit(
      req.user.id,
      req.user.tenantId,
      depositDto.amount,
      depositDto.description,
    );
  }

  @Get('transactions')
  @ApiOperation({ summary: 'Get transaction history' })
  async getTransactions(@Request() req, @Query('limit') limit?: number) {
    return this.billingService.getTransactions(req.user.id, limit || 50);
  }
}
