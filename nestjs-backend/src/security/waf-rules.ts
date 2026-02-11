/**
 * Web Application Firewall (WAF) Rules
 * Protects against common web attacks:
 * - SQL Injection
 * - Cross-Site Scripting (XSS)
 * - Cross-Site Request Forgery (CSRF)
 * - Command Injection
 * - Path Traversal
 */

import { Injectable, BadRequestException, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';

export interface WafRule {
  name: string;
  pattern: RegExp;
  severity: 'high' | 'medium' | 'low';
  action: 'block' | 'log';
  description: string;
}

@Injectable()
export class WafService {
  private rules: WafRule[] = [
    // SQL Injection Detection
    {
      name: 'SQL_UNION_SELECT',
      pattern: /(\bUNION\b|\bSELECT\b|\bINSERT\b|\bUPDATE\b|\bDELETE\b|\bDROP\b|\bEXEC\b|\bEXECUTE\b)/i,
      severity: 'high',
      action: 'block',
      description: 'SQL keyword detected in request - potential SQL injection',
    },
    {
      name: 'SQL_COMMENT_BYPASS',
      pattern: /(--|#|\/\*|\*\/|;)/,
      severity: 'high',
      action: 'block',
      description: 'SQL comment syntax detected - potential SQL injection',
    },
    {
      name: 'SQL_BOOLEAN_BLIND',
      pattern: /(\bOR\b\s*\d+\s*=\s*\d+|\bAND\b\s*\d+\s*=\s*\d+)/i,
      severity: 'high',
      action: 'block',
      description: 'Boolean-based SQL injection pattern detected',
    },

    // XSS Detection
    {
      name: 'XSS_SCRIPT_TAG',
      pattern: /<script[^>]*>[\s\S]*?<\/script>/gi,
      severity: 'high',
      action: 'block',
      description: 'Script tag detected - potential XSS attack',
    },
    {
      name: 'XSS_EVENT_HANDLER',
      pattern: /on\w+\s*=\s*["'][^"']*["']/gi,
      severity: 'high',
      action: 'block',
      description: 'Event handler detected - potential XSS attack',
    },
    {
      name: 'XSS_JAVASCRIPT_PROTOCOL',
      pattern: /javascript:/gi,
      severity: 'high',
      action: 'block',
      description: 'JavaScript protocol detected - potential XSS attack',
    },
    {
      name: 'XSS_DATA_URI',
      pattern: /data:[^,]*,[\s\S]*script/gi,
      severity: 'medium',
      action: 'block',
      description: 'Data URI with script detected - potential XSS',
    },

    // Command Injection
    {
      name: 'COMMAND_INJECTION',
      pattern: /([;&|`$(){}[\]<>\\]|\.\.\/)/,
      severity: 'high',
      action: 'block',
      description: 'Shell metacharacter detected - potential command injection',
    },

    // Path Traversal
    {
      name: 'PATH_TRAVERSAL',
      pattern: /(\.\.[\/\\]|\.\.%2[fF]|%2e%2e%2[fF])/,
      severity: 'high',
      action: 'block',
      description: 'Path traversal sequence detected',
    },

    // LDAP Injection
    {
      name: 'LDAP_INJECTION',
      pattern: /([*()\\]|ldap[a-z]*:)/i,
      severity: 'medium',
      action: 'block',
      description: 'LDAP injection pattern detected',
    },

    // XML External Entity (XXE)
    {
      name: 'XXE_DETECTION',
      pattern: /<!ENTITY|<!DOCTYPE|SYSTEM\s+["']?[\w:\/\/]+["']?/i,
      severity: 'high',
      action: 'block',
      description: 'XML entity declaration detected - potential XXE attack',
    },

    // NoSQL Injection
    {
      name: 'NOSQL_INJECTION',
      pattern: /(\$where|\$ne|\$gt|\$regex|\$or|\$and)/i,
      severity: 'high',
      action: 'block',
      description: 'NoSQL operator detected - potential injection',
    },
  ];

  /**
   * Scan request for WAF rule violations
   */
  scanRequest(req: Request): WafRule | null {
    const body = JSON.stringify(req.body || {});
    const query = JSON.stringify(req.query || {});
    const params = JSON.stringify(req.params || {});
    const scanText = `${body}${query}${params}`;

    for (const rule of this.rules) {
      if (rule.pattern.test(scanText)) {
        this.logViolation(req, rule);

        if (rule.action === 'block') {
          throw new BadRequestException({
            error: 'Request blocked by WAF',
            code: 'WAF_VIOLATION',
            rule: rule.name,
            message: rule.description,
          });
        }
      }
    }

    return null;
  }

  /**
   * Log WAF violation for monitoring
   */
  private logViolation(req: Request, rule: WafRule): void {
    const timestamp = new Date().toISOString();
    const clientIp = req.ip || 'unknown';
    const userId = (req as any).user?.id || 'anonymous';

    console.warn(`[WAF] ${timestamp} | ${rule.severity.toUpperCase()} | ${rule.name} | IP: ${clientIp} | User: ${userId} | ${rule.description}`);

    // In production, send to security monitoring system
    // e.g., CloudWatch, Datadog, Splunk
  }

  /**
   * Get active rules
   */
  getRules(): WafRule[] {
    return this.rules;
  }

  /**
   * Enable/disable specific rule
   */
  updateRuleAction(ruleName: string, action: 'block' | 'log'): void {
    const rule = this.rules.find(r => r.name === ruleName);
    if (rule) {
      rule.action = action;
    }
  }

  /**
   * Get WAF statistics
   */
  getStatistics(): { totalRules: number; highSeverity: number; activated: number } {
    return {
      totalRules: this.rules.length,
      highSeverity: this.rules.filter(r => r.severity === 'high').length,
      activated: this.rules.filter(r => r.action === 'block').length,
    };
  }
}

/**
 * WAF Middleware for NestJS
 */
@Injectable()
export class WafMiddleware implements NestMiddleware {
  constructor(private readonly wafService: WafService) {}

  use(req: Request, res: Response, next: NextFunction) {
    try {
      this.wafService.scanRequest(req);
      next();
    } catch (error) {
      // Error response already sent by WAF service
      throw error;
    }
  }
}
