package com.taskir.sdk.network.api

import com.taskir.sdk.data.models.*
import retrofit2.http.*

/**
 * Authentication API Interface
 */
interface AuthApi {
    @POST("/api/auth/register")
    suspend fun register(
        @Body request: RegisterRequest
    ): AuthResponse

    @POST("/api/auth/login")
    suspend fun login(
        @Body request: LoginRequest
    ): AuthResponse

    @POST("/api/auth/refresh")
    suspend fun refreshToken(): AuthResponse

    @POST("/api/auth/logout")
    suspend fun logout(): Unit

    @GET("/api/auth/profile")
    suspend fun getProfile(): User
}

data class RegisterRequest(
    val email: String,
    val password: String,
    val name: String
)

data class LoginRequest(
    val email: String,
    val password: String
)

/**
 * Campaign API Interface
 */
interface CampaignApi {
    @POST("/api/campaigns")
    suspend fun createCampaign(
        @Body campaign: Campaign
    ): Campaign

    @GET("/api/campaigns")
    suspend fun listCampaigns(
        @Query("skip") skip: Int = 0,
        @Query("limit") limit: Int = 20
    ): List<Campaign>

    @GET("/api/campaigns/{id}")
    suspend fun getCampaign(
        @Path("id") campaignId: String
    ): Campaign

    @PUT("/api/campaigns/{id}")
    suspend fun updateCampaign(
        @Path("id") campaignId: String,
        @Body updates: Map<String, Any>
    ): Campaign

    @DELETE("/api/campaigns/{id}")
    suspend fun deleteCampaign(
        @Path("id") campaignId: String
    ): Unit

    @POST("/api/campaigns/{id}/pause")
    suspend fun pauseCampaign(
        @Path("id") campaignId: String
    ): Campaign

    @POST("/api/campaigns/{id}/resume")
    suspend fun resumeCampaign(
        @Path("id") campaignId: String
    ): Campaign
}

/**
 * Analytics API Interface
 */
interface AnalyticsApi {
    @GET("/api/analytics/realtime")
    suspend fun getRealtime(): Analytics

    @GET("/api/analytics/campaigns/{id}")
    suspend fun getCampaignAnalytics(
        @Path("id") campaignId: String,
        @Query("start_date") startDate: String? = null,
        @Query("end_date") endDate: String? = null
    ): Analytics

    @GET("/api/analytics/campaigns/{id}/breakdown/{type}")
    suspend fun getBreakdown(
        @Path("id") campaignId: String,
        @Path("type") breakdownType: String
    ): Map<String, Any>

    @GET("/api/analytics/dashboard")
    suspend fun getDashboard(): Map<String, Any>
}

/**
 * Bidding API Interface
 */
interface BiddingApi {
    @POST("/api/bids")
    suspend fun submitBid(
        @Body bid: Bid
    ): Bid

    @GET("/api/bids/campaigns/{id}/recommendations")
    suspend fun getRecommendations(
        @Path("id") campaignId: String
    ): List<Map<String, Any>>

    @GET("/api/bids")
    suspend fun listBids(
        @Query("campaign_id") campaignId: String? = null,
        @Query("skip") skip: Int = 0,
        @Query("limit") limit: Int = 20
    ): List<Bid>

    @GET("/api/bids/{id}")
    suspend fun getBid(
        @Path("id") bidId: String
    ): Bid

    @GET("/api/bids/campaigns/{id}/stats")
    suspend fun getStats(
        @Path("id") campaignId: String
    ): Map<String, Any>
}

/**
 * Ad API Interface
 */
interface AdApi {
    @POST("/api/ads")
    suspend fun createAd(
        @Body ad: Ad
    ): Ad

    @GET("/api/ads")
    suspend fun listAds(
        @Query("campaign_id") campaignId: String? = null,
        @Query("skip") skip: Int = 0,
        @Query("limit") limit: Int = 20
    ): List<Ad>

    @GET("/api/ads/{id}")
    suspend fun getAd(
        @Path("id") adId: String
    ): Ad

    @PUT("/api/ads/{id}")
    suspend fun updateAd(
        @Path("id") adId: String,
        @Body updates: Map<String, Any>
    ): Ad

    @DELETE("/api/ads/{id}")
    suspend fun deleteAd(
        @Path("id") adId: String
    ): Unit
}

/**
 * Webhook API Interface
 */
interface WebhookApi {
    @POST("/api/webhooks")
    suspend fun subscribe(
        @Body webhook: Webhook
    ): Webhook

    @GET("/api/webhooks")
    suspend fun listWebhooks(
        @Query("skip") skip: Int = 0,
        @Query("limit") limit: Int = 20
    ): List<Webhook>

    @GET("/api/webhooks/{id}")
    suspend fun getWebhook(
        @Path("id") webhookId: String
    ): Webhook

    @PUT("/api/webhooks/{id}")
    suspend fun updateWebhook(
        @Path("id") webhookId: String,
        @Body updates: Map<String, Any>
    ): Webhook

    @DELETE("/api/webhooks/{id}")
    suspend fun deleteWebhook(
        @Path("id") webhookId: String
    ): Unit

    @POST("/api/webhooks/{id}/test")
    suspend fun testWebhook(
        @Path("id") webhookId: String
    ): Map<String, Any>

    @GET("/api/webhooks/{id}/logs")
    suspend fun getWebhookLogs(
        @Path("id") webhookId: String,
        @Query("limit") limit: Int = 100
    ): List<Map<String, Any>>
}
