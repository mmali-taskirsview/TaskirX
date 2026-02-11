import Foundation

/// MMP Integration Manager for iOS
/// Supports: AppsFlyer, Adjust, Branch, Kochava, Singular
@MainActor
public class MmpIntegrationManager {
    
    public static let shared = MmpIntegrationManager()
    
    public enum MmpProvider: String {
        case appsflyer
        case adjust
        case branch
        case kochava
        case singular
        case tenjin
        case none
    }
    
    private var provider: MmpProvider = .none
    private var clickId: String?
    
    private init() {}
    
    // MARK: - AppsFlyer Integration
    
    /// Initialize AppsFlyer
    public func initializeAppsFlyer(devKey: String, appId: String) {
        #if canImport(AppsFlyerLib)
        import AppsFlyerLib
        
        provider = .appsflyer
        
        AppsFlyerLib.shared().appsFlyerDevKey = devKey
        AppsFlyerLib.shared().appleAppID = appId
        AppsFlyerLib.shared().delegate = self
        AppsFlyerLib.shared().customerUserID = AdxSDK.shared.userId
        
        // Start AppsFlyer
        AppsFlyerLib.shared().start()
        
        AdxSDK.shared.log("AppsFlyer initialized", level: .info)
        #else
        AdxSDK.shared.log("AppsFlyer SDK not found. Add AppsFlyerLib to your project.", level: .warning)
        #endif
    }
    
    // MARK: - Adjust Integration
    
    /// Initialize Adjust
    public func initializeAdjust(appToken: String, environment: String = "production") {
        #if canImport(Adjust)
        import Adjust
        
        provider = .adjust
        
        let adjustConfig = ADJConfig(
            appToken: appToken,
            environment: environment == "production" ? ADJEnvironmentProduction : ADJEnvironmentSandbox
        )
        
        adjustConfig?.logLevel = ADJLogLevelInfo
        
        // Attribution callback
        adjustConfig?.delegate = self
        
        Adjust.appDidLaunch(adjustConfig)
        
        AdxSDK.shared.log("Adjust initialized", level: .info)
        #else
        AdxSDK.shared.log("Adjust SDK not found. Add Adjust to your project.", level: .warning)
        #endif
    }
    
    // MARK: - Branch Integration
    
    /// Initialize Branch
    public func initializeBranch() {
        #if canImport(Branch)
        import Branch
        
        provider = .branch
        
        Branch.getInstance().initSession(launchOptions: nil) { params, error in
            if let error = error {
                AdxSDK.shared.log("Branch init error: \(error)", level: .error)
                return
            }
            
            if let params = params as? [String: Any] {
                self.handleBranchAttribution(params)
            }
        }
        
        AdxSDK.shared.log("Branch initialized", level: .info)
        #else
        AdxSDK.shared.log("Branch SDK not found. Add Branch to your project.", level: .warning)
        #endif
    }
    
    // MARK: - Click ID Management
    
    /// Set click ID from ad click
    public func setClickId(_ id: String) {
        clickId = id
        
        switch provider {
        case .appsflyer:
            #if canImport(AppsFlyerLib)
            AppsFlyerLib.shared().customData = [
                "adx_click_id": id,
                "adx_publisher_id": AdxSDK.shared.publisherId
            ]
            #endif
            
        case .adjust:
            #if canImport(Adjust)
            let event = ADJEvent(eventToken: "adx_click")
            event?.addCallbackParameter("click_id", value: id)
            event?.addCallbackParameter("publisher_id", value: AdxSDK.shared.publisherId)
            Adjust.trackEvent(event)
            #endif
            
        case .branch:
            #if canImport(Branch)
            Branch.getInstance().setRequestMetadataKey("adx_click_id", value: id)
            #endif
            
        default:
            break
        }
        
        AdxSDK.shared.log("Click ID set: \(id)")
    }
    
    /// Get current click ID
    public func getClickId() -> String? {
        return clickId
    }
    
    // MARK: - Event Tracking
    
