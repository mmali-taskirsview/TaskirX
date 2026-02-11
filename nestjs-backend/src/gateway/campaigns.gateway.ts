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
 * Campaigns Gateway - Real-time campaign updates via WebSocket
 * 
 * Provides:
 * - Live campaign status updates
 * - Budget burn tracking
 * - Bid adjustments
 * - Campaign performance streams
 * - Multi-user collaboration
 * 
 * Features:
 * - Collaborative editing
 * - Change notifications
 * - User activity tracking
 * - Broadcast queue
 */
@Injectable()
@WebSocketGateway({
  cors: {
    origin: process.env.WEB_URL || 'http://localhost:3000',
    credentials: true,
  },
  namespace: '/campaigns',
})
export class CampaignsGateway implements OnGatewayConnection, OnGatewayDisconnect {
  private readonly logger = new Logger(CampaignsGateway.name);

  @WebSocketServer()
  server: Server;

  // Track users viewing each campaign
  private campaignViewers: Map<string, Set<string>> = new Map();
  
  // Track user presence
  private userPresence: Map<string, { campaignIds: Set<string>; lastSeen: Date }> = new Map();
  
  // Redis subscriber
  private redisSubscriber: Redis;

  constructor(
    @Inject('REDIS_CLIENT')
    private redisClient: Redis,
  ) {
    this.redisSubscriber = this.redisClient.duplicate();
    this.setupRedisListeners();
  }

  /**
   * Setup Redis pub/sub for campaign updates
   */
  private setupRedisListeners() {
    this.redisSubscriber.on('message', (channel, message) => {
      if (channel.startsWith('campaign:')) {
        try {
          const update = JSON.parse(message);
          const campaignId = channel.split(':')[1];

          // Broadcast to all clients viewing this campaign
          const viewers = this.campaignViewers.get(campaignId);
          if (viewers && viewers.size > 0) {
            this.server.to(Array.from(viewers)).emit('campaign:update', {
              campaignId,
              ...update,
              receivedAt: new Date(),
            });
          }
        } catch (error) {
          this.logger.error(`Failed to parse campaign update: ${error.message}`);
        }
      }
    });

    // Subscribe to campaign updates (pattern subscription requires psubscribe)
    this.redisSubscriber.psubscribe('campaign:*');
  }

  /**
   * Handle client connection
   */
  handleConnection(client: Socket) {
    const userId = client.handshake.auth.userId;
    const token = client.handshake.auth.token;

    if (!userId || !token) {
      this.logger.warn(`Campaign connection attempt without credentials from ${client.id}`);
      client.disconnect();
      return;
    }

    this.userPresence.set(client.id, {
      campaignIds: new Set(),
      lastSeen: new Date(),
    });

    this.logger.log(
      `Campaign client ${client.id} connected (User: ${userId}). Total connections: ${this.server.engine.clientsCount}`,
    );

    client.emit('campaigns:connected', {
      clientId: client.id,
      userId,
      timestamp: new Date(),
    });
  }

  /**
   * Handle client disconnection
   */
  handleDisconnect(client: Socket) {
    const presence = this.userPresence.get(client.id);

    if (presence) {
      // Notify other viewers about this user leaving
      presence.campaignIds.forEach((campaignId) => {
        const viewers = this.campaignViewers.get(campaignId);
        if (viewers) {
          viewers.delete(client.id);
          
          // Broadcast presence update
          this.server.to(Array.from(viewers)).emit('user:left', {
            userId: client.handshake.auth.userId,
            campaignId,
            timestamp: new Date(),
          });

          if (viewers.size === 0) {
            this.campaignViewers.delete(campaignId);
          }
        }
      });

      this.userPresence.delete(client.id);
    }

    this.logger.log(
      `Campaign client ${client.id} disconnected. Total connections: ${this.server.engine.clientsCount}`,
    );
  }

  /**
   * Join campaign room (start viewing/editing)
   */
  @SubscribeMessage('join:campaign')
  joinCampaign(
    @MessageBody() payload: { campaignId: string },
    @ConnectedSocket() client: Socket,
  ) {
    const { campaignId } = payload;
    const userId = client.handshake.auth.userId;

    // Add to viewers
    if (!this.campaignViewers.has(campaignId)) {
      this.campaignViewers.set(campaignId, new Set());
    }
    this.campaignViewers.get(campaignId).add(client.id);

    // Track user presence
    const presence = this.userPresence.get(client.id);
    if (presence) {
      presence.campaignIds.add(campaignId);
      presence.lastSeen = new Date();
    }

    // Join socket.io room
    client.join(`campaign:${campaignId}`);

    // Notify others
    this.server.to(`campaign:${campaignId}`).emit('user:joined', {
      userId,
      clientId: client.id,
      campaignId,
      timestamp: new Date(),
    });

    // Send viewer list
    const viewers = Array.from(this.campaignViewers.get(campaignId) || []);
    client.emit('viewers:list', {
      campaignId,
      viewers,
      count: viewers.length,
    });

    return {
      success: true,
      campaignId,
      viewers: viewers.length,
    };
  }

