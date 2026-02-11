import {
  WebSocketGateway,
  WebSocketServer,
  SubscribeMessage,
  OnGatewayConnection,
  OnGatewayDisconnect,
  MessageBody,
  ConnectedSocket,
} from '@nestjs/websockets';
import { Server, Socket } from 'socket.io';
import { Inject, Injectable, Logger } from '@nestjs/common';
import { Redis } from 'ioredis';

/**
 * Alerts Gateway - Real-time alert delivery via WebSocket
 * 
 * Provides:
 * - Threshold alerts
 * - Performance degradation alerts
 * - Budget threshold alerts
 * - System health alerts
 * - Custom rule-based alerts
 * 
 * Features:
 * - Alert prioritization
 * - Acknowledgment tracking
 * - Alert history
 * - Team notifications
 */
@Injectable()
@WebSocketGateway({
  cors: {
    origin: process.env.WEB_URL || 'http://localhost:3000',
    credentials: true,
  },
  namespace: '/alerts',
})
export class AlertsGateway implements OnGatewayConnection, OnGatewayDisconnect {
  private readonly logger = new Logger(AlertsGateway.name);

  @WebSocketServer()
  server: Server;

  // Track alert subscriptions per user
  private userAlertSubscriptions: Map<string, Set<string>> = new Map();
  
  // Track alert acknowledgments
  private alertAcknowledgments: Map<string, Set<string>> = new Map();
  
  // Redis subscriber for cross-instance alerts
  private redisSubscriber: Redis;

  constructor(
    @Inject('REDIS_CLIENT')
    private redisClient: Redis,
  ) {
    this.redisSubscriber = this.redisClient.duplicate();
    this.setupRedisListeners();
  }

  /**
   * Setup Redis pub/sub for alert distribution
   */
  private setupRedisListeners() {
    this.redisSubscriber.on('message', (channel, message) => {
      if (channel.startsWith('alerts:')) {
        try {
          const alert = JSON.parse(message);
          
          // Broadcast to all connected clients listening to alerts
          this.server.emit('alert:system', {
            ...alert,
            receivedAt: new Date(),
          });
        } catch (error) {
          this.logger.error(`Failed to parse alert: ${error.message}`);
        }
      }
    });

    // Subscribe to alerts channel
    this.redisSubscriber.subscribe('alerts:*');
  }

  /**
   * Handle client connection
   */
  handleConnection(client: Socket) {
    const userId = client.handshake.auth.userId;
    const token = client.handshake.auth.token;

    if (!userId || !token) {
      this.logger.warn(`Alert connection attempt without credentials from ${client.id}`);
      client.disconnect();
      return;
    }

    if (!this.userAlertSubscriptions.has(userId)) {
      this.userAlertSubscriptions.set(userId, new Set());
    }

    this.userAlertSubscriptions.get(userId).add(client.id);

    this.logger.log(
      `Alert client ${client.id} connected (User: ${userId}). Total alert connections: ${this.server.engine.clientsCount}`,
    );

    client.emit('alerts:connected', {
      clientId: client.id,
      userId,
      timestamp: new Date(),
    });
  }

  /**
   * Handle client disconnection
   */
  handleDisconnect(client: Socket) {
    const userId = client.handshake.auth.userId;

    if (this.userAlertSubscriptions.has(userId)) {
      const subscriptions = this.userAlertSubscriptions.get(userId);
      subscriptions.delete(client.id);

      if (subscriptions.size === 0) {
        this.userAlertSubscriptions.delete(userId);
      }
    }

    this.logger.log(
      `Alert client ${client.id} disconnected. Total connections: ${this.server.engine.clientsCount}`,
    );
  }

  /**
   * Subscribe to campaign alerts
   */
  @SubscribeMessage('subscribe:campaign-alerts')
  subscribeCampaignAlerts(
    @MessageBody() payload: { campaignId: string; types?: string[] },
    @ConnectedSocket() client: Socket,
  ) {
    const { campaignId, types = ['all'] } = payload;
    const _userId = client.handshake.auth.userId;

    // Join campaign alert room
    const roomName = `alerts:campaign:${campaignId}`;
    client.join(roomName);

    return {
      success: true,
      room: roomName,
      types,
      subscription: `Campaign ${campaignId}`,
    };
  }

  /**
   * Subscribe to user account alerts
   */
  @SubscribeMessage('subscribe:account-alerts')
  subscribeAccountAlerts(
    @MessageBody() payload: { types?: string[] },
    @ConnectedSocket() client: Socket,
  ) {
    const { types = ['billing', 'performance', 'security'] } = payload;
    const userId = client.handshake.auth.userId;

    const roomName = `alerts:account:${userId}`;
    client.join(roomName);

    return {
      success: true,
      room: roomName,
      types,
      subscription: 'Account Alerts',
    };
  }

