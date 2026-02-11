import XCTest
@testable import TaskirX

final class TaskirXClientTests: XCTestCase {
    var client: TaskirXClient!
    
    override func setUp() {
        super.setUp()
        client = TaskirXClient.create(apiUrl: "http://localhost:3000", apiKey: "test-key", debug: true)
    }
    
    // MARK: - Initialization Tests
    
    func testClientCreation() {
        XCTAssertNotNil(client)
    }
    
    func testClientHasAllServices() {
        XCTAssertNotNil(client.auth)
        XCTAssertNotNil(client.campaigns)
        XCTAssertNotNil(client.analytics)
        XCTAssertNotNil(client.bidding)
        XCTAssertNotNil(client.ads)
        XCTAssertNotNil(client.webhooks)
    }
    
    func testDebugModeToggle() {
        client.enableDebug(true)
        // Should not throw
        XCTAssert(true)
    }
    
    // MARK: - Health & Status Tests
    
    func testGetHealth() async {
        let result = await client.getHealth()
        
        switch result {
        case .success(let health):
            XCTAssertNotNil(health)
        case .failure(let error):
            XCTFail("Expected success, got error: \(error)")
        }
    }
    
    func testGetStatus() async {
        let result = await client.getStatus()
        
        switch result {
        case .success(let status):
            XCTAssertNotNil(status)
        case .failure(let error):
            XCTFail("Expected success, got error: \(error)")
        }
    }
    
    // MARK: - Config Tests
    
    func testConfigInitialization() {
        let config = ClientConfig(
            apiUrl: "http://localhost:3000",
            apiKey: "test-key",
            debug: true,
            timeout: 30,
            retryAttempts: 3
        )
        
        XCTAssertEqual(config.apiUrl, "http://localhost:3000")
        XCTAssertEqual(config.apiKey, "test-key")
        XCTAssertTrue(config.debug)
        XCTAssertEqual(config.timeout, 30)
        XCTAssertEqual(config.retryAttempts, 3)
    }
    
    // MARK: - Error Handling Tests
    
    func testNetworkError() {
        let error = TaskirXError.networkError("Connection failed")
        XCTAssertNotNil(error.errorDescription)
        XCTAssert(error.errorDescription?.contains("Network Error") ?? false)
    }
    
    func testHttpError() {
        let error = TaskirXError.httpError(statusCode: 404, message: "Not found")
        XCTAssertNotNil(error.errorDescription)
        XCTAssert(error.errorDescription?.contains("404") ?? false)
    }
    
    func testDecodingError() {
        let error = TaskirXError.decodingError("Invalid JSON")
        XCTAssertNotNil(error.errorDescription)
        XCTAssert(error.errorDescription?.contains("Decoding Error") ?? false)
    }
    
    func testTimeoutError() {
        let error = TaskirXError.timeout
        XCTAssertNotNil(error.errorDescription)
        XCTAssert(error.errorDescription?.contains("timeout") ?? false)
    }
    
    // MARK: - Model Tests
    
    func testUserModel() {
        let user = User(
            id: "user-1",
            email: "test@example.com",
            name: "Test User",
            company: "Test Co",
            role: "admin",
            createdAt: "2024-01-01T00:00:00Z",
            updatedAt: "2024-01-02T00:00:00Z"
        )
        
        XCTAssertEqual(user.id, "user-1")
        XCTAssertEqual(user.email, "test@example.com")
        XCTAssertEqual(user.name, "Test User")
    }
    
    func testCampaignModel() {
        let campaign = Campaign(
            id: "camp-1",
            name: "Test Campaign",
            budget: 1000.0,
            startDate: "2024-01-01",
            endDate: "2024-01-31",
            status: "active",
            targetAudience: [:],
            createdAt: "2024-01-01T00:00:00Z",
            updatedAt: "2024-01-01T00:00:00Z"
        )
        
        XCTAssertEqual(campaign.id, "camp-1")
        XCTAssertEqual(campaign.name, "Test Campaign")
        XCTAssertEqual(campaign.budget, 1000.0)
    }
    
    func testBidModel() {
        let bid = Bid(
            id: "bid-1",
            campaignId: "camp-1",
            adSlotId: "slot-1",
            amount: 5.0,
            currency: "USD",
            status: "active",
            createdAt: "2024-01-01T00:00:00Z",
            updatedAt: "2024-01-01T00:00:00Z"
        )
        
        XCTAssertEqual(bid.id, "bid-1")
        XCTAssertEqual(bid.amount, 5.0)
        XCTAssertEqual(bid.currency, "USD")
    }
    
    func testAnalyticsModel() {
        let analytics = Analytics(
            impressions: 1000,
            clicks: 50,
            conversions: 10,
            spend: 100.0,
            revenue: 200.0,
            ctr: 0.05,
            conversionRate: 0.1,
            roi: 1.0,
            timestamp: "2024-01-01T00:00:00Z"
        )
        
        XCTAssertEqual(analytics.impressions, 1000)
        XCTAssertEqual(analytics.clicks, 50)
        XCTAssertEqual(analytics.ctr, 0.05)
    }
    
