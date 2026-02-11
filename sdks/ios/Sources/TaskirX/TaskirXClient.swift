import Foundation

// MARK: - Result Type

public enum Result<T> {
    case success(T)
    case failure(TaskirXError)
    
    public func onSuccess(_ handler: (T) -> Void) {
        if case .success(let value) = self {
            handler(value)
        }
    }
    
    public func onFailure(_ handler: (TaskirXError) -> Void) {
        if case .failure(let error) = self {
            handler(error)
        }
    }
    
    public var value: T? {
        if case .success(let value) = self {
            return value
        }
        return nil
    }
    
    public var error: TaskirXError? {
        if case .failure(let error) = self {
            return error
        }
        return nil
    }
}

// MARK: - Main TaskirX Client

public class TaskirXClient {
    private let config: ClientConfig
    private let requestManager: RequestManager
    
    public let auth: AuthService
    public let campaigns: CampaignService
    public let analytics: AnalyticsService
    public let bidding: BiddingService
    public let ads: AdService
    public let webhooks: WebhookService
    
    private init(config: ClientConfig) {
        self.config = config
        self.requestManager = RequestManager(config: config)
        
        self.auth = AuthService(requestManager: requestManager)
        self.campaigns = CampaignService(requestManager: requestManager)
        self.analytics = AnalyticsService(requestManager: requestManager)
        self.bidding = BiddingService(requestManager: requestManager)
        self.ads = AdService(requestManager: requestManager)
        self.webhooks = WebhookService(requestManager: requestManager)
    }
    
    // MARK: - Factory Method
    
    public static func create(apiUrl: String, apiKey: String, debug: Bool = false) -> TaskirXClient {
        let config = ClientConfig(apiUrl: apiUrl, apiKey: apiKey, debug: debug)
        return TaskirXClient(config: config)
    }
    
    // MARK: - Health & Status
    
    public func getHealth() async -> Result<[String: AnyCodable]> {
        do {
            let health: [String: AnyCodable] = try await requestManager.get("/health")
            return .success(health)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getStatus() async -> Result<[String: AnyCodable]> {
        do {
            let status: [String: AnyCodable] = try await requestManager.get("/status")
            return .success(status)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Profile Operations
    
    public func getProfile() async -> Result<User> {
        do {
            let user = try await auth.getProfile()
            return .success(user)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func logout() async -> Result<[String: AnyCodable]> {
        do {
            let result = try await auth.logout()
            return .success(result)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Dashboard & Analytics
    
    public func getDashboard() async -> Result<[String: AnyCodable]> {
        do {
            let dashboard = try await analytics.dashboard()
            return .success(dashboard)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getCampaignPerformance(campaignId: String) async -> Result<Analytics> {
        do {
            let analytics = try await analytics.campaign(id: campaignId)
            return .success(analytics)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getRealtimeAnalytics() async -> Result<Analytics> {
        do {
            let analytics = try await analytics.realtime()
            return .success(analytics)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Campaign Operations
    
    public func createCampaign(name: String, budget: Double, startDate: String, 
                              endDate: String, targetAudience: [String: AnyCodable]) async -> Result<Campaign> {
        do {
            let campaign = try await campaigns.create(name: name, budget: budget, startDate: startDate, 
                                                     endDate: endDate, targetAudience: targetAudience)
            return .success(campaign)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getCampaigns(limit: Int = 50, offset: Int = 0) async -> Result<[Campaign]> {
        do {
            let campaigns = try await campaigns.list(limit: limit, offset: offset)
            return .success(campaigns)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getCampaign(id: String) async -> Result<Campaign> {
        do {
            let campaign = try await campaigns.get(id: id)
            return .success(campaign)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func pauseCampaign(id: String) async -> Result<Campaign> {
        do {
            let campaign = try await campaigns.pause(id: id)
            return .success(campaign)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func resumeCampaign(id: String) async -> Result<Campaign> {
        do {
            let campaign = try await campaigns.resume(id: id)
            return .success(campaign)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Bidding Operations
    
    public func submitBid(campaignId: String, adSlotId: String, amount: Double, 
                         currency: String = "USD") async -> Result<Bid> {
        do {
            let bid = try await bidding.submitBid(campaignId: campaignId, adSlotId: adSlotId, 
                                                  amount: amount, currency: currency)
            return .success(bid)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getBidRecommendations() async -> Result<[[String: AnyCodable]]> {
        do {
            let recommendations = try await bidding.recommendations()
            return .success(recommendations)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getBidStatistics() async -> Result<[String: AnyCodable]> {
        do {
            let stats = try await bidding.stats()
            return .success(stats)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Ad Operations
    
    public func createAd(campaignId: String, placement: String, imageUrl: String, 
                        clickUrl: String, dimensions: String) async -> Result<Ad> {
        do {
            let ad = try await ads.create(campaignId: campaignId, placement: placement, 
                                         imageUrl: imageUrl, clickUrl: clickUrl, dimensions: dimensions)
            return .success(ad)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getAds(campaignId: String, limit: Int = 50) async -> Result<[Ad]> {
        do {
            let ads = try await ads.list(campaignId: campaignId, limit: limit)
            return .success(ads)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Webhook Operations
    
    public func subscribeWebhook(url: String, events: [String]) async -> Result<Webhook> {
        do {
            let webhook = try await webhooks.subscribe(url: url, events: events)
            return .success(webhook)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func getWebhooks(limit: Int = 50) async -> Result<[Webhook]> {
        do {
            let webhooks = try await webhooks.list(limit: limit)
            return .success(webhooks)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func testWebhook(id: String) async -> Result<[String: AnyCodable]> {
        do {
            let result = try await webhooks.test(id: id)
            return .success(result)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    public func onWebhookEvent(_ type: String, handler: @escaping (WebhookEvent) -> Void) {
        webhooks.onEvent(type, handler: handler)
    }
    
    // MARK: - Batch Operations
    
    public func getStatistics() async -> Result<[String: AnyCodable]> {
        do {
            let campaigns = try await campaigns.list(limit: 1000)
            let analytics = try await analytics.realtime()
            let bids = try await bidding.list(limit: 1000)
            
            var stats: [String: AnyCodable] = [:]
            stats["campaignCount"] = .int(campaigns.count)
            stats["analytics"] = .object([:])
            stats["bidCount"] = .int(bids.count)
            
            return .success(stats)
        } catch let error as TaskirXError {
            return .failure(error)
        } catch {
            return .failure(.networkError(error.localizedDescription))
        }
    }
    
    // MARK: - Debug
    
    public func enableDebug(_ enabled: Bool) {
        if enabled {
            print("[TaskirX] 🐛 Debug mode enabled")
        }
    }
}
