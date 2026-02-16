import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { DataSourceOptions } from 'typeorm';
import { newDb } from 'pg-mem';
import { randomUUID } from 'crypto';
import { AppController } from '../src/app.controller';
import { HealthModule } from '../src/modules/health/health.module';
import { SspModule } from '../src/modules/ssp/ssp.module';
import { DspModule } from '../src/modules/dsp/dsp.module';

@Module({
  imports: [
    TypeOrmModule.forRootAsync({
      useFactory: async () => ({
        type: 'postgres',
        synchronize: true,
        entities: [__dirname + '/../src/**/*.entity{.ts,.js}'],
      }),
      dataSourceFactory: async (options) => {
        const db = newDb({ autoCreateForeignKeyIndices: true });
        db.public.registerFunction({
          name: 'current_database',
          implementation: () => 'test',
        });
        db.public.registerFunction({
          name: 'version',
          implementation: () => 'pg-mem',
        });
        db.public.registerFunction({
          name: 'uuid_generate_v4',
          implementation: () => randomUUID(),
        });
        const dataSource = await db.adapters.createTypeormDataSource(
          options as DataSourceOptions,
        );
        return dataSource.initialize();
      },
    }),
    SspModule,
    HealthModule,
    DspModule,
  ],
  controllers: [AppController],
})
export class TestAppModule {}
