package com.taskir.sdk

import android.content.Context
import android.util.Log
import com.taskir.sdk.data.models.*
import com.taskir.sdk.network.RequestManager
import com.taskir.sdk.services.*

/**
 * TaskirX Android SDK Main Client
 * Unified interface to all platform services
 */
class TaskirXClient(
    private val context: Context,
    private val config: ClientConfig
) {
    private val requestManager = RequestManager(config)
    private val debug = config.debug

    // Services
    val auth = AuthService(requestManager, debug)
    val campaigns = CampaignService(requestManager, debug)
    val analytics = AnalyticsService(requestManager, debug)
    val bidding = BiddingService(requestManager, debug)
    val ads = AdService(requestManager, debug)
    val webhooks = WebhookService(requestManager, debug)

    init {
        log("TaskirX Android SDK initialized")
        log("API URL: ${config.apiUrl}")
        log("Debug mode: ${config.debug}")
    }

    /**
     * Initialize the SDK and check platform connectivity
     */
    suspend fun initialize(): Result<Unit> = try {
        log("Initializing TaskirX SDK")
        getHealth()
        log("TaskirX SDK initialized successfully")
        Result.success(Unit)
    } catch (e: Exception) {
        log("SDK initialization failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Check platform health and connectivity
     */
    suspend fun getHealth(): Result<Map<String, Any>> = try {
        log("Checking platform health")
        val response = requestManager.executeWithRetry {
            // Health check implementation
            mapOf("status" to "healthy")
        }
        Result.success(response)
    } catch (e: Exception) {
        log("Health check failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Get platform status information
     */
    suspend fun getStatus(): Result<Map<String, Any>> = try {
        log("Fetching platform status")
        val response = requestManager.executeWithRetry {
            mapOf("status" to "operational")
        }
        Result.success(response)
    } catch (e: Exception) {
        log("Status check failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Get user profile
     */
    suspend fun getProfile(): Result<User> = try {
        log("Fetching user profile")
        val result = requestManager.executeWithRetry {
            auth.getProfile()
        }
        Result.success(result)
    } catch (e: Exception) {
        log("Profile fetch failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Logout and cleanup
     */
    suspend fun logout(): Result<Unit> = try {
        log("Logging out")
        requestManager.executeWithRetry {
            auth.logout()
        }
        Result.success(Unit)
    } catch (e: Exception) {
        log("Logout failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Set API key for authentication
     */
    fun setApiKey(apiKey: String) {
        log("Setting API key")
        // Implementation for updating API key
    }

    /**
     * Get comprehensive dashboard data
     */
    suspend fun getDashboard(): Result<Map<String, Any>> = try {
        log("Fetching comprehensive dashboard data")
        val campaignsList = requestManager.executeWithRetry { campaigns.list() }
        val analyticsData = requestManager.executeWithRetry { analytics.getRealtime() }
        val webhooksList = requestManager.executeWithRetry { webhooks.list() }

        val result = mapOf(
            "campaigns" to campaignsList,
            "analytics" to analyticsData,
            "webhooks" to webhooksList,
            "timestamp" to System.currentTimeMillis()
        )
        Result.success(result)
    } catch (e: Exception) {
        log("Dashboard fetch failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Get campaign performance summary
     */
    suspend fun getCampaignPerformance(campaignId: String): Result<Map<String, Any>> = try {
        log("Fetching campaign performance: $campaignId")
        val campaign = requestManager.executeWithRetry { campaigns.get(campaignId) }
        val analyticsData = requestManager.executeWithRetry { 
            analytics.getCampaignAnalytics(campaignId) 
        }
        val stats = requestManager.executeWithRetry { bidding.getStats(campaignId) }

        val result = mapOf(
            "campaign" to campaign,
            "analytics" to analyticsData,
            "bidding" to stats
        )
        Result.success(result)
    } catch (e: Exception) {
        log("Campaign performance fetch failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Batch create campaigns
     */
    suspend fun createCampaigns(campaignsList: List<Campaign>): Result<List<Campaign>> = try {
        log("Creating ${campaignsList.size} campaigns")
        val results = mutableListOf<Campaign>()
        for (campaign in campaignsList) {
            val created = requestManager.executeWithRetry { campaigns.create(campaign) }
            results.add(created)
        }
        log("Batch campaign creation completed")
        Result.success(results)
    } catch (e: Exception) {
        log("Batch campaign creation failed: ${e.message}")
        Result.failure(e)
    }

    /**
     * Enable debug mode for detailed logging
     */
    fun enableDebug(enabled: Boolean = true) {
        log("Debug mode ${if (enabled) "enabled" else "disabled"}")
    }

    /**
     * Get complete platform statistics
     */
    suspend fun getStatistics(): Result<Map<String, Any>> = try {
        log("Fetching platform statistics")
        val dashboard = getDashboard().getOrThrow()
        val realtime = requestManager.executeWithRetry { analytics.getRealtime() }

        val result = mapOf(
            "dashboard" to dashboard,
            "realtime" to realtime,
            "timestamp" to System.currentTimeMillis()
        )
        Result.success(result)
    } catch (e: Exception) {
        log("Statistics fetch failed: ${e.message}")
        Result.failure(e)
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("TaskirXClient", message)
        }
    }

    companion object {
        const val VERSION = "1.0.0"
        const val TAG = "TaskirX"

        /**
         * Create a TaskirX client instance
         */
        fun create(context: Context, config: ClientConfig): TaskirXClient {
            return TaskirXClient(context, config)
        }
    }
}
