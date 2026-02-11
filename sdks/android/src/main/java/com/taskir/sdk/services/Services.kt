package com.taskir.sdk.services

import android.util.Log
import com.taskir.sdk.data.models.*
import com.taskir.sdk.network.RequestManager
import com.taskir.sdk.network.api.*
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

/**
 * Authentication Service
 * Handles user registration, login, and token management
 */
class AuthService(private val requestManager: RequestManager, private val debug: Boolean) {
    private val authApi = requestManager.retrofit.create(AuthApi::class.java)
    private var currentToken: String? = null
    private val tokenLock = Mutex()

    suspend fun register(email: String, password: String, name: String): AuthResponse {
        log("Registering user: $email")
        val request = RegisterRequest(email, password, name)
        val response = requestManager.executeWithRetry {
            authApi.register(request)
        }
        currentToken = response.token
        return response
    }

    suspend fun login(email: String, password: String): AuthResponse {
        log("Logging in user: $email")
        val request = LoginRequest(email, password)
        val response = requestManager.executeWithRetry {
            authApi.login(request)
        }
        tokenLock.withLock {
            currentToken = response.token
        }
        return response
    }

    suspend fun logout(): Unit {
        log("Logging out")
        requestManager.executeWithRetry {
            authApi.logout()
        }
        tokenLock.withLock {
            currentToken = null
        }
    }

    suspend fun getProfile(): User {
        log("Fetching user profile")
        return requestManager.executeWithRetry {
            authApi.getProfile()
        }
    }

    suspend fun refreshToken(): AuthResponse {
        log("Refreshing authentication token")
        val response = requestManager.executeWithRetry {
            authApi.refreshToken()
        }
        tokenLock.withLock {
            currentToken = response.token
        }
        return response
    }

    suspend fun getToken(): String? = tokenLock.withLock { currentToken }

    suspend fun setToken(token: String) {
        tokenLock.withLock {
            currentToken = token
        }
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("AuthService", message)
        }
    }
}

/**
 * Campaign Service
 * Handles campaign management operations
 */
class CampaignService(private val requestManager: RequestManager, private val debug: Boolean) {
    private val campaignApi = requestManager.retrofit.create(CampaignApi::class.java)

    suspend fun create(campaign: Campaign): Campaign {
        log("Creating campaign: ${campaign.name}")
        return requestManager.executeWithRetry {
            campaignApi.createCampaign(campaign)
        }
    }

    suspend fun list(skip: Int = 0, limit: Int = 20): List<Campaign> {
        log("Listing campaigns (skip=$skip, limit=$limit)")
        return requestManager.executeWithRetry {
            campaignApi.listCampaigns(skip, limit)
        }
    }

    suspend fun get(campaignId: String): Campaign {
        log("Getting campaign: $campaignId")
        return requestManager.executeWithRetry {
            campaignApi.getCampaign(campaignId)
        }
    }

    suspend fun update(campaignId: String, updates: Map<String, Any>): Campaign {
        log("Updating campaign: $campaignId")
        return requestManager.executeWithRetry {
            campaignApi.updateCampaign(campaignId, updates)
        }
    }

    suspend fun delete(campaignId: String): Unit {
        log("Deleting campaign: $campaignId")
        return requestManager.executeWithRetry {
            campaignApi.deleteCampaign(campaignId)
        }
    }

    suspend fun pause(campaignId: String): Campaign {
        log("Pausing campaign: $campaignId")
        return requestManager.executeWithRetry {
            campaignApi.pauseCampaign(campaignId)
        }
    }

    suspend fun resume(campaignId: String): Campaign {
        log("Resuming campaign: $campaignId")
        return requestManager.executeWithRetry {
            campaignApi.resumeCampaign(campaignId)
        }
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("CampaignService", message)
        }
    }
}

/**
 * Analytics Service
 * Handles analytics and reporting operations
 */
class AnalyticsService(private val requestManager: RequestManager, private val debug: Boolean) {
    private val analyticsApi = requestManager.retrofit.create(AnalyticsApi::class.java)

    suspend fun getRealtime(): Analytics {
        log("Fetching real-time analytics")
        return requestManager.executeWithRetry {
            analyticsApi.getRealtime()
        }
    }

    suspend fun getCampaignAnalytics(
        campaignId: String,
        startDate: String? = null,
        endDate: String? = null
    ): Analytics {
        log("Fetching campaign analytics: $campaignId")
        return requestManager.executeWithRetry {
            analyticsApi.getCampaignAnalytics(campaignId, startDate, endDate)
        }
    }

    suspend fun getBreakdown(campaignId: String, breakdownType: String): Map<String, Any> {
        log("Fetching $breakdownType breakdown for campaign: $campaignId")
        return requestManager.executeWithRetry {
            analyticsApi.getBreakdown(campaignId, breakdownType)
        }
    }

    suspend fun getDashboard(): Map<String, Any> {
        log("Fetching dashboard analytics")
        return requestManager.executeWithRetry {
            analyticsApi.getDashboard()
        }
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("AnalyticsService", message)
        }
    }
}

/**
 * Bidding Service
 * Handles bidding operations
 */
