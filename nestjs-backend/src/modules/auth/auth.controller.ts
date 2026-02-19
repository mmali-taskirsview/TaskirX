import { Controller, Post, Body, UnauthorizedException, UseGuards, Request } from '@nestjs/common';
import { ApiTags, ApiOperation } from '@nestjs/swagger';
import { Throttle } from '@nestjs/throttler'; // Import Throttle
import { AuthService } from './auth.service';
import { JwtAuthGuard } from './guards/jwt-auth.guard';

@ApiTags('auth')
@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Throttle({ default: { limit: 5, ttl: 60000 } }) // Limit login attempts
  @Post('login')
  @ApiOperation({ summary: 'Login with email and password' })
  async login(@Body() loginDto: { email: string; password: string }) {
    const user = await this.authService.validateUser(loginDto.email, loginDto.password);
    if (!user) {
      throw new UnauthorizedException('Invalid credentials');
    }
    return this.authService.login(user);
  }

  @Post('register')
  @ApiOperation({ summary: 'Register new user' })
  async register(
    @Body() registerDto: { email: string; password: string; role: string; companyName?: string },
  ) {
    return this.authService.register(
      registerDto.email,
      registerDto.password,
      registerDto.role as any,
      registerDto.companyName,
    );
  }

  @Post('refresh')
  @UseGuards(JwtAuthGuard)
  @ApiOperation({ summary: 'Refresh access token' })
  async refresh(@Request() req) {
    return this.authService.refreshToken(req.user.id);
  }
}
