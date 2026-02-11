import Foundation
import Combine

/// Main entry point for AdxSDK
@MainActor
public class AdxSDK {
    
    // MARK: - Properties
    
    public static let shared = AdxSDK()
    public private(set) var isInitialized = false
    
    private(set) var publisherId: String = ""
    private(set) var configuration: AdxConfiguration!
    private(set) var sessionId: String = UUID().uuidString
    
    private let networkClient: NetworkClient
    private let deviceInfo: DeviceInfo
    private let userIdManager: UserIdManager
    
    // MARK: - Constants
    
    public static let version = "1.0.0"
    
    // MARK: - Initialization
    
    private init() {
        self.networkClient = NetworkClient()
        self.deviceInfo = DeviceInfo()
        self.userIdManager = UserIdManager()
    }
    
    /// Initialize the AdxSDK
    /// - Parameters:
    ///   - publisherId: Your publisher ID from taskirx
    ///   - configuration: SDK configuration
    public static func initialize(
        publisherId: String,
        configuration: AdxConfiguration = AdxConfiguration()
    ) {
        Task { @MainActor in
            await shared.configure(publisherId: publisherId, configuration: configuration)
        }
    }
    
    private func configure(publisherId: String, configuration: AdxConfiguration) async {
        guard !publisherId.isEmpty else {
            log("Error: Publisher ID cannot be empty", level: .error)
            return
        }
        
        self.publisherId = publisherId
        self.configuration = configuration
        self.networkClient.configure(with: configuration)
        self.isInitialized = true
        
        log("AdxSDK v\(Self.version) initialized for publisher: \(publisherId)")
        log("API Endpoint: \(configuration.apiEndpoint)")
        log("Session ID: \(sessionId)")
        
        // Request tracking permission if needed
        if configuration.enableATT {
            await requestTrackingPermission()
        }
    }
    
    // MARK: - Public Methods
    
    /// Get the user ID (persistent across sessions)
    public var userId: String {
        userIdManager.getUserId()
    }
    
    /// Set custom user ID
    public func setUserId(_ userId: String) {
        userIdManager.setCustomUserId(userId)
    }
    
    /// Get IDFA (Identifier for Advertisers) if available
    public var advertisingIdentifier: String? {
        deviceInfo.advertisingIdentifier
    }
    
    /// Start new session
    public func startNewSession() {
        sessionId = UUID().uuidString
        log("New session started: \(sessionId)")
    }
    
    /// Enable or disable debug logging
    public func setDebugEnabled(_ enabled: Bool) {
        configuration.enableDebug = enabled
    }
    
    // MARK: - Internal Access
    
    internal func getNetworkClient() -> NetworkClient {
        return networkClient
    }
    
    internal func getDeviceInfo() -> DeviceInfo {
        return deviceInfo
    }
    
    internal func getConfiguration() -> AdxConfiguration {
        return configuration
    }
    
    // MARK: - Logging
    
    internal func log(_ message: String, level: LogLevel = .debug) {
        guard configuration.enableDebug else { return }
        
        let prefix: String
        switch level {
        case .debug:
            prefix = "🔵 [AdxSDK]"
        case .info:
            prefix = "ℹ️ [AdxSDK]"
        case .warning:
            prefix = "⚠️ [AdxSDK]"
        case .error:
            prefix = "❌ [AdxSDK]"
        }
        
        print("\(prefix) \(message)")
    }
    
    // MARK: - Private Methods
    
    private func requestTrackingPermission() async {
        #if canImport(AppTrackingTransparency)
        import AppTrackingTransparency
        
        let status = await ATTrackingManager.requestTrackingAuthorization()
        
        switch status {
        case .authorized:
            log("Tracking authorized")
        case .denied:
            log("Tracking denied", level: .warning)
        case .notDetermined:
            log("Tracking not determined", level: .warning)
        case .restricted:
            log("Tracking restricted", level: .warning)
        @unknown default:
            break
        }
        #endif
    }
}

// MARK: - Log Level

internal enum LogLevel {
    case debug
    case info
    case warning
    case error
}

// MARK: - Ensure Initialized

internal func ensureInitialized() throws {
    guard AdxSDK.shared.isInitialized else {
        throw AdxError.notInitialized
    }
}
