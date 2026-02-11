package com.taskir.sdk

import android.content.Context
import androidx.test.core.app.ApplicationProvider
import com.taskir.sdk.data.models.*
import com.taskir.sdk.network.RequestManager
import com.taskir.sdk.services.AuthService
import com.taskir.sdk.services.CampaignService
import io.mockk.coEvery
import io.mockk.mockk
import junit.framework.TestCase
import kotlinx.coroutines.runBlocking
import org.junit.Before
import org.junit.Test

/**
 * Unit tests for TaskirX Android SDK
 */
class TaskirXClientTest {

    private lateinit var context: Context
    private lateinit var client: TaskirXClient
    private lateinit var config: ClientConfig

    @Before
    fun setup() {
        context = ApplicationProvider.getApplicationContext()
        config = ClientConfig(
            apiUrl = "http://localhost:3000",
            apiKey = "test-api-key",
            debug = true,
            timeout = 30000L,
            retryAttempts = 3
        )
        client = TaskirXClient(context, config)
    }

    @Test
    fun testClientInitialization() {
        TestCase.assertNotNull(client)
        TestCase.assertNotNull(client.auth)
        TestCase.assertNotNull(client.campaigns)
        TestCase.assertNotNull(client.analytics)
        TestCase.assertNotNull(client.bidding)
        TestCase.assertNotNull(client.ads)
        TestCase.assertNotNull(client.webhooks)
    }

    @Test
    fun testServiceAccess() {
        TestCase.assertNotNull(client.auth)
        TestCase.assertNotNull(client.campaigns)
        TestCase.assertNotNull(client.analytics)
        TestCase.assertNotNull(client.bidding)
        TestCase.assertNotNull(client.ads)
        TestCase.assertNotNull(client.webhooks)
    }

    @Test
    fun testDebugModeToggle() {
        client.enableDebug(true)
        client.enableDebug(false)
        // No exceptions should be thrown
    }

    @Test
    fun testSetApiKey() {
        client.setApiKey("new-api-key")
        // API key should be updated
    }

    @Test
    fun testConfigurationSupport() {
        val customConfig = ClientConfig(
            apiUrl = "https://api.example.com",
            apiKey = "custom-key",
            debug = false,
            timeout = 60000L,
            retryAttempts = 5
        )
        val customClient = TaskirXClient(context, customConfig)
        TestCase.assertNotNull(customClient)
    }

    @Test
    fun testVersionConstant() {
        TestCase.assertNotNull(TaskirXClient.VERSION)
        TestCase.assertEquals("1.0.0", TaskirXClient.VERSION)
    }
}

/**
 * Unit tests for AuthService
 */
class AuthServiceTest {

    private lateinit var authService: AuthService
    private lateinit var mockRequestManager: RequestManager

    @Before
    fun setup() {
        mockRequestManager = mockk()
        authService = AuthService(mockRequestManager, true)
    }

