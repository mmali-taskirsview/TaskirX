import { Controller, Get, Res } from '@nestjs/common';
import { ApiTags, ApiOperation } from '@nestjs/swagger';
import { Response } from 'express';

@Controller()
export class AppController {
  @Get()
  @ApiTags('Root')
  @ApiOperation({ summary: 'Redirect to API Docs or Frontend' })
  root(@Res() res: Response) {
    // Check if it's a browser asking
    return res.json({
      message: 'TaskirX V3 API is Running',
      documentation: '/api/docs',
      frontend: 'http://localhost:3000',
      status: 'active'
    });
  }
}
