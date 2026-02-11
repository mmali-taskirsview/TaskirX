import Foundation

/// Configuration for AdxSDK
public struct AdxConfiguration {
    
    /// Base API endpoint
    public let apiEndpoint: String
    
    /// Enable debug logging
    public var enableDebug: Bool
    
    /// Connection timeout in seconds
    public let connectionTimeout: TimeInterval
    
    /// Read timeout in seconds
    public let readTimeout: TimeInterval
    
    /// Enable test mode (uses test ads)
    public let testMode: Bool
    
    /// Maximum ad request retries
    public let maxRetries: Int
    
    /// Enable location-based targeting
    public let enableLocation: Bool
    
    /// Cache ads for offline viewing
    public let enableAdCaching: Bool
    
    /// Maximum cached ads per placement
    public let maxCachedAds: Int
    
    /// Auto-refresh banner ads (in seconds, 0 to disable)
    public let bannerRefreshInterval: TimeInterval
    
    /// Video ad timeout (in seconds)
    public let videoTimeout: TimeInterval
    
    /// Enable App Tracking Transparency prompt
    public let enableATT: Bool
    
    public init(
        apiEndpoint: String = "https://api.taskirx.com",
        enableDebug: Bool = false,
        connectionTimeout: TimeInterval = 10.0,
        readTimeout: TimeInterval = 10.0,
        testMode: Bool = false,
        maxRetries: Int = 3,
        enableLocation: Bool = false,
        enableAdCaching: Bool = true,
        maxCachedAds: Int = 3,
        bannerRefreshInterval: TimeInterval = 30,
        videoTimeout: TimeInterval = 30,
        enableATT: Bool = true
    ) {
        self.apiEndpoint = apiEndpoint
        self.enableDebug = enableDebug
        self.connectionTimeout = connectionTimeout
        self.readTimeout = readTimeout
        self.testMode = testMode
        self.maxRetries = maxRetries
        self.enableLocation = enableLocation
        self.enableAdCaching = enableAdCaching
        self.maxCachedAds = maxCachedAds
        self.bannerRefreshInterval = bannerRefreshInterval
        self.videoTimeout = videoTimeout
        self.enableATT = enableATT
    }
}

/// Ad size presets
public enum AdSize {
    case banner320x50
    case banner320x100
    case banner300x250
    case banner728x90
    case mediumRectangle300x250
    case fullBanner468x60
    case largeBanner320x100
    case custom(width: CGFloat, height: CGFloat)
    
    public var size: CGSize {
        switch self {
        case .banner320x50:
            return CGSize(width: 320, height: 50)
        case .banner320x100, .largeBanner320x100:
            return CGSize(width: 320, height: 100)
        case .banner300x250, .mediumRectangle300x250:
            return CGSize(width: 300, height: 250)
        case .banner728x90:
            return CGSize(width: 728, height: 90)
        case .fullBanner468x60:
            return CGSize(width: 468, height: 60)
        case .custom(let width, let height):
            return CGSize(width: width, height: height)
        }
    }
}

/// Ad format types
public enum AdFormat: String {
    case banner
    case interstitial
    case native
    case video
    case rewardedVideo = "rewarded-video"
}

/// Gender for targeting
public enum Gender: String {
    case male
    case female
    case other
    case unknown
}

/// Ad targeting parameters
public struct AdxTargeting {
    public let age: Int?
    public let gender: Gender?
    public let interests: [String]?
    public let keywords: [String]?
    public let location: CLLocation?
    public let customTargeting: [String: String]?
    
    public init(
        age: Int? = nil,
        gender: Gender? = nil,
        interests: [String]? = nil,
        keywords: [String]? = nil,
        location: CLLocation? = nil,
        customTargeting: [String: String]? = nil
    ) {
        self.age = age
        self.gender = gender
        self.interests = interests
        self.keywords = keywords
        self.location = location
        self.customTargeting = customTargeting
    }
}

import CoreLocation