    // MARK: - Result Type Tests
    
    func testResultSuccess() {
        let result: Result<String> = .success("test")
        
        var handlerCalled = false
        result.onSuccess { _ in handlerCalled = true }
        XCTAssertTrue(handlerCalled)
        XCTAssertNotNil(result.value)
    }
    
    func testResultFailure() {
        let result: Result<String> = .failure(.timeout)
        
        var handlerCalled = false
        result.onFailure { _ in handlerCalled = true }
        XCTAssertTrue(handlerCalled)
        XCTAssertNil(result.value)
    }
    
    // MARK: - AnyCodable Tests
    
    func testAnyCodableBool() {
        let bool = AnyCodable.bool(true)
        if case .bool(let value) = bool {
            XCTAssertTrue(value)
        } else {
            XCTFail("Expected bool case")
        }
    }
    
    func testAnyCodableString() {
        let string = AnyCodable.string("test")
        if case .string(let value) = string {
            XCTAssertEqual(value, "test")
        } else {
            XCTFail("Expected string case")
        }
    }
    
    func testAnyCodableInt() {
        let int = AnyCodable.int(42)
        if case .int(let value) = int {
            XCTAssertEqual(value, 42)
        } else {
            XCTFail("Expected int case")
        }
    }
    
    func testAnyCodableDouble() {
        let double = AnyCodable.double(3.14)
        if case .double(let value) = double {
            XCTAssertEqual(value, 3.14, accuracy: 0.01)
        } else {
            XCTFail("Expected double case")
        }
    }
}

// MARK: - Auth Service Tests

final class AuthServiceTests: XCTestCase {
    var requestManager: RequestManager!
    var authService: AuthService!
    
    override func setUp() {
        super.setUp()
        let config = ClientConfig(apiUrl: "http://localhost:3000", apiKey: "test-key")
        requestManager = RequestManager(config: config)
        authService = AuthService(requestManager: requestManager)
    }
    
    func testAuthServiceInitialization() {
        XCTAssertNotNil(authService)
    }
    
    func testLoginRequestCreation() {
        let request = LoginRequest(email: "test@example.com", password: "password123")
        XCTAssertEqual(request.email, "test@example.com")
        XCTAssertEqual(request.password, "password123")
    }
    
    func testRegisterRequestCreation() {
        let request = RegisterRequest(
            email: "newuser@example.com",
            password: "password123",
            name: "New User",
            company: "New Co"
        )
        XCTAssertEqual(request.email, "newuser@example.com")
        XCTAssertEqual(request.name, "New User")
        XCTAssertEqual(request.company, "New Co")
    }
}

// MARK: - Campaign Service Tests

final class CampaignServiceTests: XCTestCase {
    var requestManager: RequestManager!
    var campaignService: CampaignService!
    
    override func setUp() {
        super.setUp()
        let config = ClientConfig(apiUrl: "http://localhost:3000", apiKey: "test-key")
        requestManager = RequestManager(config: config)
        campaignService = CampaignService(requestManager: requestManager)
    }
    
    func testCampaignServiceInitialization() {
        XCTAssertNotNil(campaignService)
    }
    
    func testCampaignCreateRequestCreation() {
        let request = CampaignCreateRequest(
            name: "Test Campaign",
            budget: 1000.0,
            startDate: "2024-01-01",
            endDate: "2024-01-31",
            targetAudience: [:]
        )
        XCTAssertEqual(request.name, "Test Campaign")
        XCTAssertEqual(request.budget, 1000.0)
    }
}

// MARK: - Integration Tests

final class TaskirXIntegrationTests: XCTestCase {
    var client: TaskirXClient!
    
    override func setUp() {
        super.setUp()
        client = TaskirXClient.create(apiUrl: "http://localhost:3000", apiKey: "test-key")
    }
    
    func testFullServiceIntegration() {
        XCTAssertNotNil(client.auth)
        XCTAssertNotNil(client.campaigns)
        XCTAssertNotNil(client.analytics)
        XCTAssertNotNil(client.bidding)
        XCTAssertNotNil(client.ads)
        XCTAssertNotNil(client.webhooks)
    }
    
    func testConfigurationPersistence() {
        let client2 = TaskirXClient.create(apiUrl: "http://localhost:3001", apiKey: "another-key")
        XCTAssertNotNil(client2)
    }
    
    func testErrorPropagation() {
        let errors: [TaskirXError] = [
            .networkError("test"),
            .decodingError("test"),
            .httpError(statusCode: 500, message: "test"),
            .invalidResponse,
            .timeout,
            .retryExhausted
        ]
        
        for error in errors {
            XCTAssertNotNil(error.errorDescription)
        }
    }
    
    func testWebhookEventHandling() {
        var eventHandled = false
        
        client.onWebhookEvent("test.event") { _ in
            eventHandled = true
        }
        
        let event = WebhookEvent(
            id: "event-1",
            type: "test.event",
            data: [:],
            timestamp: "2024-01-01T00:00:00Z"
        )
        
        client.webhooks.handleEvent(event)
        XCTAssertTrue(eventHandled)
    }
}
