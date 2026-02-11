package com.taskir.sdk.data.models

/**
 * Configuration for TaskirX Android SDK
 */
data class ClientConfig(
    val apiUrl: String,
    val apiKey: String,
    val debug: Boolean = false,
    val timeout: Long = 30000L,
    val retryAttempts: Int = 3
)

/**
 * Campaign data model
 */
data class Campaign(
    val id: String,
    val name: String,
    val budget: Double,
    val startDate: String,
    val endDate: String,
    val status: String = "active",
    val targetAudience: Map<String, Any>? = null,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

/**
 * Bid data model
 */
data class Bid(
    val id: String,
    val campaignId: String,
    val adSlotId: String,
    val amount: Double,
    val currency: String = "USD",
    val status: String = "pending",
    val createdAt: String? = null,
    val updatedAt: String? = null
)

/**
 * Analytics metrics
 */
data class Analytics(
    val impressions: Int = 0,
    val clicks: Int = 0,
    val conversions: Int = 0,
    val spend: Double = 0.0,
    val revenue: Double = 0.0,
    val ctr: Double = 0.0,
    val conversionRate: Double = 0.0,
    val roi: Double = 0.0,
    val timestamp: String? = null
)

/**
 * User/Profile data model
 */
data class User(
    val id: String,
    val email: String,
    val name: String,
    val company: String? = null,
    val role: String = "user",
    val createdAt: String? = null,
    val updatedAt: String? = null
)

/**
 * Authentication response
 */
data class AuthResponse(
    val token: String,
    val refreshToken: String? = null,
    val user: User? = null,
    val expiresIn: Long? = null
)

/**
 * Webhook data model
 */
data class Webhook(
    val id: String,
    val url: String,
    val events: List<String>,
    val active: Boolean = true,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

/**
 * Webhook event payload
 */
data class WebhookEvent(
    val id: String,
    val type: String,
    val data: Map<String, Any>,
    val timestamp: String
)

/**
 * Ad placement data model
 */
data class Ad(
    val id: String,
    val campaignId: String,
    val placement: String,
    val creativeUrl: String,
    val clickUrl: String,
    val width: Int,
    val height: Int,
    val status: String = "active",
    val createdAt: String? = null,
    val updatedAt: String? = null
)

/**
 * Error response model
 */
data class ErrorResponse(
    val code: String,
    val message: String,
    val details: Map<String, Any>? = null
)

/**
 * API response wrapper
 */
data class ApiResponse<T>(
    val success: Boolean,
    val data: T? = null,
    val error: ErrorResponse? = null
)
