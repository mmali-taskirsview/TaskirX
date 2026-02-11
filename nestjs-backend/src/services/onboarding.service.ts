import { Injectable, Logger, BadRequestException, Inject } from '@nestjs/common';
import { Redis } from 'ioredis';

/**
 * Customer Onboarding Service
 * 
 * Automated onboarding workflow with:
 * - Welcome email sequences
 * - Progressive capability unlocking
 * - API key management
 * - Getting started guides
 * - Progress tracking
 * - Success metrics
 * 
 * Onboarding Stages:
 * 1. Welcome & Setup (Day 0)
 * 2. First Campaign (Day 1)
 * 3. Integration (Day 2-3)
 * 4. Live Launch (Day 4-5)
 * 5. Optimization (Day 6-7)
 * 6. Success (Day 8+)
 */
@Injectable()
export class OnboardingService {
  private readonly logger = new Logger(OnboardingService.name);

  private readonly ONBOARDING_STAGES = {
    WELCOME: {
      stage: 1,
      name: 'Welcome & Setup',
      duration: 1, // days
      emails: [
        'welcome',
        'platform_tour',
        'documentation_links',
      ],
      tasks: [
        'Complete profile',
        'Verify email',
        'Set up first campaign',
      ],
      unlocks: ['campaign_creation', 'basic_analytics'],
    },
    FIRST_CAMPAIGN: {
      stage: 2,
      name: 'First Campaign',
      duration: 1,
      emails: ['first_campaign_guide', 'campaign_tips'],
      tasks: [
        'Create first campaign',
        'Add targeting',
        'Set budget',
      ],
      unlocks: ['advanced_targeting', 'bidding_optimization'],
    },
    INTEGRATION: {
      stage: 3,
      name: 'Integration & SDK Setup',
      duration: 2,
      emails: ['sdk_integration', 'sdk_examples'],
      tasks: [
        'Integrate tracking SDK',
        'Test tracking',
        'Verify conversions',
      ],
      unlocks: ['conversion_tracking', 'advanced_analytics'],
    },
    LIVE_LAUNCH: {
      stage: 4,
      name: 'Live Launch',
      duration: 2,
      emails: ['launch_checklist', 'launch_support'],
      tasks: [
        'Launch campaign',
        'Monitor performance',
        'Optimize bids',
      ],
      unlocks: ['real_time_dashboard', 'ml_optimization'],
    },
    OPTIMIZATION: {
      stage: 5,
      name: 'Optimization & Scaling',
      duration: 2,
      emails: ['optimization_guide', 'scaling_tips'],
      tasks: [
        'Analyze performance',
        'Optimize targeting',
        'Scale budget',
      ],
      unlocks: ['advanced_reporting', 'custom_segments'],
    },
    SUCCESS: {
      stage: 6,
      name: 'Success & Support',
      duration: Infinity,
      emails: ['success_story', 'continued_support'],
      tasks: [],
      unlocks: ['enterprise_features', 'api_access'],
    },
  };

  constructor(
    @Inject('REDIS_CLIENT')
    private redisClient: Redis,
  ) {}

  /**
   * Start onboarding for a new customer
   */
  async startOnboarding(
    userId: string,
    email: string,
    companyName: string,
    tier: 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE' = 'STARTER',
  ): Promise<{
    onboardingId: string;
    stage: number;
    completionPercentage: number;
    nextSteps: string[];
  }> {
    try {
      const onboardingId = `onb_${userId}_${Date.now()}`;
      const startedAt = new Date();

      const onboardingData = {
        onboardingId,
        userId,
        email,
        companyName,
        tier,
        stage: 1,
        status: 'active',
        progress: {
          completedTasks: [],
          unlockedFeatures: [],
          completionPercentage: 0,
        },
        timeline: {
          startedAt,
          currentStageStartedAt: startedAt,
          estimatedCompletionAt: new Date(startedAt.getTime() + 8 * 24 * 60 * 60 * 1000), // 8 days
        },
        emails: {
          sent: [],
          scheduled: [],
          opened: [],
        },
        metrics: {
          apiKeysGenerated: 0,
          campaignsCreated: 0,
          conversionsTracked: 0,
          impressions: 0,
        },
      };

      // Store in Redis
      await this.redisClient.hset(
        `onboarding:${userId}`,
        'data',
        JSON.stringify(onboardingData),
      );

      // Queue welcome emails
      await this.scheduleWelcomeEmails(userId, email, companyName);

      // Generate API key
  const _apiKey = await this.generateApiKey(userId);

      this.logger.log(`Onboarding started for ${email} (ID: ${onboardingId})`);

      return {
        onboardingId,
        stage: 1,
        completionPercentage: 0,
        nextSteps: [
          'Complete your profile',
          'Verify your email address',
          'Create your first campaign',
        ],
      };
    } catch (error) {
      this.logger.error(`Failed to start onboarding: ${error.message}`);
      throw new BadRequestException('Onboarding initialization failed');
    }
  }