class BiddingService(private val requestManager: RequestManager, private val debug: Boolean) {
    private val biddingApi = requestManager.retrofit.create(BiddingApi::class.java)

    suspend fun submitBid(bid: Bid): Bid {
        log("Submitting bid for campaign: ${bid.campaignId}")
        return requestManager.executeWithRetry {
            biddingApi.submitBid(bid)
        }
    }

    suspend fun getRecommendations(campaignId: String): List<Map<String, Any>> {
        log("Fetching bid recommendations for campaign: $campaignId")
        return requestManager.executeWithRetry {
            biddingApi.getRecommendations(campaignId)
        }
    }

    suspend fun list(campaignId: String? = null, skip: Int = 0, limit: Int = 20): List<Bid> {
        log("Listing bids (skip=$skip, limit=$limit)")
        return requestManager.executeWithRetry {
            biddingApi.listBids(campaignId, skip, limit)
        }
    }

    suspend fun get(bidId: String): Bid {
        log("Getting bid: $bidId")
        return requestManager.executeWithRetry {
            biddingApi.getBid(bidId)
        }
    }

    suspend fun getStats(campaignId: String): Map<String, Any> {
        log("Fetching bid statistics for campaign: $campaignId")
        return requestManager.executeWithRetry {
            biddingApi.getStats(campaignId)
        }
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("BiddingService", message)
        }
    }
}

/**
 * Ad Service
 * Handles ad placement management
 */
class AdService(private val requestManager: RequestManager, private val debug: Boolean) {
    private val adApi = requestManager.retrofit.create(AdApi::class.java)

    suspend fun create(ad: Ad): Ad {
        log("Creating ad for campaign: ${ad.campaignId}")
        return requestManager.executeWithRetry {
            adApi.createAd(ad)
        }
    }

    suspend fun list(campaignId: String? = null, skip: Int = 0, limit: Int = 20): List<Ad> {
        log("Listing ads (skip=$skip, limit=$limit)")
        return requestManager.executeWithRetry {
            adApi.listAds(campaignId, skip, limit)
        }
    }

    suspend fun get(adId: String): Ad {
        log("Getting ad: $adId")
        return requestManager.executeWithRetry {
            adApi.getAd(adId)
        }
    }

    suspend fun update(adId: String, updates: Map<String, Any>): Ad {
        log("Updating ad: $adId")
        return requestManager.executeWithRetry {
            adApi.updateAd(adId, updates)
        }
    }

    suspend fun delete(adId: String): Unit {
        log("Deleting ad: $adId")
        return requestManager.executeWithRetry {
            adApi.deleteAd(adId)
        }
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("AdService", message)
        }
    }
}

/**
 * Webhook Service
 * Handles webhook subscriptions and event management
 */
class WebhookService(private val requestManager: RequestManager, private val debug: Boolean) {
    private val webhookApi = requestManager.retrofit.create(WebhookApi::class.java)
    private val eventHandlers = mutableMapOf<String, (WebhookEvent) -> Unit>()

    suspend fun subscribe(webhook: Webhook): Webhook {
        log("Subscribing to webhook: ${webhook.url}")
        return requestManager.executeWithRetry {
            webhookApi.subscribe(webhook)
        }
    }

    suspend fun list(skip: Int = 0, limit: Int = 20): List<Webhook> {
        log("Listing webhooks (skip=$skip, limit=$limit)")
        return requestManager.executeWithRetry {
            webhookApi.listWebhooks(skip, limit)
        }
    }

    suspend fun get(webhookId: String): Webhook {
        log("Getting webhook: $webhookId")
        return requestManager.executeWithRetry {
            webhookApi.getWebhook(webhookId)
        }
    }

    suspend fun update(webhookId: String, updates: Map<String, Any>): Webhook {
        log("Updating webhook: $webhookId")
        return requestManager.executeWithRetry {
            webhookApi.updateWebhook(webhookId, updates)
        }
    }

    suspend fun delete(webhookId: String): Unit {
        log("Deleting webhook: $webhookId")
        return requestManager.executeWithRetry {
            webhookApi.deleteWebhook(webhookId)
        }
    }

    suspend fun test(webhookId: String): Map<String, Any> {
        log("Testing webhook: $webhookId")
        return requestManager.executeWithRetry {
            webhookApi.testWebhook(webhookId)
        }
    }

    suspend fun getLogs(webhookId: String, limit: Int = 100): List<Map<String, Any>> {
        log("Fetching webhook logs: $webhookId")
        return requestManager.executeWithRetry {
            webhookApi.getWebhookLogs(webhookId, limit)
        }
    }

    fun onEvent(eventType: String, handler: (WebhookEvent) -> Unit) {
        log("Registering event handler for: $eventType")
        eventHandlers[eventType] = handler
    }

    fun offEvent(eventType: String) {
        log("Removing event handler for: $eventType")
        eventHandlers.remove(eventType)
    }

    fun handleEvent(event: WebhookEvent) {
        val handler = eventHandlers[event.type]
        if (handler != null) {
            log("Handling event: ${event.type}")
            handler(event)
        }
    }

    private fun log(message: String) {
        if (debug) {
            Log.d("WebhookService", message)
        }
    }
}
