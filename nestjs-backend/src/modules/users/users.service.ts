import { Injectable, ConflictException, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { User } from './user.entity';
import * as bcrypt from 'bcryptjs';
import { v4 as uuidv4 } from 'uuid';

@Injectable()
export class UsersService {
  constructor(
    @InjectRepository(User)
    private usersRepository: Repository<User>,
  ) {}

  async create(createUserDto: {
    email: string;
    password: string;
    role: string;
    companyName?: string;
    tenantId?: string;
  }): Promise<User> {
    // Check if user already exists
    const existingUser = await this.usersRepository.findOne({
      where: { email: createUserDto.email },
    });

    if (existingUser) {
      throw new ConflictException('User with this email already exists');
    }

    // Hash password
    const passwordHash = await bcrypt.hash(createUserDto.password, 10);

    // Create user
    const user = this.usersRepository.create({
      email: createUserDto.email,
      passwordHash,
      role: createUserDto.role as any,
      companyName: createUserDto.companyName,
      tenantId: createUserDto.tenantId || uuidv4(), // Generate tenant ID if not provided
    });

    return this.usersRepository.save(user);
  }

  async findByEmail(email: string): Promise<User | null> {
    return this.usersRepository.findOne({ where: { email } });
  }

  async findById(id: string): Promise<User> {
    const user = await this.usersRepository.findOne({ where: { id } });
    if (!user) {
      throw new NotFoundException('User not found');
    }
    return user;
  }

  async findByTenant(tenantId: string): Promise<User[]> {
    return this.usersRepository.find({ where: { tenantId } });
  }

  async findByRole(role: string): Promise<User[]> {
    return this.usersRepository.find({ where: { role: role as any } });
  }

  async validatePassword(password: string, passwordHash: string): Promise<boolean> {
    return bcrypt.compare(password, passwordHash);
  }

  async updateLastLogin(userId: string): Promise<void> {
    await this.usersRepository.update(userId, { lastLoginAt: new Date() });
  }

  async update(id: string, updateData: Partial<User>): Promise<User> {
    await this.usersRepository.update(id, updateData);
    return this.findById(id);
  }

  async remove(id: string): Promise<void> {
    const user = await this.findById(id);
    await this.usersRepository.remove(user);
  }
}
