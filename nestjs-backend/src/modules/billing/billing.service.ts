import { Injectable, NotFoundException, BadRequestException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, DataSource } from 'typeorm';
import { Wallet } from './wallet.entity';
import { Transaction, TransactionType, TransactionStatus } from './transaction.entity';

@Injectable()
export class BillingService {
  constructor(
    @InjectRepository(Wallet)
    private walletsRepository: Repository<Wallet>,
    @InjectRepository(Transaction)
    private transactionsRepository: Repository<Transaction>,
    private dataSource: DataSource,
  ) {}

  async createWallet(userId: string, tenantId: string, currency: string = 'USD'): Promise<Wallet> {
    const existingWallet = await this.walletsRepository.findOne({ where: { userId } });
    if (existingWallet) {
      return existingWallet;
    }

    const wallet = this.walletsRepository.create({
      userId,
      tenantId,
      currency,
      balance: 0,
      totalDeposited: 0,
      totalSpent: 0,
    });

    return this.walletsRepository.save(wallet);
  }

  async getWallet(userId: string): Promise<Wallet> {
    const wallet = await this.walletsRepository.findOne({ where: { userId } });
    if (!wallet) {
      throw new NotFoundException('Wallet not found');
    }
    return wallet;
  }

  async deposit(
    userId: string,
    tenantId: string,
    amount: number,
    description?: string,
  ): Promise<Transaction> {
    if (amount <= 0) {
      throw new BadRequestException('Amount must be greater than 0');
    }

    const queryRunner = this.dataSource.createQueryRunner();
    await queryRunner.connect();
    await queryRunner.startTransaction();

    try {
      let wallet = await queryRunner.manager.findOne(Wallet, { where: { userId } });
      
      if (!wallet) {
        wallet = await queryRunner.manager.save(Wallet, {
          userId,
          tenantId,
          balance: 0,
          totalDeposited: 0,
          totalSpent: 0,
          currency: 'USD',
        });
      }

      const balanceBefore = Number(wallet.balance);
      const balanceAfter = balanceBefore + amount;

      await queryRunner.manager.update(Wallet, { id: wallet.id }, {
        balance: balanceAfter,
        totalDeposited: Number(wallet.totalDeposited) + amount,
      });

      const transaction = await queryRunner.manager.save(Transaction, {
        walletId: wallet.id,
        userId,
        tenantId,
        type: TransactionType.DEPOSIT,
        status: TransactionStatus.COMPLETED,
        amount,
        balanceBefore,
        balanceAfter,
        currency: wallet.currency,
        description,
      });

      await queryRunner.commitTransaction();
      return transaction;
    } catch (error) {
      await queryRunner.rollbackTransaction();
      throw error;
    } finally {
      await queryRunner.release();
    }
  }

  async charge(
    userId: string,
    amount: number,
    campaignId: string,
    description?: string,
  ): Promise<Transaction> {
    if (amount <= 0) {
      throw new BadRequestException('Amount must be greater than 0');
    }

    const queryRunner = this.dataSource.createQueryRunner();
    await queryRunner.connect();
    await queryRunner.startTransaction();

    try {
      const wallet = await queryRunner.manager.findOne(Wallet, { where: { userId } });
      
      if (!wallet) {
        throw new NotFoundException('Wallet not found');
      }

      const balanceBefore = Number(wallet.balance);
      if (balanceBefore < amount) {
        throw new BadRequestException('Insufficient balance');
      }

      const balanceAfter = balanceBefore - amount;

      await queryRunner.manager.update(Wallet, { id: wallet.id }, {
        balance: balanceAfter,
        totalSpent: Number(wallet.totalSpent) + amount,
      });

      const transaction = await queryRunner.manager.save(Transaction, {
        walletId: wallet.id,
        userId,
        tenantId: wallet.tenantId,
        type: TransactionType.CAMPAIGN_CHARGE,
        status: TransactionStatus.COMPLETED,
        amount,
        balanceBefore,
        balanceAfter,
        currency: wallet.currency,
        relatedId: campaignId,
        description,
      });

      await queryRunner.commitTransaction();
      return transaction;
    } catch (error) {
      await queryRunner.rollbackTransaction();
      throw error;
    } finally {
      await queryRunner.release();
    }
  }

  async getTransactions(userId: string, limit: number = 50): Promise<Transaction[]> {
    return this.transactionsRepository.find({
      where: { userId },
      order: { createdAt: 'DESC' },
      take: limit,
    });
  }

  async getBalance(userId: string): Promise<number> {
    const wallet = await this.getWallet(userId);
    return Number(wallet.balance);
  }
}