  /**
   * Get onboarding progress
   */
  async getOnboardingProgress(userId: string): Promise<{
    stage: number;
    stageName: string;
    completionPercentage: number;
    completedTasks: string[];
    remainingTasks: string[];
    unlockedFeatures: string[];
    daysInStage: number;
    timelineEstimate: string;
  }> {
    const data = await this.getOnboardingData(userId);

    if (!data) {
      throw new BadRequestException('No active onboarding found');
    }

    const stageConfig = this.getStageConfig(data.stage);
    const completedTasks = data.progress.completedTasks;
    const remainingTasks = stageConfig.tasks.filter(
      (t) => !completedTasks.includes(t),
    );

    const daysSinceStart = Math.floor(
      (Date.now() - data.timeline.currentStageStartedAt.getTime()) / (1000 * 60 * 60 * 24),
    );

    const completionPercentage = (completedTasks.length / stageConfig.tasks.length) * 100;

    return {
      stage: data.stage,
      stageName: stageConfig.name,
      completionPercentage,
      completedTasks,
      remainingTasks,
      unlockedFeatures: data.progress.unlockedFeatures,
      daysInStage: daysSinceStart,
      timelineEstimate: `${stageConfig.duration} days to complete`,
    };
  }

  /**
   * Complete an onboarding task
   */
  async completeTask(userId: string, taskName: string): Promise<{
    completed: boolean;
    nextTask?: string;
    stageProgress: number;
    unlockedFeatures?: string[];
    shouldAdvanceStage?: boolean;
  }> {
    const data = await this.getOnboardingData(userId);

    if (!data) {
      throw new BadRequestException('No active onboarding found');
    }

    const stageConfig = this.getStageConfig(data.stage);

    if (!stageConfig.tasks.includes(taskName)) {
      throw new BadRequestException(`Invalid task: ${taskName}`);
    }

    if (data.progress.completedTasks.includes(taskName)) {
      throw new BadRequestException('Task already completed');
    }

    // Mark task as complete
    data.progress.completedTasks.push(taskName);
    data.progress.completionPercentage =
      (data.progress.completedTasks.length / stageConfig.tasks.length) * 100;

    const shouldAdvance = data.progress.completedTasks.length === stageConfig.tasks.length;

    // If all tasks complete, advance to next stage
    if (shouldAdvance && data.stage < 6) {
      data.stage++;
      data.progress.completedTasks = [];
      data.timeline.currentStageStartedAt = new Date();

      const nextStageConfig = this.getStageConfig(data.stage);

      // Unlock new features
      data.progress.unlockedFeatures = [
        ...data.progress.unlockedFeatures,
        ...nextStageConfig.unlocks,
      ];

      // Schedule next stage emails
      await this.scheduleStageEmails(userId, data.email, data.stage);
    }

    // Persist updates
    await this.redisClient.hset(
      `onboarding:${userId}`,
      'data',
      JSON.stringify(data),
    );

    const nextStageConfig = this.getStageConfig(data.stage);
    const nextTask = nextStageConfig.tasks.find(
      (t) => !data.progress.completedTasks.includes(t),
    );

    return {
      completed: true,
      nextTask,
      stageProgress: data.progress.completionPercentage,
      unlockedFeatures: shouldAdvance ? nextStageConfig.unlocks : undefined,
      shouldAdvanceStage: shouldAdvance,
    };
  }

  /**
   * Generate API key for customer
   */
  async generateApiKey(userId: string): Promise<string> {
    const key = `sk_${userId}_${Math.random().toString(36).substring(2, 15)}_${Date.now()}`;

    await this.redisClient.hset(
      `api_keys:${userId}`,
      key,
      JSON.stringify({
        createdAt: new Date(),
        tier: 'full_access',
        scopes: ['read', 'write'],
      }),
    );

    // Log for audit
    await this.redisClient.lpush(
      `audit:api_keys:created`,
      JSON.stringify({
        userId,
        apiKey: key,
        timestamp: new Date(),
      }),
    );

    return key;
  }

  /**
   * Get onboarding resources and documentation
   */
  async getOnboardingResources(_userId: string, _stage?: number): Promise<{
    guides: Array<{ title: string; url: string; duration: string }>;
    videos: Array<{ title: string; url: string; duration: string }>;
    apiDocs: Array<{ title: string; url: string }>;
    support: { email: string; phone: string; chat: string };
  }> {
    const resources = {
      guides: [
        { title: 'Getting Started', url: '/docs/getting-started', duration: '10 min' },
        { title: 'Platform Tour', url: '/docs/platform-tour', duration: '15 min' },
        { title: 'SDK Integration', url: '/docs/sdk-integration', duration: '20 min' },
        { title: 'Campaign Creation', url: '/docs/campaigns', duration: '10 min' },
        { title: 'Analytics Guide', url: '/docs/analytics', duration: '15 min' },
        { title: 'Optimization Guide', url: '/docs/optimization', duration: '20 min' },
      ],
      videos: [
        { title: 'Platform Overview', url: 'https://video.taskirx.com/overview', duration: '5 min' },
        { title: 'Creating Your First Campaign', url: 'https://video.taskirx.com/first-campaign', duration: '8 min' },
        { title: 'SDK Integration Tutorial', url: 'https://video.taskirx.com/sdk-setup', duration: '12 min' },
        { title: 'Performance Optimization', url: 'https://video.taskirx.com/optimization', duration: '10 min' },
      ],
      apiDocs: [
        { title: 'API Reference', url: '/docs/api' },
        { title: 'SDK Documentation', url: '/docs/sdks' },
        { title: 'Webhooks Guide', url: '/docs/webhooks' },
        { title: 'Code Examples', url: '/docs/examples' },
      ],
      support: {
        email: 'support@taskirx.com',
        phone: '+1-800-TASKIRX',
        chat: 'https://chat.taskirx.com',
      },
    };

    return resources;
  }

