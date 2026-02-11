package com.taskir.sdk.example

import android.content.Context
import androidx.lifecycle.lifecycleScope
import androidx.appcompat.app.AppCompatActivity
import com.taskir.sdk.TaskirXClient
import com.taskir.sdk.data.models.Campaign
import com.taskir.sdk.data.models.ClientConfig
import kotlinx.coroutines.launch

/**
 * TaskirX Android SDK - Complete Example
 * Demonstrates all major features and use cases
 */
class TaskirXExampleActivity : AppCompatActivity() {

    private lateinit var client: TaskirXClient

    override fun onCreate(savedInstanceState: android.os.Bundle?) {
        super.onCreate(savedInstanceState)

        // Initialize the client
        val config = ClientConfig(
            apiUrl = "https://api.taskir.io",
            apiKey = System.getenv("TASKIR_API_KEY") ?: "your-api-key",
            debug = true,
            timeout = 30000L,
            retryAttempts = 3
        )

        client = TaskirXClient.create(this, config)

        // Run examples
        runExamples()
    }

    private fun runExamples() {
        lifecycleScope.launch {
            try {
                println("=== TaskirX Android SDK Examples ===\n")

                // 1. Platform health check
                println("1. Platform Health Check")
                val health = client.getHealth()
                health.onSuccess {
                    println("✓ Platform is healthy: $it")
                }

                // 2. Authentication
                println("\n2. Authentication Examples")

                // Register
                val registerResult = client.auth.register(
                    "developer@example.com",
                    "SecurePassword123!",
                    "Developer User"
                )
                println("✓ User registered")

                // Login
                val loginResult = client.auth.login(
                    "developer@example.com",
                    "SecurePassword123!"
                )
                loginResult.onSuccess {
                    println("✓ User logged in, token: ${it.token.take(20)}...")
                }

                // Get profile
                val profile = client.getProfile()
                profile.onSuccess { user ->
                    println("✓ User profile: ${user.name} (${user.email})")
                }

                // 3. Campaign Management
                println("\n3. Campaign Management Examples")

                // Create campaign
                val newCampaign = Campaign(
                    id = "",
                    name = "Summer Sale 2024",
                    budget = 50000.0,
                    startDate = "2024-06-01",
                    endDate = "2024-08-31"
                )

                val campaignResult = client.campaigns.create(newCampaign)
                var campaignId = ""
                campaignResult.onSuccess { campaign ->
                    campaignId = campaign.id
                    println("✓ Campaign created: ${campaign.id} (${campaign.name})")
                }

                // List campaigns
                val campaigns = client.campaigns.list()
                println("✓ Total campaigns: ${campaigns.size}")

                if (campaignId.isNotEmpty()) {
                    // Get campaign details
                    val campaignDetails = client.campaigns.get(campaignId)
                    println("✓ Campaign budget: ${campaignDetails.budget}")

                    // Update campaign
                    val updateResult = client.campaigns.update(
                        campaignId,
                        mapOf("budget" to 75000.0)
                    )
                    println("✓ Campaign updated")

                    // 4. Analytics
                    println("\n4. Analytics Examples")

                    // Real-time analytics
                    val realtime = client.analytics.getRealtime()
                    println("✓ Real-time analytics:")
                    println("  - Impressions: ${realtime.impressions}")
                    println("  - Clicks: ${realtime.clicks}")
                    println("  - Conversions: ${realtime.conversions}")
                    println("  - CTR: ${(realtime.ctr * 100).toInt()}%")

                    // Campaign analytics
                    val campaignAnalytics = client.analytics.getCampaignAnalytics(
                        campaignId,
                        "2024-01-01",
                        "2024-12-31"
                    )
                    println("✓ Campaign analytics retrieved")

                    // Device breakdown
                    val deviceBreakdown = client.analytics.getBreakdown(
                        campaignId,
                        "device"
                    )
                    println("✓ Device breakdown retrieved")

                    // 5. Ad Management
                    println("\n5. Ad Management Examples")

                    // Create ad
                    val adResult = client.ads.create(
                        com.taskir.sdk.data.models.Ad(
                            id = "",
                            campaignId = campaignId,
                            placement = "homepage-banner",
                            creativeUrl = "https://example.com/ads/summer-sale.html",
                            clickUrl = "https://example.com/summer-sale",
                            width = 728,
                            height = 90
                        )
                    )
                    var adId = ""
                    adResult.onSuccess { ad ->
                        adId = ad.id
                        println("✓ Ad created: $adId")
                    }

                    // List ads
                    val ads = client.ads.list(campaignId)
                    println("✓ Total ads: ${ads.size}")

                    // 6. Bidding Engine
                    println("\n6. Bidding Engine Examples")

                    // Submit bid
                    val bidResult = client.bidding.submitBid(
                        com.taskir.sdk.data.models.Bid(
                            id = "",
                            campaignId = campaignId,
                            adSlotId = "slot-premium-001",
                            amount = 2.5,
                            currency = "USD"
                        )
                    )
                    bidResult.onSuccess { bid ->
                        println("✓ Bid submitted: ${bid.id} (\$${bid.amount})")
                    }

                    // Get recommendations
                    val recommendations = client.bidding.getRecommendations(campaignId)
                    println("✓ Bid recommendations: ${recommendations.size} found")

                    // Get bid stats
                    val bidStats = client.bidding.getStats(campaignId)
                    println("✓ Bid statistics retrieved")

                    // 7. Webhook Management
                    println("\n7. Webhook Management Examples")

                    // Subscribe to webhook
                    val webhookResult = client.webhooks.subscribe(
                        com.taskir.sdk.data.models.Webhook(
                            id = "",
                            url = "https://example.com/webhooks/taskir",
                            events = listOf(
                                "campaign.created",
                                "campaign.updated",
                                "bid.won",
                                "conversion.recorded"
                            ),
                            active = true
                        )
                    )
                    var webhookId = ""
                    webhookResult.onSuccess { webhook ->
                        webhookId = webhook.id
                        println("✓ Webhook subscribed: $webhookId")
                    }

                    // List webhooks
                    val webhooks = client.webhooks.list()
                    println("✓ Total webhooks: ${webhooks.size}")

                    // Register event handler
                    client.webhooks.onEvent("conversion.recorded") { event ->
                        println("✓ Conversion recorded: ${event.data}")
                    }

                    // Test webhook
                    if (webhookId.isNotEmpty()) {
                        val testResult = client.webhooks.test(webhookId)
                        println("✓ Webhook test completed")
                    }

                    // 8. Advanced Operations
                    println("\n8. Advanced Operations Examples")

                    // Dashboard
                    val dashboardResult = client.getDashboard()
                    dashboardResult.onSuccess { dashboard ->
                        println("✓ Dashboard retrieved with campaigns")
                    }

                    // Campaign performance
                    val performanceResult = client.getCampaignPerformance(campaignId)
                    performanceResult.onSuccess { performance ->
                        println("✓ Campaign performance retrieved")
                    }

                    // Batch create campaigns
                    val batchResult = client.createCampaigns(
                        listOf(
                            Campaign(
                                id = "",
                                name = "Spring Sale",
                                budget = 30000.0,
                                startDate = "2024-03-01",
                                endDate = "2024-05-31"
                            ),
                            Campaign(
                                id = "",
                                name = "Fall Sale",
                                budget = 40000.0,
                                startDate = "2024-09-01",
                                endDate = "2024-11-30"
                            )
                        )
                    )
                    batchResult.onSuccess { created ->
                        println("✓ Batch created ${created.size} campaigns")
                    }

                    // Statistics
                    val statsResult = client.getStatistics()
                    statsResult.onSuccess { stats ->
                        println("✓ Platform statistics retrieved")
                    }

                    // 9. Campaign Control
                    println("\n9. Campaign Control Examples")

                    // Pause campaign
                    val pauseResult = client.campaigns.pause(campaignId)
                    println("✓ Campaign paused")

                    // Resume campaign
                    val resumeResult = client.campaigns.resume(campaignId)
                    println("✓ Campaign resumed")

                    // 10. Debug Features
                    println("\n10. Debug Features")
                    client.enableDebug(true)
                    println("✓ Debug mode enabled")
                    client.enableDebug(false)
                    println("✓ Debug mode disabled")

                    // 11. Cleanup
                    println("\n11. Cleanup")

                    if (adId.isNotEmpty()) {
                        client.ads.delete(adId)
                        println("✓ Ad deleted")
                    }

                    client.campaigns.delete(campaignId)
                    println("✓ Campaign deleted")

                    client.logout()
                    println("✓ User logged out")
                }

                println("\n=== All Examples Completed Successfully ===")

            } catch (e: Exception) {
                println("Error: ${e.message}")
                e.printStackTrace()
            }
        }
    }
}
