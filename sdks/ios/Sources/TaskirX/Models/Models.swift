import Foundation

// MARK: - Configuration

public struct ClientConfig: Codable {
    public let apiUrl: String
    public let apiKey: String
    public let debug: Bool
    public let timeout: TimeInterval
    public let retryAttempts: Int
    
    public init(apiUrl: String, apiKey: String, debug: Bool = false, 
                timeout: TimeInterval = 30, retryAttempts: Int = 3) {
        self.apiUrl = apiUrl
        self.apiKey = apiKey
        self.debug = debug
        self.timeout = timeout
        self.retryAttempts = retryAttempts
    }
}

// MARK: - Authentication

public struct AuthResponse: Codable {
    public let token: String
    public let refreshToken: String
    public let user: User
    public let expiresIn: Int
}

public struct LoginRequest: Codable {
    public let email: String
    public let password: String
    
    public init(email: String, password: String) {
        self.email = email
        self.password = password
    }
}

public struct RegisterRequest: Codable {
    public let email: String
    public let password: String
    public let name: String
    public let company: String?
    
    public init(email: String, password: String, name: String, company: String? = nil) {
        self.email = email
        self.password = password
        self.name = name
        self.company = company
    }
}

// MARK: - User

public struct User: Codable {
    public let id: String
    public let email: String
    public let name: String
    public let company: String?
    public let role: String
    public let createdAt: String
    public let updatedAt: String
}

// MARK: - Campaign

public struct Campaign: Codable {
    public let id: String
    public let name: String
    public let budget: Double
    public let startDate: String
    public let endDate: String
    public let status: String
    public let targetAudience: [String: AnyCodable]
    public let createdAt: String
    public let updatedAt: String
}

public struct CampaignCreateRequest: Codable {
    public let name: String
    public let budget: Double
    public let startDate: String
    public let endDate: String
    public let targetAudience: [String: AnyCodable]
    
    public init(name: String, budget: Double, startDate: String, 
                endDate: String, targetAudience: [String: AnyCodable]) {
        self.name = name
        self.budget = budget
        self.startDate = startDate
        self.endDate = endDate
        self.targetAudience = targetAudience
    }
}

// MARK: - Bid

public struct Bid: Codable {
    public let id: String
    public let campaignId: String
    public let adSlotId: String
    public let amount: Double
    public let currency: String
    public let status: String
    public let createdAt: String
    public let updatedAt: String
}

public struct BidSubmitRequest: Codable {
    public let campaignId: String
    public let adSlotId: String
    public let amount: Double
    public let currency: String
    
    public init(campaignId: String, adSlotId: String, amount: Double, currency: String = "USD") {
        self.campaignId = campaignId
        self.adSlotId = adSlotId
        self.amount = amount
        self.currency = currency
    }
}

// MARK: - Analytics

public struct Analytics: Codable {
    public let impressions: Int
    public let clicks: Int
    public let conversions: Int
    public let spend: Double
    public let revenue: Double
    public let ctr: Double
    public let conversionRate: Double
    public let roi: Double
    public let timestamp: String
}

// MARK: - Ad

public struct Ad: Codable {
    public let id: String
    public let campaignId: String
    public let placement: String
    public let imageUrl: String
    public let clickUrl: String
    public let dimensions: String
    public let status: String
    public let createdAt: String
    public let updatedAt: String
}

public struct AdCreateRequest: Codable {
    public let campaignId: String
    public let placement: String
    public let imageUrl: String
    public let clickUrl: String
    public let dimensions: String
    
    public init(campaignId: String, placement: String, imageUrl: String, 
                clickUrl: String, dimensions: String) {
        self.campaignId = campaignId
        self.placement = placement
        self.imageUrl = imageUrl
        self.clickUrl = clickUrl
        self.dimensions = dimensions
    }
}

// MARK: - Webhook

public struct Webhook: Codable {
    public let id: String
    public let url: String
    public let events: [String]
    public let active: Bool
    public let createdAt: String
    public let updatedAt: String
}

public struct WebhookCreateRequest: Codable {
    public let url: String
    public let events: [String]
    
    public init(url: String, events: [String]) {
        self.url = url
        self.events = events
    }
}

public struct WebhookEvent: Codable {
    public let id: String
    public let type: String
    public let data: [String: AnyCodable]
    public let timestamp: String
}

// MARK: - Error Response

public struct ErrorResponse: Codable {
    public let code: String
    public let message: String
    public let details: [String: AnyCodable]?
}

// MARK: - API Response

public struct ApiResponse<T: Codable>: Codable {
    public let success: Bool
    public let data: T?
    public let error: ErrorResponse?
}

// MARK: - Helper: AnyCodable

public enum AnyCodable: Codable {
    case null
    case bool(Bool)
    case int(Int)
    case double(Double)
    case string(String)
    case array([AnyCodable])
    case object([String: AnyCodable])
    
    public init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if container.decodeNil() {
            self = .null
        } else if let bool = try? container.decode(Bool.self) {
            self = .bool(bool)
        } else if let int = try? container.decode(Int.self) {
            self = .int(int)
        } else if let double = try? container.decode(Double.self) {
            self = .double(double)
        } else if let string = try? container.decode(String.self) {
            self = .string(string)
        } else if let array = try? container.decode([AnyCodable].self) {
            self = .array(array)
        } else if let object = try? container.decode([String: AnyCodable].self) {
            self = .object(object)
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode AnyCodable")
        }
    }
    
    public func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        switch self {
        case .null:
            try container.encodeNil()
        case .bool(let bool):
            try container.encode(bool)
        case .int(let int):
            try container.encode(int)
        case .double(let double):
            try container.encode(double)
        case .string(let string):
            try container.encode(string)
        case .array(let array):
            try container.encode(array)
        case .object(let object):
            try container.encode(object)
        }
    }
}