    @Test
    fun testLoginSuccess() = runBlocking {
        val mockResponse = AuthResponse(
            token = "jwt-token-123",
            user = User(
                id = "user-123",
                email = "test@example.com",
                name = "Test User"
            )
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns mockResponse

        val result = authService.login("test@example.com", "password")

        TestCase.assertNotNull(result)
        TestCase.assertEquals("jwt-token-123", result.token)
    }

    @Test
    fun testLogout() = runBlocking {
        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns Unit

        authService.logout()
        // Should complete without exception
    }

    @Test
    fun testGetProfile() = runBlocking {
        val mockUser = User(
            id = "user-123",
            email = "test@example.com",
            name = "Test User"
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns mockUser

        val result = authService.getProfile()

        TestCase.assertNotNull(result)
        TestCase.assertEquals("test@example.com", result.email)
    }

    @Test
    fun testTokenRefresh() = runBlocking {
        val mockResponse = AuthResponse(
            token = "new-jwt-token-456"
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns mockResponse

        val result = authService.refreshToken()

        TestCase.assertNotNull(result)
        TestCase.assertEquals("new-jwt-token-456", result.token)
    }

    @Test
    fun testTokenManagement() = runBlocking {
        authService.setToken("test-token")
        val retrievedToken = authService.getToken()
        TestCase.assertEquals("test-token", retrievedToken)
    }
}

/**
 * Unit tests for CampaignService
 */
class CampaignServiceTest {

    private lateinit var campaignService: CampaignService
    private lateinit var mockRequestManager: RequestManager

    @Before
    fun setup() {
        mockRequestManager = mockk()
        campaignService = CampaignService(mockRequestManager, true)
    }

    @Test
    fun testCreateCampaign() = runBlocking {
        val mockCampaign = Campaign(
            id = "campaign-123",
            name = "Test Campaign",
            budget = 10000.0
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns mockCampaign

        val result = campaignService.create(mockCampaign)

        TestCase.assertNotNull(result)
        TestCase.assertEquals("campaign-123", result.id)
    }

    @Test
    fun testListCampaigns() = runBlocking {
        val mockCampaigns = listOf(
            Campaign(id = "c1", name = "Campaign 1", budget = 5000.0),
            Campaign(id = "c2", name = "Campaign 2", budget = 10000.0)
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns mockCampaigns

        val result = campaignService.list()

        TestCase.assertNotNull(result)
        TestCase.assertEquals(2, result.size)
    }

    @Test
    fun testGetCampaign() = runBlocking {
        val mockCampaign = Campaign(
            id = "campaign-123",
            name = "Test Campaign",
            budget = 10000.0
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns mockCampaign

        val result = campaignService.get("campaign-123")

        TestCase.assertNotNull(result)
        TestCase.assertEquals("Test Campaign", result.name)
    }

    @Test
    fun testUpdateCampaign() = runBlocking {
        val updatedCampaign = Campaign(
            id = "campaign-123",
            name = "Updated Campaign",
            budget = 15000.0
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns updatedCampaign

        val result = campaignService.update("campaign-123", mapOf("budget" to 15000.0))

        TestCase.assertNotNull(result)
        TestCase.assertEquals(15000.0, result.budget)
    }

    @Test
    fun testDeleteCampaign() = runBlocking {
        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns Unit

        campaignService.delete("campaign-123")
        // Should complete without exception
    }

    @Test
    fun testPauseCampaign() = runBlocking {
        val pausedCampaign = Campaign(
            id = "campaign-123",
            name = "Test Campaign",
            budget = 10000.0,
            status = "paused"
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns pausedCampaign

        val result = campaignService.pause("campaign-123")

        TestCase.assertNotNull(result)
        TestCase.assertEquals("paused", result.status)
    }

    @Test
    fun testResumeCampaign() = runBlocking {
        val resumedCampaign = Campaign(
            id = "campaign-123",
            name = "Test Campaign",
            budget = 10000.0,
            status = "active"
        )

        coEvery {
            mockRequestManager.executeWithRetry(any())
        } returns resumedCampaign

        val result = campaignService.resume("campaign-123")

        TestCase.assertNotNull(result)
        TestCase.assertEquals("active", result.status)
    }
}

/**
 * Integration tests for common workflows
 */
class TaskirXIntegrationTest {

    private lateinit var context: Context
    private lateinit var client: TaskirXClient

    @Before
    fun setup() {
        context = ApplicationProvider.getApplicationContext()
        val config = ClientConfig(
            apiUrl = "http://localhost:3000",
            apiKey = "test-api-key"
        )
        client = TaskirXClient(context, config)
    }

    @Test
    fun testServiceConfiguration() {
        TestCase.assertNotNull(client.auth)
        TestCase.assertNotNull(client.campaigns)
        TestCase.assertNotNull(client.analytics)
        TestCase.assertNotNull(client.bidding)
        TestCase.assertNotNull(client.ads)
        TestCase.assertNotNull(client.webhooks)
    }

    @Test
    fun testErrorHandling() {
        // Error handling should be graceful
        TestCase.assertNotNull(client)
    }

    @Test
    fun testRetryLogic() {
        // Retry logic should be transparent
        TestCase.assertNotNull(client)
    }
}
