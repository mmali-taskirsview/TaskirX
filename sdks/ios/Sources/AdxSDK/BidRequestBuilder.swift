import Foundation
import CoreLocation

internal class BidRequestBuilder {
    
    private let deviceInfo: DeviceInfo
    
    init(deviceInfo: DeviceInfo) {
        self.deviceInfo = deviceInfo
    }
    
    func buildRequest(
        placementId: String,
        adFormat: AdFormat,
        adSize: AdSize?,
        targeting: AdxTargeting? = nil
    ) -> BidRequest {
        let impressionId = UUID().uuidString
        
        // Create impression based on ad format
        let impression: Impression
        switch adFormat {
        case .banner:
            guard let size = adSize else {
                fatalError("Banner ads require adSize parameter")
            }
            impression = Impression(
                id: impressionId,
                banner: Banner(
                    w: Int(size.size.width),
                    h: Int(size.size.height),
                    format: nil
                ),
                video: nil,
                native: nil,
                bidfloor: 0.0,
                tagid: placementId
            )
            
        case .video, .rewardedVideo:
            impression = Impression(
                id: impressionId,
                banner: nil,
                video: Video(
                    mimes: ["video/mp4", "video/quicktime"],
                    minduration: 5,
                    maxduration: 30,
                    protocols: [2, 3, 5, 6],
                    w: deviceInfo.screenWidth,
                    h: deviceInfo.screenHeight
                ),
                native: nil,
                bidfloor: 0.0,
                tagid: placementId
            )
            
        case .native:
            impression = Impression(
                id: impressionId,
                banner: nil,
                video: nil,
                native: Native(request: buildNativeRequest()),
                bidfloor: 0.0,
                tagid: placementId
            )
            
        case .interstitial:
            impression = Impression(
                id: impressionId,
                banner: Banner(
                    w: deviceInfo.screenWidth,
                    h: deviceInfo.screenHeight,
                    format: nil
                ),
                video: nil,
                native: nil,
                bidfloor: 0.0,
                tagid: placementId
            )
        }
        
        // Create device
        let device = Device(
            ua: deviceInfo.userAgent,
            ip: "",
            devicetype: deviceInfo.deviceType,
            make: deviceInfo.manufacturer,
            model: deviceInfo.model,
            os: "iOS",
            osv: deviceInfo.osVersion,
            w: deviceInfo.screenWidth,
            h: deviceInfo.screenHeight,
            ifa: deviceInfo.advertisingIdentifier,
            lmt: deviceInfo.isLimitAdTrackingEnabled ? 1 : 0,
            connectiontype: deviceInfo.connectionType,
            language: deviceInfo.language
        )
        
        // Create user
        let user = User(
            id: AdxSDK.shared.userId,
            yob: nil,
            gender: targeting?.gender?.rawValue,
            keywords: targeting?.keywords?.joined(separator: ","),
            geo: targeting?.location.map { location in
                Geo(
                    lat: location.coordinate.latitude,
                    lon: location.coordinate.longitude,
                    type: 1,
                    accuracy: Int(location.horizontalAccuracy),
                    country: nil,
                    city: nil
                )
            }
        )
        
        // Create app
        let app = App(
            id: AdxSDK.shared.publisherId,
            name: deviceInfo.appName,
            bundle: deviceInfo.bundleIdentifier,
            storeurl: deviceInfo.appStoreURL,
            publisher: Publisher(
                id: AdxSDK.shared.publisherId,
                name: nil
            )
        )
        
        // Create bid request
        return BidRequest(
            id: UUID().uuidString,
            imp: [impression],
            device: device,
            user: user,
            app: app,
            test: AdxSDK.shared.getConfiguration().testMode ? 1 : 0
        )
    }
    
    private func buildNativeRequest() -> String {
        return """
        {
            "ver": "1.2",
            "assets": [
                {"id": 1, "required": 1, "title": {"len": 90}},
                {"id": 2, "required": 0, "img": {"type": 3, "wmin": 300, "hmin": 250}},
                {"id": 3, "required": 0, "data": {"type": 2, "len": 150}}
            ]
        }
        """
    }
}