  /**
   * Get onboarding checklist
   */
  async getOnboardingChecklist(userId: string): Promise<{
    overall: number;
    stages: Array<{
      name: string;
      completed: boolean;
      tasks: Array<{ name: string; completed: boolean }>;
    }>;
  }> {
    const data = await this.getOnboardingData(userId);

    if (!data) {
      throw new BadRequestException('No active onboarding found');
    }

  const stages = Object.entries(this.ONBOARDING_STAGES).map(([_key, config]) => ({
      name: config.name,
      completed: data.stage > config.stage,
      tasks: config.tasks.map((task) => ({
        name: task,
        completed:
          data.stage > config.stage ||
          (data.stage === config.stage && data.progress.completedTasks.includes(task)),
      })),
    }));

    const totalTasks = stages.reduce((sum, s) => sum + s.tasks.length, 0);
    const completedTasks = stages.reduce(
      (sum, s) => sum + s.tasks.filter((t) => t.completed).length,
      0,
    );

    return {
      overall: totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0,
      stages,
    };
  }

  /**
   * Track onboarding metrics
   */
  async trackMetric(userId: string, metric: string, value: number): Promise<void> {
    const data = await this.getOnboardingData(userId);

    if (data && data.metrics[metric] !== undefined) {
      data.metrics[metric] = value;

      await this.redisClient.hset(
        `onboarding:${userId}`,
        'data',
        JSON.stringify(data),
      );

      // Check for milestone completion
      if (metric === 'campaignsCreated' && value >= 1) {
        await this.completeTask(userId, 'Create first campaign');
      }

      if (metric === 'conversionsTracked' && value >= 10) {
        await this.completeTask(userId, 'Verify conversions');
      }
    }
  }

  /**
   * Get onboarding completion status for all customers
   */
  async getOnboardingStats(): Promise<{
    totalOnboardings: number;
    completedOnboardings: number;
    averageTimeToCompletion: number;
    dropoffRates: Record<number, number>;
    conversionRate: number;
  }> {
    // Mock implementation - would query actual data in production
    return {
      totalOnboardings: 150,
      completedOnboardings: 132,
      averageTimeToCompletion: 7.2, // days
      dropoffRates: {
        1: 0.05, // 5% drop-off after stage 1
        2: 0.08,
        3: 0.12,
        4: 0.06,
        5: 0.02,
      },
      conversionRate: 0.88, // 88% of customers complete onboarding
    };
  }

  /**
   * Private helper methods
   */

  private async getOnboardingData(userId: string): Promise<any> {
    const data = await this.redisClient.hget(`onboarding:${userId}`, 'data');
    return data ? JSON.parse(data) : null;
  }

  private getStageConfig(stage: number): any {
    const stages = Object.values(this.ONBOARDING_STAGES);
    return stages.find((s: any) => s.stage === stage) || stages[5];
  }

  private async scheduleWelcomeEmails(userId: string, email: string, companyName: string): Promise<void> {
    const emails = [
      {
        type: 'welcome',
        to: email,
        subject: 'Welcome to TaskirX!',
        template: 'welcome',
        data: { companyName },
        scheduledAt: new Date(),
      },
      {
        type: 'platform_tour',
        to: email,
        subject: 'Your TaskirX Platform Tour',
        template: 'platform_tour',
        scheduledAt: new Date(Date.now() + 6 * 60 * 60 * 1000), // 6 hours later
      },
      {
        type: 'documentation_links',
        to: email,
        subject: 'Getting Started with TaskirX',
        template: 'documentation_links',
        scheduledAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 1 day later
      },
    ];

    for (const emailConfig of emails) {
      await this.redisClient.lpush(
        `email:queue:${emailConfig.type}`,
        JSON.stringify({
          userId,
          ...emailConfig,
        }),
      );
    }

    this.logger.log(`Scheduled ${emails.length} welcome emails for ${email}`);
  }

  private async scheduleStageEmails(userId: string, email: string, stage: number): Promise<void> {
    const stageConfig = this.getStageConfig(stage);

    for (const emailType of stageConfig.emails) {
      await this.redisClient.lpush(
        `email:queue:${emailType}`,
        JSON.stringify({
          userId,
          to: email,
          type: emailType,
          scheduledAt: new Date(),
        }),
      );
    }

    this.logger.log(`Scheduled ${stageConfig.emails.length} emails for stage ${stage}`);
  }
}
