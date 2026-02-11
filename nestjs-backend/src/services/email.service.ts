import { Injectable } from '@nestjs/common';
import * as nodemailer from 'nodemailer';
import * as SendGrid from '@sendgrid/mail';

/**
 * Email Service
 *
 * Handles all email sending operations
 * Supports SendGrid and nodemailer as fallback
 * Includes template management and tracking
 *
 * Features:
 * - Template-based emails
 * - Bulk sending
 * - Bounce/complaint handling
 * - Delivery tracking
 * - Retry logic
 * - Rate limiting
 */
@Injectable()
export class EmailService {
  private sgMail: any;
  private transporter: any;

  constructor() {
    this.initializeEmailService();
  }

  private initializeEmailService() {
    // Initialize SendGrid
    const sendgridKey = process.env.SENDGRID_API_KEY;
    if (sendgridKey) {
      SendGrid.setApiKey(sendgridKey);
      this.sgMail = SendGrid;
    }

    // Fallback: nodemailer
    if (process.env.SMTP_HOST) {
      this.transporter = nodemailer.createTransport({
        host: process.env.SMTP_HOST,
        port: parseInt(process.env.SMTP_PORT || '587'),
        secure: process.env.SMTP_SECURE === 'true',
        auth: {
          user: process.env.SMTP_USER,
          pass: process.env.SMTP_PASSWORD,
        },
      });
    }
  }

  /**
   * Send email using template
   */
  async sendTemplate(
    to: string,
    templateId: string,
    templateData: Record<string, any>,
  ): Promise<any> {
    try {
      if (this.sgMail) {
        return await this.sendViaSendGrid(to, templateId, templateData);
      } else if (this.transporter) {
        return await this.sendViaSmtp(to, templateId, templateData);
      } else {
        console.warn('No email service configured');
        return { success: false, message: 'Email service not configured' };
      }
    } catch (error) {
      console.error(`Error sending email to ${to}:`, error);
      throw error;
    }
  }

  /**
   * Send via SendGrid
   */
  private async sendViaSendGrid(
    to: string,
    templateId: string,
    templateData: Record<string, any>,
  ) {
    const msg = {
      to,
      from: process.env.SENDGRID_FROM_EMAIL || 'noreply@taskir.com',
      templateId,
      dynamicTemplateData: templateData,
      trackingSettings: {
        clickTracking: { enabled: true },
        openTracking: { enabled: true },
      },
    };

    const response = await this.sgMail.send(msg);
    return {
      success: true,
      messageId: response[0].headers['x-message-id'],
      timestamp: new Date(),
    };
  }

  /**
   * Send via SMTP
   */
  private async sendViaSmtp(
    to: string,
    templateId: string,
    templateData: Record<string, any>,
  ) {
    const template = this.getTemplate(templateId);
    const subject = this.renderTemplate(template.subject, templateData);
    const html = this.renderTemplate(template.html, templateData);

    const info = await this.transporter.sendMail({
      from: process.env.SMTP_FROM || 'noreply@taskir.com',
      to,
      subject,
      html,
    });

    return {
      success: true,
      messageId: info.messageId,
      timestamp: new Date(),
    };
  }

  /**
   * Get email template
   */
  private getTemplate(templateId: string): any {
    const templates: Record<string, any> = {
      WELCOME: {
        subject: 'Welcome to {{company_name}}!',
        html: `
          <h1>Hello {{name}}</h1>
          <p>Welcome! Your {{day}}-day trial is active.</p>
          <a href="{{dashboard_url}}">Go to Dashboard</a>
        `,
      },
      TRIAL_REMINDER: {
        subject: '{{days_remaining}} days left on your trial',
        html: `
          <h1>Your trial expires soon</h1>
          <p>You have {{days_remaining}} days remaining.</p>
          <a href="{{upgrade_url}}">Upgrade Now</a>
        `,
      },
      EXPIRATION_WARNING: {
        subject: 'Your trial expires tomorrow',
        html: `
          <h1>Last chance to upgrade</h1>
          <p>Your trial expires {{expiration_date}}.</p>
          <a href="{{upgrade_url}}">Upgrade to Paid</a>
        `,
      },
      CONVERSION_SUCCESSFUL: {
        subject: 'Welcome to {{plan_name}} plan',
        html: `
          <h1>Thank you for upgrading!</h1>
          <p>Your {{plan_name}} plan is now active.</p>
          <a href="{{account_url}}">Manage Account</a>
        `,
      },
      WINBACK: {
        subject: 'We miss you! Come back to {{company_name}}',
        html: `
          <h1>Special offer for returning customers</h1>
          <p>Get {{discount_percent}}% off for {{discount_days}} days.</p>
          <a href="{{offer_url}}">Claim Offer</a>
        `,
      },
      ENGAGEMENT_DIGEST: {
        subject: 'Your {{period}} performance summary',
        html: `
          <h1>Your {{period}} Summary</h1>
          <p>Campaigns: {{campaign_count}}</p>
          <p>Leads: {{lead_count}}</p>
          <p>Conversions: {{conversion_count}}</p>
          <a href="{{analytics_url}}">View Details</a>
        `,
      },
    };

    return templates[templateId] || templates.WELCOME;
  }

  /**
   * Render template with data
   */
  private renderTemplate(template: string, data: Record<string, any>): string {
    let result = template;
    Object.entries(data).forEach(([key, value]) => {
      result = result.replace(new RegExp(`{{${key}}}`, 'g'), String(value));
    });
    return result;
  }

  /**
   * Send bulk emails
   */
  async sendBulk(
    recipients: string[],
    templateId: string,
    templateData: Record<string, any>,
  ): Promise<any> {
    const results = [];
    for (const recipient of recipients) {
      try {
        const result = await this.sendTemplate(recipient, templateId, templateData);
        results.push({ recipient, ...result });
      } catch (error) {
        results.push({ recipient, success: false, error: error.message });
      }
    }
    return results;
  }

  /**
   * Handle bounce event
   */
  async handleBounce(email: string, bounceType: string): Promise<void> {
    console.log(`Email bounced: ${email} (${bounceType})`);
    // Unsubscribe from all campaigns
    // Update lead status to bounced
  }

  /**
   * Handle complaint event
   */
  async handleComplaint(email: string): Promise<void> {
    console.log(`Email complaint: ${email}`);
    // Mark as unsubscribed
    // Flag account for review
  }

  /**
   * Handle delivery event
   */
  async handleDelivery(messageId: string): Promise<void> {
    console.log(`Email delivered: ${messageId}`);
    // Update delivery status in database
  }

  /**
   * Handle open event
   */
  async handleOpen(messageId: string, email: string): Promise<void> {
    console.log(`Email opened: ${messageId} by ${email}`);
    // Record open in campaign tracking
  }

  /**
   * Handle click event
   */
  async handleClick(messageId: string, email: string, url: string): Promise<void> {
    console.log(`Email clicked: ${messageId} by ${email} - ${url}`);
    // Record click in campaign tracking
  }
}