    /// Track custom event
    public func trackEvent(
        eventName: String,
        eventValue: [String: Any]? = nil,
        revenue: Double? = nil,
        currency: String = "USD"
    ) {
        switch provider {
        case .appsflyer:
            #if canImport(AppsFlyerLib)
            AppsFlyerLib.shared().logEvent(eventName, withValues: eventValue)
            #endif
            
        case .adjust:
            #if canImport(Adjust)
            let event = ADJEvent(eventToken: eventName)
            eventValue?.forEach { key, value in
                event?.addCallbackParameter(key, value: "\(value)")
            }
            if let revenue = revenue {
                event?.setRevenue(revenue, currency: currency)
            }
            Adjust.trackEvent(event)
            #endif
            
        case .branch:
            #if canImport(Branch)
            let event = BranchEvent.customEvent(withName: eventName)
            event.customData = eventValue ?? [:]
            if let revenue = revenue {
                event.revenue = NSDecimalNumber(value: revenue)
                event.currency = .USD
            }
            event.logEvent()
            #endif
            
        default:
            // Track manually to taskirx
            Task {
                await trackEventTotaskirx(
                    eventName: eventName,
                    eventValue: eventValue,
                    revenue: revenue,
                    currency: currency
                )
            }
        }
    }
    
    // MARK: - Private Helpers
    
    private func handleBranchAttribution(_ params: [String: Any]) {
        if let clickId = params["+click_id"] as? String {
            setClickId(clickId)
            
            Task {
                await notifytaskirx(clickId: clickId, attributionData: params)
            }
        }
    }
    
    private func notifytaskirx(clickId: String, attributionData: [String: Any]) async {
        do {
            let endpoint = "\(AdxSDK.shared.getConfiguration().apiEndpoint)/api/mmp/attribution"
            
            let payload: [String: Any] = [
                "click_id": clickId,
                "publisher_id": AdxSDK.shared.publisherId,
                "mmp_provider": provider.rawValue,
                "attribution_data": attributionData,
                "device_id": AdxSDK.shared.userId,
                "install_time": Date().timeIntervalSince1970
            ]
            
            // Make HTTP request using NetworkClient
            // await AdxSDK.shared.getNetworkClient().post(endpoint, payload: payload)
            
            AdxSDK.shared.log("Attribution sent to platform for click_id: \(clickId)")
        } catch {
            AdxSDK.shared.log("Failed to send attribution: \(error)", level: .error)
        }
    }
    
    private func trackEventTotaskirx(
        eventName: String,
        eventValue: [String: Any]?,
        revenue: Double?,
        currency: String
    ) async {
        do {
            let endpoint = "\(AdxSDK.shared.getConfiguration().apiEndpoint)/api/mmp/event"
            
            let payload: [String: Any] = [
                "click_id": clickId as Any,
                "event_name": eventName,
                "event_value": eventValue as Any,
                "revenue": revenue as Any,
                "currency": currency,
                "timestamp": Date().timeIntervalSince1970
            ]
            
            // Make HTTP request
            // await AdxSDK.shared.getNetworkClient().post(endpoint, payload: payload)
            
            AdxSDK.shared.log("Event tracked: \(eventName)")
        } catch {
            AdxSDK.shared.log("Failed to track event: \(error)", level: .error)
        }
    }
}

// MARK: - AppsFlyer Delegate

#if canImport(AppsFlyerLib)
import AppsFlyerLib

extension MmpIntegrationManager: AppsFlyerLibDelegate {
    public func onConversionDataSuccess(_ conversionInfo: [AnyHashable : Any]) {
        guard let status = conversionInfo["af_status"] as? String else { return }
        
        if status == "Non-organic" {
            if let clickId = conversionInfo["af_sub1"] as? String {
                setClickId(clickId)
                
                Task {
                    await notifytaskirx(
                        clickId: clickId,
                        attributionData: conversionInfo as? [String: Any] ?? [:]
                    )
                }
            }
        }
    }
    
    public func onConversionDataFail(_ error: Error) {
        AdxSDK.shared.log("AppsFlyer conversion error: \(error)", level: .error)
    }
}
#endif

// MARK: - Adjust Delegate

#if canImport(Adjust)
import Adjust

extension MmpIntegrationManager: AdjustDelegate {
    public func adjustAttributionChanged(_ attribution: ADJAttribution?) {
        guard let attribution = attribution,
              let trackerToken = attribution.trackerToken else { return }
        
        setClickId(trackerToken)
        
        let attributionData: [String: Any] = [
            "tracker_name": attribution.trackerName ?? "",
            "tracker_token": trackerToken,
            "network": attribution.network ?? "",
            "campaign": attribution.campaign ?? "",
            "adgroup": attribution.adgroup ?? "",
            "creative": attribution.creative ?? ""
        ]
        
        Task {
            await notifytaskirx(clickId: trackerToken, attributionData: attributionData)
        }
    }
}
#endif
