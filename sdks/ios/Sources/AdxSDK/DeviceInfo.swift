import UIKit
import AdSupport
import AppTrackingTransparency

internal class DeviceInfo {
    
    // MARK: - Device Information
    
    var deviceType: Int {
        switch UIDevice.current.userInterfaceIdiom {
        case .phone:
            return 1 // Mobile
        case .pad:
            return 5 // Tablet
        case .tv:
            return 3 // TV
        default:
            return 0 // Unknown
        }
    }
    
    var manufacturer: String {
        return "Apple"
    }
    
    var model: String {
        return UIDevice.current.model
    }
    
    var osVersion: String {
        return UIDevice.current.systemVersion
    }
    
    var screenWidth: Int {
        return Int(UIScreen.main.bounds.width * UIScreen.main.scale)
    }
    
    var screenHeight: Int {
        return Int(UIScreen.main.bounds.height * UIScreen.main.scale)
    }
    
    var language: String {
        return Locale.current.language.languageCode?.identifier ?? "en"
    }
    
    var userAgent: String {
        let appName = Bundle.main.object(forInfoDictionaryKey: "CFBundleName") as? String ?? "App"
        let appVersion = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0"
        let osVersion = UIDevice.current.systemVersion
        let model = UIDevice.current.model
        
        return "Mozilla/5.0 (\(model); iOS \(osVersion)) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/\(appName)/\(appVersion) AdxSDK/\(AdxSDK.version)"
    }
    
    // MARK: - App Information
    
    var appName: String {
        return Bundle.main.object(forInfoDictionaryKey: "CFBundleName") as? String ?? "Unknown"
    }
    
    var bundleIdentifier: String {
        return Bundle.main.bundleIdentifier ?? "com.unknown.app"
    }
    
    var appVersion: String {
        return Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0.0"
    }
    
    var appStoreURL: String? {
        guard let appID = Bundle.main.object(forInfoDictionaryKey: "AppStoreID") as? String else {
            return nil
        }
        return "https://apps.apple.com/app/id\(appID)"
    }
    
    // MARK: - Advertising Identifier
    
    var advertisingIdentifier: String? {
        guard ASIdentifierManager.shared().isAdvertisingTrackingEnabled else {
            return nil
        }
        
        let idfa = ASIdentifierManager.shared().advertisingIdentifier
        return idfa.uuidString
    }
    
    var isLimitAdTrackingEnabled: Bool {
        return !ASIdentifierManager.shared().isAdvertisingTrackingEnabled
    }
    
    @available(iOS 14, *)
    var trackingAuthorizationStatus: ATTrackingManager.AuthorizationStatus {
        return ATTrackingManager.trackingAuthorizationStatus
    }
    
    // MARK: - Network Information
    
    var connectionType: Int {
        // 0 = unknown, 2 = wifi, 3 = cellular
        // Simplified - could use Network framework for more detail
        return 0
    }
    
    // MARK: - Vendor Identifier (fallback)
    
    var vendorIdentifier: String {
        return UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString
    }
}
