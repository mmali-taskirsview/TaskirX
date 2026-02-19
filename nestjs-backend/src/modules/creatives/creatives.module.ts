import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { MulterModule } from '@nestjs/platform-express';
import { CreativesService } from './creatives.service';
import { CreativesController } from './creatives.controller';
import { Creative } from './creative.entity';
import * as multer from 'multer';
import * as path from 'path';

@Module({
  imports: [
    TypeOrmModule.forFeature([Creative]),
    MulterModule.register({
      dest: './uploads/creatives',
      fileFilter: (req, file, callback) => {
        // Accept images, videos, HTML files, and ZIP files
        if (
          file.mimetype.startsWith('image/') ||
          file.mimetype.startsWith('video/') ||
          file.mimetype === 'text/html' ||
          file.mimetype === 'application/zip' ||
          file.originalname.endsWith('.html') ||
          file.originalname.endsWith('.zip')
        ) {
          callback(null, true);
        } else {
          callback(new Error('Invalid file type'), false);
        }
      },
      limits: {
        fileSize: 10 * 1024 * 1024, // 10MB limit
      },
    }),
  ],
  controllers: [CreativesController],
  providers: [CreativesService],
  exports: [CreativesService],
})
export class CreativesModule {}