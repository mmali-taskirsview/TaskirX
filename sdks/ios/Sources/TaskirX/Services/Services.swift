import Foundation

// MARK: - Auth Service

public class AuthService {
    private let requestManager: RequestManager
    private var currentUser: User?
    private let userLock = NSLock()
    
    init(requestManager: RequestManager) {
        self.requestManager = requestManager
    }
    
    public func register(email: String, password: String, name: String, company: String? = nil) async throws -> AuthResponse {
        let request = RegisterRequest(email: email, password: password, name: name, company: company)
        let response: AuthResponse = try await requestManager.post("/auth/register", body: request)
        
        userLock.lock()
        defer { userLock.unlock() }
        currentUser = response.user
        requestManager.setAuthToken(response.token)
        
        return response
    }
    
    public func login(email: String, password: String) async throws -> AuthResponse {
        let request = LoginRequest(email: email, password: password)
        let response: AuthResponse = try await requestManager.post("/auth/login", body: request)
        
        userLock.lock()
        defer { userLock.unlock() }
        currentUser = response.user
        requestManager.setAuthToken(response.token)
        
        return response
    }
    
    public func logout() async throws -> [String: AnyCodable] {
        requestManager.clearAuthToken()
        userLock.lock()
        defer { userLock.unlock() }
        currentUser = nil
        return [:]
    }
    
    public func getProfile() async throws -> User {
        let user: User = try await requestManager.get("/auth/profile")
        userLock.lock()
        defer { userLock.unlock() }
        currentUser = user
        return user
    }
    
    public func refreshToken(refreshToken: String) async throws -> AuthResponse {
        let request = ["refreshToken": AnyCodable.string(refreshToken)]
        let response: AuthResponse = try await requestManager.post("/auth/refresh", body: request)
        requestManager.setAuthToken(response.token)
        return response
    }
}

// MARK: - Campaign Service

public class CampaignService {
    private let requestManager: RequestManager
    
    init(requestManager: RequestManager) {
        self.requestManager = requestManager
    }
    
    public func create(name: String, budget: Double, startDate: String, 
                      endDate: String, targetAudience: [String: AnyCodable]) async throws -> Campaign {
        let request = CampaignCreateRequest(name: name, budget: budget, startDate: startDate, 
                                           endDate: endDate, targetAudience: targetAudience)
        return try await requestManager.post("/campaigns", body: request)
    }
    
    public func list(limit: Int = 50, offset: Int = 0) async throws -> [Campaign] {
        let endpoint = "/campaigns?limit=\(limit)&offset=\(offset)"
        return try await requestManager.get(endpoint)
    }
    
    public func get(id: String) async throws -> Campaign {
        return try await requestManager.get("/campaigns/\(id)")
    }
    
    public func update(id: String, name: String? = nil, budget: Double? = nil) async throws -> Campaign {
        var updates: [String: AnyCodable] = [:]
        if let name = name { updates["name"] = .string(name) }
        if let budget = budget { updates["budget"] = .double(budget) }
        return try await requestManager.put("/campaigns/\(id)", body: updates)
    }
    
    public func delete(id: String) async throws -> [String: AnyCodable] {
        return try await requestManager.delete("/campaigns/\(id)")
    }
    
    public func pause(id: String) async throws -> Campaign {
        return try await requestManager.put("/campaigns/\(id)/pause")
    }
    
    public func resume(id: String) async throws -> Campaign {
        return try await requestManager.put("/campaigns/\(id)/resume")
    }
}

// MARK: - Analytics Service

public class AnalyticsService {
    private let requestManager: RequestManager
    
    init(requestManager: RequestManager) {
        self.requestManager = requestManager
    }
    
    public func realtime() async throws -> Analytics {
        return try await requestManager.get("/analytics/realtime")
    }
    
    public func campaign(id: String) async throws -> Analytics {
        return try await requestManager.get("/analytics/campaigns/\(id)")
    }
    
    public func breakdown(type: String) async throws -> [[String: AnyCodable]] {
        return try await requestManager.get("/analytics/breakdown?type=\(type)")
    }
    
    public func dashboard() async throws -> [String: AnyCodable] {
        return try await requestManager.get("/analytics/dashboard")
    }
}

// MARK: - Bidding Service

public class BiddingService {
    private let requestManager: RequestManager
    
