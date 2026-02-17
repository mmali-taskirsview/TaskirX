import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import { AppModule } from './app.module';
import helmet from 'helmet';
import compression from 'compression';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, {
    logger: ['error', 'warn', 'log', 'debug', 'verbose'],
  });

  // Security
  app.use(helmet());
  app.enableCors({
    origin: process.env.NODE_ENV === 'production' 
      ? [
          'https://dashboard.taskirx.com', 
          'https://api.taskirx.com',
          /\.taskirx\.com$/ // Allow all subdomains
        ]
      : true,
    credentials: true,
  });

  // Compression with optimized settings for performance
  // Threshold: compress responses > 1KB
  // Level: 6 (good balance between speed and compression)
  // Filter: compress all responses except those with specific headers
  app.use(compression({
    threshold: 1024,           // Only compress responses > 1KB
    level: 6,                  // Compression level 1-9 (6 is good balance)
    filter: (req, res) => {
      // Don't compress responses with this request header
      if (req.headers['x-no-compression']) {
        return false;
      }
      // Use compression filter function
      return compression.filter(req, res);
    },
    chunkSize: 64 * 1024,      // 64KB chunks
  }));

  // Global prefix
  app.setGlobalPrefix(process.env.API_PREFIX || 'api', {
    exclude: ['/', '/ads.txt', '/app-ads.txt', '/sellers.json'],
  });

  // Global validation pipe
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      forbidNonWhitelisted: true,
      transform: true,
      transformOptions: {
        enableImplicitConversion: true,
      },
    }),
  );

  // Swagger documentation
  const config = new DocumentBuilder()
    .setTitle('TaskirX v3 API')
    .setDescription('High-Performance Ad Exchange Platform API')
    .setVersion('3.0.0')
    .addBearerAuth()
    .addTag('auth', 'Authentication endpoints')
    .addTag('users', 'User management')
    .addTag('campaigns', 'Campaign management')
    .addTag('billing', 'Billing and wallet operations')
    .addTag('analytics', 'Analytics and reporting')
    .addTag('ai-agents', 'AI agent orchestration')
    .build();

  const document = SwaggerModule.createDocument(app, config);
  SwaggerModule.setup('api/docs', app, document);

  const port = process.env.PORT || 4000;
  await app.listen(port);

  console.log(`
  🚀 TaskirX v3 NestJS Backend Started!
  ====================================
  📡 API: http://localhost:${port}/${process.env.API_PREFIX || 'api'}
  📚 Swagger Docs: http://localhost:${port}/api/docs
  🔧 Environment: ${process.env.NODE_ENV || 'development'}
  ====================================
  `);
}

bootstrap();