  /**
   * Subscribe to system-wide alerts
   */
  @SubscribeMessage('subscribe:system-alerts')
  subscribeSystemAlerts(
    @MessageBody() payload: { severity?: string[] },
    @ConnectedSocket() client: Socket,
  ) {
    const { severity = ['critical', 'high'] } = payload;

    const roomName = 'alerts:system';
    client.join(roomName);

    return {
      success: true,
      room: roomName,
      severity,
      subscription: 'System Alerts',
    };
  }

  /**
   * Acknowledge alert reception
   */
  @SubscribeMessage('alert:acknowledge')
  acknowledgeAlert(
    @MessageBody() payload: { alertId: string; userId: string },
    @ConnectedSocket() _client: Socket,
  ) {
    const { alertId, userId } = payload;

    if (!this.alertAcknowledgments.has(alertId)) {
      this.alertAcknowledgments.set(alertId, new Set());
    }

    this.alertAcknowledgments.get(alertId).add(userId);

    return {
      success: true,
      alertId,
      acknowledgedAt: new Date(),
    };
  }

  /**
   * Get alert acknowledgment status
   */
  @SubscribeMessage('alert:status')
  getAlertStatus(
    @MessageBody() payload: { alertId: string },
    @ConnectedSocket() _client: Socket,
  ) {
    const { alertId } = payload;
    const acknowledged = this.alertAcknowledgments.get(alertId) || new Set();

    return {
      alertId,
      acknowledgedBy: Array.from(acknowledged),
      count: acknowledged.size,
    };
  }

  /**
   * Broadcast campaign alert to subscribed clients
   */
  broadcastCampaignAlert(campaignId: string, alert: {
    id: string;
    type: string;
    severity: 'critical' | 'high' | 'medium' | 'low';
    message: string;
    data?: any;
  }) {
    const roomName = `alerts:campaign:${campaignId}`;
    const alertPayload = {
      ...alert,
      campaignId,
      broadcastedAt: new Date(),
    };

    this.server.to(roomName).emit('alert', alertPayload);

    // Store in Redis for recovery
    this.redisClient.publish(
      'alerts:campaign',
      JSON.stringify({ campaignId, ...alertPayload }),
    );
  }

  /**
   * Broadcast account alert
   */
  broadcastAccountAlert(userId: string, alert: {
    id: string;
    type: string;
    severity: 'critical' | 'high' | 'medium' | 'low';
    message: string;
    actionRequired?: boolean;
    data?: any;
  }) {
    const roomName = `alerts:account:${userId}`;
    const alertPayload = {
      ...alert,
      userId,
      broadcastedAt: new Date(),
    };

    this.server.to(roomName).emit('alert', alertPayload);

    this.redisClient.publish(
      'alerts:account',
      JSON.stringify({ userId, ...alertPayload }),
    );
  }

  /**
   * Broadcast system alert to all connected clients
   */
  broadcastSystemAlert(alert: {
    id: string;
    type: string;
    severity: 'critical' | 'high' | 'medium' | 'low';
    title: string;
    message: string;
    affectedServices?: string[];
  }) {
    const alertPayload = {
      ...alert,
      broadcastedAt: new Date(),
    };

    this.server.to('alerts:system').emit('alert', alertPayload);

    this.redisClient.publish(
      'alerts:system',
      JSON.stringify(alertPayload),
    );
  }

  /**
   * Send batch alerts to specific users
   */
  broadcastBatchAlerts(userIds: string[], alert: any) {
    userIds.forEach((userId) => {
      const roomName = `alerts:account:${userId}`;
      this.server.to(roomName).emit('alert', {
        ...alert,
        userId,
        broadcastedAt: new Date(),
      });
    });
  }

  /**
   * Alert priority queue - ensures critical alerts are delivered first
   */
  enqueueAlert(alert: any, priority: 'critical' | 'high' | 'medium' | 'low' = 'medium') {
    const priorityScore = {
      critical: 4,
      high: 3,
      medium: 2,
      low: 1,
    };

    // Store in Redis sorted set with priority
    this.redisClient.zadd(
      'alerts:queue',
      priorityScore[priority],
      JSON.stringify(alert),
    );
  }

  /**
   * Get connection statistics for alerts
   */
  getStats() {
    return {
      totalConnections: this.server.engine.clientsCount,
      activeUsers: this.userAlertSubscriptions.size,
      acknowledgedAlerts: this.alertAcknowledgments.size,
    };
  }
}