    init(requestManager: RequestManager) {
        self.requestManager = requestManager
    }
    
    public func submitBid(campaignId: String, adSlotId: String, amount: Double, currency: String = "USD") async throws -> Bid {
        let request = BidSubmitRequest(campaignId: campaignId, adSlotId: adSlotId, amount: amount, currency: currency)
        return try await requestManager.post("/bids", body: request)
    }
    
    public func recommendations() async throws -> [[String: AnyCodable]] {
        return try await requestManager.get("/bids/recommendations")
    }
    
    public func list(limit: Int = 50) async throws -> [Bid] {
        return try await requestManager.get("/bids?limit=\(limit)")
    }
    
    public func get(id: String) async throws -> Bid {
        return try await requestManager.get("/bids/\(id)")
    }
    
    public func stats() async throws -> [String: AnyCodable] {
        return try await requestManager.get("/bids/stats")
    }
}

// MARK: - Ad Service

public class AdService {
    private let requestManager: RequestManager
    
    init(requestManager: RequestManager) {
        self.requestManager = requestManager
    }
    
    public func create(campaignId: String, placement: String, imageUrl: String, 
                      clickUrl: String, dimensions: String) async throws -> Ad {
        let request = AdCreateRequest(campaignId: campaignId, placement: placement, 
                                     imageUrl: imageUrl, clickUrl: clickUrl, dimensions: dimensions)
        return try await requestManager.post("/ads", body: request)
    }
    
    public func list(campaignId: String, limit: Int = 50) async throws -> [Ad] {
        return try await requestManager.get("/ads?campaignId=\(campaignId)&limit=\(limit)")
    }
    
    public func get(id: String) async throws -> Ad {
        return try await requestManager.get("/ads/\(id)")
    }
    
    public func update(id: String, placement: String? = nil) async throws -> Ad {
        var updates: [String: AnyCodable] = [:]
        if let placement = placement { updates["placement"] = .string(placement) }
        return try await requestManager.put("/ads/\(id)", body: updates)
    }
    
    public func delete(id: String) async throws -> [String: AnyCodable] {
        return try await requestManager.delete("/ads/\(id)")
    }
}

// MARK: - Webhook Service

public class WebhookService {
    private let requestManager: RequestManager
    private var eventHandlers: [String: [(WebhookEvent) -> Void]] = [:]
    private let handlerLock = NSLock()
    
    init(requestManager: RequestManager) {
        self.requestManager = requestManager
    }
    
    public func subscribe(url: String, events: [String]) async throws -> Webhook {
        let request = WebhookCreateRequest(url: url, events: events)
        return try await requestManager.post("/webhooks", body: request)
    }
    
    public func list(limit: Int = 50) async throws -> [Webhook] {
        return try await requestManager.get("/webhooks?limit=\(limit)")
    }
    
    public func get(id: String) async throws -> Webhook {
        return try await requestManager.get("/webhooks/\(id)")
    }
    
    public func update(id: String, active: Bool? = nil) async throws -> Webhook {
        var updates: [String: AnyCodable] = [:]
        if let active = active { updates["active"] = .bool(active) }
        return try await requestManager.put("/webhooks/\(id)", body: updates)
    }
    
    public func delete(id: String) async throws -> [String: AnyCodable] {
        return try await requestManager.delete("/webhooks/\(id)")
    }
    
    public func test(id: String) async throws -> [String: AnyCodable] {
        return try await requestManager.post("/webhooks/\(id)/test")
    }
    
    public func getLogs(id: String, limit: Int = 50) async throws -> [[String: AnyCodable]] {
        return try await requestManager.get("/webhooks/\(id)/logs?limit=\(limit)")
    }
    
    public func onEvent(_ type: String, handler: @escaping (WebhookEvent) -> Void) {
        handlerLock.lock()
        defer { handlerLock.unlock() }
        if eventHandlers[type] == nil {
            eventHandlers[type] = []
        }
        eventHandlers[type]?.append(handler)
    }
    
    public func offEvent(_ type: String) {
        handlerLock.lock()
        defer { handlerLock.unlock() }
        eventHandlers[type] = nil
    }
    
    public func handleEvent(_ event: WebhookEvent) {
        handlerLock.lock()
        defer { handlerLock.unlock() }
        eventHandlers[event.type]?.forEach { $0(event) }
    }
}
