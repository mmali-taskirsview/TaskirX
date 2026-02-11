import Foundation

/// SDK errors
public enum AdxError: LocalizedError {
    case notInitialized
    case networkError(Error)
    case noFill
    case invalidRequest(String)
    case internalError(String)
    case timeout
    case adLoadFailed(String)
    case adShowFailed(String)
    case invalidPlacement
    
    public var errorDescription: String? {
        switch self {
        case .notInitialized:
            return "AdxSDK is not initialized. Call AdxSDK.initialize() first."
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        case .noFill:
            return "No ad available for this request."
        case .invalidRequest(let message):
            return "Invalid request: \(message)"
        case .internalError(let message):
            return "Internal error: \(message)"
        case .timeout:
            return "Request timed out."
        case .adLoadFailed(let message):
            return "Failed to load ad: \(message)"
        case .adShowFailed(let message):
            return "Failed to show ad: \(message)"
        case .invalidPlacement:
            return "Invalid placement ID."
        }
    }
}

/// Reward information for rewarded video ads
public struct AdxReward {
    public let type: String
    public let amount: Int
    
    public init(type: String, amount: Int) {
        self.type = type
        self.amount = amount
    }
}
