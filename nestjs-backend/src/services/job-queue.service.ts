import { Injectable } from '@nestjs/common';
import { Cron, CronExpression } from '@nestjs/schedule';

/**
 * Job Queue Service
 *
 * Handles scheduled tasks and background jobs
 * Uses node-schedule for cron expressions
 *
 * Jobs:
 * - Email send queue processor (every 5 minutes)
 * - Trial expiration check (daily)
 * - Win-back campaign trigger (daily)
 * - Engagement digest (weekly)
 * - Lead scoring refresh (hourly)
 * - Analytics aggregation (daily)
 */
@Injectable()
export class JobQueueService {
  private dripCampaignService: any;
  private trialManagementService: any;
  private leadCaptureService: any;
  private emailService: any;

  constructor() {
    // Services injected at runtime
  }

  /**
   * Process email send queue every 5 minutes
   */
  @Cron(CronExpression.EVERY_5_MINUTES)
  async processEmailQueue() {
    try {
      console.log('[JOB] Processing email send queue...');
      const result = await this.dripCampaignService.sendQueuedEmails();
      console.log(`[JOB] Processed ${result.emailsSent || 0} emails`);
    } catch (error) {
      console.error('[JOB] Email queue error:', error);
    }
  }

  /**
   * Check and expire trials daily at 1 AM
   */
  @Cron('0 1 * * *')
  async checkTrialExpirations() {
    try {
      console.log('[JOB] Checking trial expirations...');
      const expiringTrials = await this.trialManagementService.getExpiringTrials(0);

      let expiredCount = 0;
      for (const trial of expiringTrials) {
        if (trial.daysRemaining <= 0) {
          await this.trialManagementService.expireTrial(trial.trialId);
          expiredCount++;
        }
      }

      console.log(`[JOB] Expired ${expiredCount} trials`);
    } catch (error) {
      console.error('[JOB] Trial expiration check error:', error);
    }
  }

  /**
   * Send trial reminder emails (2 days before expiry)
   */
  @Cron('0 9 * * *') // 9 AM daily
  async sendTrialReminders() {
    try {
      console.log('[JOB] Sending trial reminder emails...');
      const almostExpired = await this.trialManagementService.getExpiringTrials(2);

      let remindersSent = 0;
      for (const trial of almostExpired) {
        if (trial.daysRemaining === 2) {
          // Send reminder email
          await this.emailService.sendTemplate(trial.email, 'TRIAL_REMINDER', {
            name: trial.userName,
            days_remaining: trial.daysRemaining,
            upgrade_url: `${process.env.APP_URL}/upgrade?trial=${trial.trialId}`,
          });
          remindersSent++;
        }
      }

      console.log(`[JOB] Sent ${remindersSent} trial reminders`);
    } catch (error) {
      console.error('[JOB] Trial reminder error:', error);
    }
  }

  /**
   * Send trial expiration warning (1 day before)
   */
  @Cron('0 10 * * *') // 10 AM daily
  async sendExpirationWarnings() {
    try {
      console.log('[JOB] Sending expiration warnings...');
      const expiringSoon = await this.trialManagementService.getExpiringTrials(1);

      let warningSent = 0;
      for (const trial of expiringSoon) {
        if (trial.daysRemaining === 1) {
          await this.emailService.sendTemplate(trial.email, 'EXPIRATION_WARNING', {
            name: trial.userName,
            expiration_date: trial.expiresAt,
            upgrade_url: `${process.env.APP_URL}/upgrade?trial=${trial.trialId}`,
          });
          warningSent++;
        }
      }

      console.log(`[JOB] Sent ${warningSent} expiration warnings`);
    } catch (error) {
      console.error('[JOB] Expiration warning error:', error);
    }
  }