  /**
   * Leave campaign room
   */
  @SubscribeMessage('leave:campaign')
  leaveCampaign(
    @MessageBody() payload: { campaignId: string },
    @ConnectedSocket() client: Socket,
  ) {
    const { campaignId } = payload;

    const viewers = this.campaignViewers.get(campaignId);
    if (viewers) {
      viewers.delete(client.id);
    }

    const presence = this.userPresence.get(client.id);
    if (presence) {
      presence.campaignIds.delete(campaignId);
    }

    client.leave(`campaign:${campaignId}`);

    return { success: true, campaignId };
  }

  /**
   * Stream campaign status updates
   */
  @SubscribeMessage('stream:campaign-status')
  streamCampaignStatus(
    @MessageBody() payload: { campaignId: string },
    @ConnectedSocket() client: Socket,
  ) {
    const { campaignId } = payload;

    const roomName = `campaign:${campaignId}:status`;
    client.join(roomName);

    return {
      success: true,
      room: roomName,
      streaming: true,
    };
  }

  /**
   * Stream budget burn for a campaign
   */
  @SubscribeMessage('stream:budget-burn')
  streamBudgetBurn(
    @MessageBody() payload: { campaignId: string; interval?: number },
    @ConnectedSocket() client: Socket,
  ) {
    const { campaignId, interval = 5000 } = payload;

    const roomName = `campaign:${campaignId}:budget`;
    client.join(roomName);

    // Start sending budget updates at specified interval
    const intervalId = setInterval(() => {
      // TODO: Calculate current budget burn
      client.emit('budget:update', {
        campaignId,
        timestamp: new Date(),
      });
    }, interval);

    // Cleanup on disconnect
    client.on('disconnect', () => {
      clearInterval(intervalId);
    });

    return {
      success: true,
      room: roomName,
      interval,
    };
  }

  /**
   * Broadcast campaign update (status, pause, resume, etc.)
   */
  broadcastCampaignUpdate(campaignId: string, update: {
    status?: string;
    reason?: string;
    data?: any;
  }) {
    const roomName = `campaign:${campaignId}`;
    const updatePayload = {
      campaignId,
      ...update,
      broadcastedAt: new Date(),
    };

    this.server.to(roomName).emit('campaign:update', updatePayload);

    // Also distribute via Redis for multi-instance support
    this.redisClient.publish(
      `campaign:${campaignId}`,
      JSON.stringify(updatePayload),
    );
  }

  /**
   * Broadcast bid adjustment event
   */
  broadcastBidAdjustment(campaignId: string, adjustment: {
    type: 'automatic' | 'manual';
    oldBid: number;
    newBid: number;
    reason: string;
    appliedAt: Date;
  }) {
    const payload = {
      campaignId,
      ...adjustment,
    };

    this.server.to(`campaign:${campaignId}`).emit('bid:adjusted', payload);
    this.redisClient.publish(
      `campaign:${campaignId}:bids`,
      JSON.stringify(payload),
    );
  }

  /**
   * Broadcast budget threshold alert
   */
  broadcastBudgetThreshold(campaignId: string, threshold: {
    current: number;
    limit: number;
    percentage: number;
    message: string;
  }) {
    const payload = {
      campaignId,
      ...threshold,
      timestamp: new Date(),
    };

    this.server.to(`campaign:${campaignId}`).emit('budget:threshold', payload);
  }

  /**
   * Get active viewers for a campaign
   */
  getCampaignViewers(campaignId: string): number {
    return this.campaignViewers.get(campaignId)?.size || 0;
  }

  /**
   * Get user presence information
   */
  getUserPresence(clientId: string) {
    return this.userPresence.get(clientId);
  }

  /**
   * Get gateway statistics
   */
  getStats() {
    return {
      totalConnections: this.server.engine.clientsCount,
      activeCampaigns: this.campaignViewers.size,
      totalViewers: Array.from(this.campaignViewers.values()).reduce(
        (sum, viewers) => sum + viewers.size,
        0,
      ),
    };
  }
}
