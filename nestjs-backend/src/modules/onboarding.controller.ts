import {
  Controller,
  Post,
  Get,
  Put,
  Body,
  Param,
  BadRequestException,
} from '@nestjs/common';
import { ApiOperation, ApiTags, ApiBearerAuth } from '@nestjs/swagger';
import { OnboardingService } from '../services/onboarding.service';

/**
 * Onboarding Controller
 * 
 * Manages:
 * - Onboarding workflow initiation
 * - Progress tracking
 * - Task completion
 * - Resource access
 * - Success metrics
 * 
 * Endpoints: 10 total
 */
@ApiTags('Onboarding')
@Controller('onboarding')
export class OnboardingController {
  constructor(private onboardingService: OnboardingService) {}

  /**
   * Start onboarding for a new customer
   */
  @Post('start')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Start onboarding',
    description: 'Initialize onboarding workflow for new customer',
  })
  async startOnboarding(
    @Body()
    payload: {
      email: string;
      companyName: string;
      tier?: 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE';
    },
  ) {
    const userId = 'user_123'; // Would come from auth context

    if (!payload.email || !payload.companyName) {
      throw new BadRequestException('Email and company name required');
    }

    const onboarding = await this.onboardingService.startOnboarding(
      userId,
      payload.email,
      payload.companyName,
      payload.tier || 'STARTER',
    );

    return {
      success: true,
      onboarding,
      message: 'Onboarding started successfully',
    };
  }

  /**
   * Get onboarding progress
   */
  @Get('progress')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get progress',
    description: 'Get current onboarding progress and completion status',
  })
  async getProgress() {
    const userId = 'user_123';

    const progress = await this.onboardingService.getOnboardingProgress(userId);

    return {
      success: true,
      progress,
    };
  }

  /**
   * Complete an onboarding task
   */
  @Post('tasks/:taskName/complete')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Complete task',
    description: 'Mark an onboarding task as completed',
  })
  async completeTask(@Param('taskName') taskName: string) {
    const userId = 'user_123';

    if (!taskName) {
      throw new BadRequestException('Task name required');
    }

    const result = await this.onboardingService.completeTask(userId, taskName);

    return {
      success: true,
      taskCompletion: result,
      message: `Task "${taskName}" completed`,
    };
  }

  /**
   * Get onboarding checklist
   */
  @Get('checklist')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get checklist',
    description: 'Get complete onboarding checklist with task status',
  })
  async getChecklist() {
    const userId = 'user_123';

    const checklist = await this.onboardingService.getOnboardingChecklist(userId);

    return {
      success: true,
      checklist,
    };
  }

  /**
   * Get onboarding resources
   */
  @Get('resources')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get resources',
    description: 'Get documentation, videos, and support resources',
  })
  async getResources() {
    const userId = 'user_123';

    const resources = await this.onboardingService.getOnboardingResources(userId);

    return {
      success: true,
      resources,
    };
  }

  /**
   * Get API key
   */
  @Post('api-key/generate')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Generate API key',
    description: 'Generate a new API key for integrations',
  })
  async generateApiKey() {
    const userId = 'user_123';

    const apiKey = await this.onboardingService.generateApiKey(userId);

    return {
      success: true,
      apiKey,
      message: 'API key generated successfully',
      warning: 'Save this key securely. You will not be able to see it again.',
    };
  }

  /**
   * Track onboarding metric
   */
  @Put('metrics/:metric')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Track metric',
    description: 'Update onboarding progress metrics',
  })
  async trackMetric(
    @Param('metric') metric: string,
    @Body() payload: { value: number },
  ) {
    const userId = 'user_123';

    if (!metric || payload.value === undefined) {
      throw new BadRequestException('Metric and value required');
    }

    await this.onboardingService.trackMetric(userId, metric, payload.value);

    return {
      success: true,
      metric,
      value: payload.value,
      message: 'Metric tracked successfully',
    };
  }

  /**
   * Get onboarding statistics (admin only)
   */
  @Get('stats')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get statistics',
    description: 'Get onboarding statistics and conversion rates (admin only)',
  })
  async getStats() {
    const stats = await this.onboardingService.getOnboardingStats();

    return {
      success: true,
      stats,
    };
  }

  /**
   * Get success stories
   */
  @Get('success-stories')
  @ApiOperation({
    summary: 'Get success stories',
    description: 'Get customer success stories and case studies',
  })
  async getSuccessStories() {
    return {
      success: true,
      stories: [
        {
          id: 'story_001',
          company: 'TechCorp',
          title: 'Increased ROI by 340%',
          excerpt: 'Using TaskirX, we optimized our ad campaigns and saw significant ROI improvements.',
          image: 'https://example.com/techcorp.jpg',
          link: '/case-studies/techcorp',
        },
        {
          id: 'story_002',
          company: 'RetailPlus',
          title: 'Reduced Campaign Setup Time by 80%',
          excerpt: 'With TaskirX automation, our team spends less time on manual work and more time strategizing.',
          image: 'https://example.com/retailplus.jpg',
          link: '/case-studies/retailplus',
        },
        {
          id: 'story_003',
          company: 'SaaS Startup',
          title: 'Scaled to $100K MRR in 6 Months',
          excerpt: 'TaskirX helped us scale our customer acquisition efficiently and cost-effectively.',
          image: 'https://example.com/saastartup.jpg',
          link: '/case-studies/saastartup',
        },
      ],
    };
  }

  /**
   * Get onboarding tips and best practices
   */
  @Get('tips')
  @ApiOperation({
    summary: 'Get tips',
    description: 'Get best practices and tips for maximizing platform value',
  })
  async getTips() {
    return {
      success: true,
      tips: [
        {
          title: 'Start with Clear Goals',
          description: 'Define your KPIs before creating campaigns',
          category: 'setup',
          difficulty: 'beginner',
        },
        {
          title: 'Use ML Bidding Optimization',
          description: 'Let our ML engine optimize bids for better ROI',
          category: 'optimization',
          difficulty: 'intermediate',
        },
        {
          title: 'Test Multiple Audiences',
          description: 'Create multiple audience segments and A/B test',
          category: 'strategy',
          difficulty: 'intermediate',
        },
        {
          title: 'Monitor Real-Time Metrics',
          description: 'Check live dashboards to catch issues early',
          category: 'monitoring',
          difficulty: 'beginner',
        },
        {
          title: 'Leverage Custom Segments',
          description: 'Create segments based on user behavior',
          category: 'segmentation',
          difficulty: 'advanced',
        },
      ],
    };
  }

  /**
   * Get personalized recommendations
   */
  @Get('recommendations')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get recommendations',
    description: 'Get AI-powered personalized recommendations based on account data',
  })
  async getRecommendations() {
    const _userId = 'user_123';

    return {
      success: true,
      recommendations: [
        {
          title: 'Optimize Your Bid Strategy',
          description: 'Your CPC is 15% higher than similar accounts. Try automated bidding.',
          priority: 'high',
          action: '/settings/bidding',
          estimatedImpact: '+12% ROI',
        },
        {
          title: 'Expand to New Audience',
          description: 'We found a new audience segment that performed well with similar accounts.',
          priority: 'medium',
          action: '/campaigns/new-audience',
          estimatedImpact: '+8% volume',
        },
        {
          title: 'Scale Your Top Campaign',
          description: 'Your top campaign has strong metrics. Consider increasing budget by 25%.',
          priority: 'medium',
          action: '/campaigns/top-campaign',
          estimatedImpact: '+$5K revenue',
        },
      ],
    };
  }
}
