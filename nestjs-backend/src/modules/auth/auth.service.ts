import { Injectable } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { UsersService } from '../users/users.service';
import { UserRole } from '../users/user.entity';

@Injectable()
export class AuthService {
  constructor(
    private usersService: UsersService,
    private jwtService: JwtService,
  ) {}

  async validateUser(email: string, password: string): Promise<any> {
    const user = await this.usersService.findByEmail(email);
    if (!user) {
      return null;
    }

    const isPasswordValid = await this.usersService.validatePassword(password, user.passwordHash);
    if (!isPasswordValid) {
      return null;
    }

    await this.usersService.updateLastLogin(user.id);
  const { passwordHash: _passwordHash, ...result } = user;
    return result;
  }

  async login(user: any) {
    const payload = { email: user.email, sub: user.id, role: user.role, tenantId: user.tenantId };
    return {
      access_token: this.jwtService.sign(payload),
      user: {
        id: user.id,
        email: user.email,
        role: user.role,
        tenantId: user.tenantId,
        companyName: user.companyName,
      },
    };
  }

  async register(email: string, password: string, role: UserRole, companyName?: string) {
    const user = await this.usersService.create({
      email,
      password,
      role,
      companyName,
    });

  const { passwordHash: _passwordHash, ...result } = user;
    return this.login(result);
  }

  async refreshToken(userId: string) {
    const user = await this.usersService.findById(userId);
    const payload = { email: user.email, sub: user.id, role: user.role, tenantId: user.tenantId };
    
    return {
      access_token: this.jwtService.sign(payload),
    };
  }
}