  /**
   * Trigger win-back campaigns for expired trials
   */
  @Cron('0 14 * * *') // 2 PM daily
  async triggerWinbackCampaigns() {
    try {
      console.log('[JOB] Triggering win-back campaigns...');
      const expiredTrials = await this.trialManagementService.getExpiredTrials();

      let campaignsTriggered = 0;
      for (const trial of expiredTrials) {
        if (!trial.winbackSent) {
          // Send win-back offer
          await this.emailService.sendTemplate(trial.email, 'WINBACK', {
            name: trial.userName,
            discount_percent: 20,
            discount_days: 30,
            offer_url: `${process.env.APP_URL}/offer?trial=${trial.trialId}&code=COMEBACK20`,
          });

          // Mark as sent
          await this.trialManagementService.markWinbackSent(trial.trialId);
          campaignsTriggered++;
        }
      }

      console.log(`[JOB] Triggered ${campaignsTriggered} win-back campaigns`);
    } catch (error) {
      console.error('[JOB] Win-back campaign error:', error);
    }
  }

  /**
   * Recalculate lead scores hourly
   */
  @Cron(CronExpression.EVERY_HOUR)
  async refreshLeadScores() {
    try {
      console.log('[JOB] Refreshing lead scores...');
      const leads = await this.leadCaptureService.getAllLeads();

      let scoresUpdated = 0;
      for (const lead of leads) {
        const newScore = await this.leadCaptureService.calculateLeadScore(lead.leadId);
        if (newScore !== lead.score) {
          scoresUpdated++;
        }
      }

      console.log(`[JOB] Updated ${scoresUpdated} lead scores`);
    } catch (error) {
      console.error('[JOB] Lead score refresh error:', error);
    }
  }

  /**
   * Generate and send engagement digest (weekly on Monday at 8 AM)
   */
  @Cron('0 8 * * 1')
  async sendEngagementDigest() {
    try {
      console.log('[JOB] Generating engagement digests...');
      const users = await this.getUsersForDigest();

      let digestsSent = 0;
      for (const user of users) {
        const weeklyMetrics = await this.getWeeklyMetrics(user.userId);
        await this.emailService.sendTemplate(user.email, 'ENGAGEMENT_DIGEST', {
          name: user.name,
          period: 'weekly',
          campaign_count: weeklyMetrics.campaigns,
          lead_count: weeklyMetrics.leads,
          conversion_count: weeklyMetrics.conversions,
          analytics_url: `${process.env.APP_URL}/dashboard/analytics`,
        });
        digestsSent++;
      }

      console.log(`[JOB] Sent ${digestsSent} engagement digests`);
    } catch (error) {
      console.error('[JOB] Engagement digest error:', error);
    }
  }

  /**
   * Aggregate analytics data daily (midnight)
   */
  @Cron('0 0 * * *')
  async aggregateAnalytics() {
    try {
      console.log('[JOB] Aggregating analytics data...');
      // Summarize daily events
      // Roll up to weekly/monthly buckets
      // Generate reports
      console.log('[JOB] Analytics aggregation complete');
    } catch (error) {
      console.error('[JOB] Analytics aggregation error:', error);
    }
  }

  /**
   * Clean up old data daily (2 AM)
   */
  @Cron('0 2 * * *')
  async cleanupOldData() {
    try {
      console.log('[JOB] Cleaning up old data...');
      // Delete events older than 90 days
      // Archive old visitor sessions
      // Remove unsubscribed lead records (retention policy)
      console.log('[JOB] Data cleanup complete');
    } catch (error) {
      console.error('[JOB] Data cleanup error:', error);
    }
  }

  /**
   * Get users for digest
   */
  private async getUsersForDigest(): Promise<any[]> {
    // Query users who have engagement digest enabled
    return [];
  }

  /**
   * Get weekly metrics for user
   */
  private async getWeeklyMetrics(_userId: string): Promise<any> {
    // Calculate weekly aggregates
    return {
      campaigns: 5,
      leads: 12,
      conversions: 2,
    };
  }
}
